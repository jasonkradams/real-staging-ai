package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/virtual-staging-ai/api/internal/stripe"
	"github.com/virtual-staging-ai/api/internal/user"
)

// StripeEvent represents a Stripe webhook event.
type StripeEvent struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
	Created int64                  `json:"created"`
}

// CheckoutSession represents a Stripe checkout session.
type CheckoutSession struct {
	ID                string `json:"id"`
	CustomerID        string `json:"customer"`
	PaymentStatus     string `json:"payment_status"`
	SubscriptionID    string `json:"subscription"`
	ClientReferenceID string `json:"client_reference_id"`
}

// stripeWebhookHandler handles POST /api/v1/stripe/webhook requests.
func (s *Server) stripeWebhookHandler(c echo.Context) error {
	// Read the request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Unable to read request body")
	}

	// Check for empty body
	if len(body) == 0 {
		log.Printf("Empty webhook body received")
		return echo.NewHTTPError(http.StatusBadRequest, "Empty request body")
	}

	// In production, you should verify the webhook signature
	// using the Stripe-Signature header and your webhook secret
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	//nolint:staticcheck
	if webhookSecret != "" {
		stripeSignature := c.Request().Header.Get("Stripe-Signature")
		if err := verifyStripeSignature(body, stripeSignature, webhookSecret, 5*time.Minute, time.Now); err != nil {
			log.Printf("Stripe signature verification failed: %v", err)
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid webhook signature",
			})
		}
	} else {
		// STRIPE_WEBHOOK_SECRET is not set; skipping signature verification.
		// This is acceptable in local/dev environments, but MUST be configured in production.
		// See docs/configuration.md for environment variable setup.
		log.Printf("Stripe webhook: STRIPE_WEBHOOK_SECRET not set; skipping signature verification")
	}

	// Parse the webhook event
	var event StripeEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing webhook JSON: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format")
	}

	log.Printf("Received Stripe webhook event: %s (ID: %s)", event.Type, event.ID)

	// Idempotency check: ensure the event hasn't already been processed
	processed, err := s.alreadyProcessedStripeEvent(c.Request().Context(), event.ID)
	if err != nil {
		log.Printf("Error checking Stripe event idempotency: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
	}
	if processed {
		// Already processed; acknowledge to prevent retries
		return c.JSON(http.StatusOK, map[string]string{
			"status": "duplicate",
		})
	}

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		if err := s.handleCheckoutSessionCompleted(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling checkout.session.completed: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
		}

	case "customer.subscription.created":
		if err := s.handleSubscriptionCreated(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.subscription.created: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
		}

	case "customer.subscription.updated":
		if err := s.handleSubscriptionUpdated(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.subscription.updated: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
		}

	case "customer.subscription.deleted":
		if err := s.handleSubscriptionDeleted(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.subscription.deleted: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
		}

	case "customer.created":
		if err := s.handleCustomerCreated(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.created: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
		}

	case "customer.updated":
		if err := s.handleCustomerUpdated(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.updated: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
		}

	case "customer.deleted":
		if err := s.handleCustomerDeleted(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.deleted: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
		}

	case "invoice.payment_succeeded":
		if err := s.handleInvoicePaymentSucceeded(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling invoice.payment_succeeded: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
		}

	case "invoice.payment_failed":
		if err := s.handleInvoicePaymentFailed(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling invoice.payment_failed: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
		}

	default:
		log.Printf("Unhandled webhook event type: %s", event.Type)
	}

	// Mark event as processed (idempotency scaffold). If this fails, log and still acknowledge.
	if err := s.markStripeEventProcessed(c.Request().Context(), event.ID); err != nil {
		log.Printf("Failed to mark Stripe event processed: %v", err)
	}

	// Return 200 to acknowledge receipt of the webhook
	return c.JSON(http.StatusOK, map[string]string{
		"status": "received",
	})
}

// handleCheckoutSessionCompleted processes successful checkout sessions.
func (s *Server) handleCheckoutSessionCompleted(ctx context.Context, event *StripeEvent) error {
	// Extract checkout session data
	sessionData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid checkout session data")
	}

	customerID, _ := sessionData["customer"].(string)
	paymentStatus, _ := sessionData["payment_status"].(string)
	clientReferenceID, _ := sessionData["client_reference_id"].(string)

	log.Printf("Checkout completed - Customer: %s, Payment Status: %s, Reference: %s",
		customerID, paymentStatus, clientReferenceID)

	// Link Stripe customer to a user by client_reference_id (Auth0 sub or internal user ref)
	if clientReferenceID != "" && customerID != "" {
		userRepo := user.NewUserRepository(s.db)
		u, err := userRepo.GetByAuth0Sub(ctx, clientReferenceID)
		if err == nil {
			// Only set stripe_customer_id if it's not already set
			if !u.StripeCustomerID.Valid || u.StripeCustomerID.String == "" {
				if _, err := userRepo.UpdateStripeCustomerID(ctx, u.ID.String(), customerID); err != nil {
					log.Printf("Failed to update user's Stripe customer ID: %v", err)
				}
			}
		} else {
			// Not fatal for webhook handling; just log
			log.Printf("Could not find user by client_reference_id=%s to link Stripe customer=%s: %v", clientReferenceID, customerID, err)
		}
	}

	return nil
}

// handleSubscriptionCreated processes new subscription events.
func (s *Server) handleSubscriptionCreated(ctx context.Context, event *StripeEvent) error {
	subscriptionData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data")
	}

	customerID, _ := subscriptionData["customer"].(string)
	subscriptionID, _ := subscriptionData["id"].(string)
	status, _ := subscriptionData["status"].(string)

	log.Printf("Subscription created - Customer: %s, Subscription: %s, Status: %s",
		customerID, subscriptionID, status)

	// Persist subscription state
	subRepo := stripe.NewSubscriptionsRepository(s.db)
	userRepo := user.NewUserRepository(s.db)
	if customerID != "" && subscriptionID != "" {
		if u, err := userRepo.GetByStripeCustomerID(ctx, customerID); err == nil {
			if _, err := subRepo.UpsertByStripeID(ctx, u.ID.String(), subscriptionID, status, nil, nil, nil, nil, nil, false); err != nil {
				log.Printf("Failed to upsert subscription (created): %v", err)
			}
		} else {
			log.Printf("No user found for Stripe customer on subscription.created: %s (err=%v)", customerID, err)
		}
	}
	return nil
}

// handleSubscriptionUpdated processes subscription update events.
func (s *Server) handleSubscriptionUpdated(ctx context.Context, event *StripeEvent) error {
	subscriptionData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data")
	}

	customerID, _ := subscriptionData["customer"].(string)
	subscriptionID, _ := subscriptionData["id"].(string)
	status, _ := subscriptionData["status"].(string)

	log.Printf("Subscription updated - Customer: %s, Subscription: %s, Status: %s",
		customerID, subscriptionID, status)

	// Persist subscription state
	subRepo := stripe.NewSubscriptionsRepository(s.db)
	userRepo := user.NewUserRepository(s.db)
	if customerID != "" && subscriptionID != "" {
		if u, err := userRepo.GetByStripeCustomerID(ctx, customerID); err == nil {
			if _, err := subRepo.UpsertByStripeID(ctx, u.ID.String(), subscriptionID, status, nil, nil, nil, nil, nil, false); err != nil {
				log.Printf("Failed to upsert subscription (updated): %v", err)
			}
		} else {
			log.Printf("No user found for Stripe customer on subscription.updated: %s (err=%v)", customerID, err)
		}
	}
	return nil
}

// handleSubscriptionDeleted processes subscription cancellation events.
func (s *Server) handleSubscriptionDeleted(ctx context.Context, event *StripeEvent) error {
	subscriptionData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data")
	}

	customerID, _ := subscriptionData["customer"].(string)
	subscriptionID, _ := subscriptionData["id"].(string)

	log.Printf("Subscription deleted - Customer: %s, Subscription: %s",
		customerID, subscriptionID)

	// Mark subscription as canceled/deactivated
	subRepo := stripe.NewSubscriptionsRepository(s.db)
	userRepo := user.NewUserRepository(s.db)
	if customerID != "" && subscriptionID != "" {
		// Stripe sends a final status (typically "canceled"); persist it
		if u, err := userRepo.GetByStripeCustomerID(ctx, customerID); err == nil {
			if _, err := subRepo.UpsertByStripeID(ctx, u.ID.String(), subscriptionID, "canceled", nil, nil, nil, nil, nil, false); err != nil {
				log.Printf("Failed to upsert subscription (deleted): %v", err)
			}
		} else {
			log.Printf("No user found for Stripe customer on subscription.deleted: %s (err=%v)", customerID, err)
		}
	}
	return nil
}

// handleInvoicePaymentSucceeded processes successful payment events.
func (s *Server) handleInvoicePaymentSucceeded(ctx context.Context, event *StripeEvent) error {
	invoiceData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid invoice data")
	}

	customerID, _ := invoiceData["customer"].(string)
	subscriptionID, _ := invoiceData["subscription"].(string)
	amountPaid, _ := invoiceData["amount_paid"].(float64)

	log.Printf("Invoice payment succeeded - Customer: %s, Subscription: %s, Amount: %.2f",
		customerID, subscriptionID, amountPaid/100)

	// TODO: Record successful payment in the database
	return nil
}

// handleInvoicePaymentFailed processes failed payment events.
func (s *Server) handleInvoicePaymentFailed(ctx context.Context, event *StripeEvent) error {
	invoiceData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid invoice data")
	}

	customerID, _ := invoiceData["customer"].(string)
	subscriptionID, _ := invoiceData["subscription"].(string)

	log.Printf("Invoice payment failed - Customer: %s, Subscription: %s",
		customerID, subscriptionID)

	// TODO: Handle failed payment (send notification, update status, etc.)
	return nil
}

// handleCustomerCreated processes new customer events.
func (s *Server) handleCustomerCreated(ctx context.Context, event *StripeEvent) error {
	customerData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid customer data")
	}

	customerID, _ := customerData["id"].(string)
	email, _ := customerData["email"].(string)

	log.Printf("Customer created: ID=%s, Email=%s", customerID, email)
	// TODO: Implement customer creation logic
	// - Store customer information in database
	// - Set up user account if needed

	return nil
}

// handleCustomerUpdated processes customer update events.
func (s *Server) handleCustomerUpdated(ctx context.Context, event *StripeEvent) error {
	customerData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid customer data")
	}

	customerID, _ := customerData["id"].(string)
	email, _ := customerData["email"].(string)

	log.Printf("Customer updated: ID=%s, Email=%s", customerID, email)
	// TODO: Implement customer update logic
	// - Update customer information in database

	return nil
}

// handleCustomerDeleted processes customer deletion events.
func (s *Server) handleCustomerDeleted(ctx context.Context, event *StripeEvent) error {
	customerData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid customer data")
	}

	customerID, _ := customerData["id"].(string)

	log.Printf("Customer deleted: ID=%s", customerID)
	// TODO: Implement customer deletion logic
	// - Handle customer data cleanup
	// - Potentially disable user account

	return nil
}

// verifyStripeSignature verifies the Stripe webhook signature.
// See: https://docs.stripe.com/webhooks/signatures
func verifyStripeSignature(body []byte, sigHeader, secret string, tolerance time.Duration, now func() time.Time) error {
	if sigHeader == "" {
		return fmt.Errorf("missing Stripe-Signature header")
	}

	ts, v1s, err := parseStripeSignatureHeader(sigHeader)
	if err != nil {
		return fmt.Errorf("invalid signature header: %w", err)
	}
	if len(v1s) == 0 {
		return fmt.Errorf("no v1 signatures found")
	}

	// Enforce timestamp tolerance window
	t := time.Unix(ts, 0)
	diff := now().Sub(t)
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		return fmt.Errorf("timestamp outside tolerance window")
	}

	expected := computeStripeSignature(body, ts, secret)

	// Compare expected signature with any provided v1 (constant-time)
	for _, sig := range v1s {
		if hmac.Equal(sig, expected) {
			return nil
		}
	}

	return fmt.Errorf("no matching signature")
}

// parseStripeSignatureHeader parses "t=timestamp,v1=hex,v1=hex2,..."
func parseStripeSignatureHeader(header string) (int64, [][]byte, error) {
	var ts int64
	var haveTS bool
	var v1s [][]byte

	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "t=") {
			tStr := strings.TrimPrefix(part, "t=")
			val, err := strconv.ParseInt(tStr, 10, 64)
			if err != nil {
				return 0, nil, fmt.Errorf("invalid timestamp: %w", err)
			}
			ts = val
			haveTS = true
		} else if strings.HasPrefix(part, "v1=") {
			v := strings.TrimPrefix(part, "v1=")
			b, err := hex.DecodeString(v)
			if err != nil {
				return 0, nil, fmt.Errorf("invalid v1 signature: %w", err)
			}
			v1s = append(v1s, b)
		}
	}

	if !haveTS {
		return 0, nil, fmt.Errorf("missing timestamp")
	}

	return ts, v1s, nil
}

// computeStripeSignature computes HMAC-SHA256 over "timestamp.payload"
func computeStripeSignature(body []byte, ts int64, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = fmt.Fprintf(mac, "%d.", ts)
	_, _ = mac.Write(body)
	return mac.Sum(nil)
}

// alreadyProcessedStripeEvent checks if the event was already processed (idempotency scaffold).
// TODO: Back with a DB table (e.g., processed_events with unique(stripe_event_id)).
func (s *Server) alreadyProcessedStripeEvent(ctx context.Context, eventID string) (bool, error) {
	repo := stripe.NewProcessedEventsRepository(s.db)
	return repo.IsProcessed(ctx, eventID)
}

// markStripeEventProcessed records the processed event (idempotency scaffold).
// TODO: Insert into DB and enforce uniqueness on stripe_event_id.
func (s *Server) markStripeEventProcessed(ctx context.Context, eventID string) error {
	repo := stripe.NewProcessedEventsRepository(s.db)
	_, err := repo.Upsert(ctx, eventID, nil, nil)
	return err
}
