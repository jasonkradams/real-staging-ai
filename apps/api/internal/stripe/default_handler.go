package stripe

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

	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/user"
)

// DefaultHandler handles Stripe webhooks and related event processing.
type DefaultHandler struct {
	db storage.Database
}

// NewDefaultHandler constructs a Stripe DefaultHandler.
func NewDefaultHandler(db storage.Database) *DefaultHandler {
	return &DefaultHandler{db: db}
}

// errorResponse is a simple JSON error envelope for handler responses.
type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// StripeEvent represents a Stripe webhook event (subset).
type StripeEvent struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
	Created int64                  `json:"created"`
}

// CheckoutSession represents a Stripe checkout session (subset).
type CheckoutSession struct {
	ID                string `json:"id"`
	CustomerID        string `json:"customer"`
	PaymentStatus     string `json:"payment_status"`
	SubscriptionID    string `json:"subscription"`
	ClientReferenceID string `json:"client_reference_id"`
}

// Webhook handles POST /api/v1/stripe/webhook requests.
func (h *DefaultHandler) Webhook(c echo.Context) error {
	// Read the request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error:   "bad_request",
			Message: "Unable to read request body",
		})
	}

	// Check for empty body
	if len(body) == 0 {
		log.Printf("Empty webhook body received")
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error:   "bad_request",
			Message: "Empty request body",
		})
	}

	// In production, verify the webhook signature using STRIPE_WEBHOOK_SECRET
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	//nolint:staticcheck
	if webhookSecret != "" {
		stripeSignature := c.Request().Header.Get("Stripe-Signature")
		if err := verifyStripeSignature(body, stripeSignature, webhookSecret, 5*time.Minute, time.Now); err != nil {
			log.Printf("Stripe signature verification failed: %v", err)
			return c.JSON(http.StatusUnauthorized, errorResponse{
				Error:   "unauthorized",
				Message: "Invalid webhook signature",
			})
		}
	} else {
		// STRIPE_WEBHOOK_SECRET is not set; skipping signature verification.
		// This is acceptable in local/dev environments, but MUST be configured in production.
		log.Printf("Stripe webhook: STRIPE_WEBHOOK_SECRET not set; skipping signature verification")
	}

	// Parse the webhook event
	var event StripeEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing webhook JSON: %v", err)
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error:   "bad_request",
			Message: "Invalid JSON format",
		})
	}

	log.Printf("Received Stripe webhook event: %s (ID: %s)", event.Type, event.ID)

	// Idempotency check: ensure the event hasn't already been processed
	processed, err := h.alreadyProcessedStripeEvent(c.Request().Context(), event.ID)
	if err != nil {
		log.Printf("Error checking Stripe event idempotency: %v", err)
		return c.JSON(http.StatusInternalServerError, errorResponse{
			Error:   "internal_server_error",
			Message: "Failed to process webhook",
		})
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
		if err := h.handleCheckoutSessionCompleted(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling checkout.session.completed: %v", err)
			return c.JSON(http.StatusInternalServerError, errorResponse{
				Error:   "internal_server_error",
				Message: "Failed to process webhook",
			})
		}

	case "customer.subscription.created":
		if err := h.handleSubscriptionCreated(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.subscription.created: %v", err)
			return c.JSON(http.StatusInternalServerError, errorResponse{
				Error:   "internal_server_error",
				Message: "Failed to process webhook",
			})
		}

	case "customer.subscription.updated":
		if err := h.handleSubscriptionUpdated(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.subscription.updated: %v", err)
			return c.JSON(http.StatusInternalServerError, errorResponse{
				Error:   "internal_server_error",
				Message: "Failed to process webhook",
			})
		}

	case "customer.subscription.deleted":
		if err := h.handleSubscriptionDeleted(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.subscription.deleted: %v", err)
			return c.JSON(http.StatusInternalServerError, errorResponse{
				Error:   "internal_server_error",
				Message: "Failed to process webhook",
			})
		}

	case "customer.created":
		if err := h.handleCustomerCreated(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.created: %v", err)
			return c.JSON(http.StatusInternalServerError, errorResponse{
				Error:   "internal_server_error",
				Message: "Failed to process webhook",
			})
		}

	case "customer.updated":
		if err := h.handleCustomerUpdated(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.updated: %v", err)
			return c.JSON(http.StatusInternalServerError, errorResponse{
				Error:   "internal_server_error",
				Message: "Failed to process webhook",
			})
		}

	case "customer.deleted":
		if err := h.handleCustomerDeleted(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling customer.deleted: %v", err)
			return c.JSON(http.StatusInternalServerError, errorResponse{
				Error:   "internal_server_error",
				Message: "Failed to process webhook",
			})
		}

	case "invoice.payment_succeeded":
		if err := h.handleInvoicePaymentSucceeded(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling invoice.payment_succeeded: %v", err)
			return c.JSON(http.StatusInternalServerError, errorResponse{
				Error:   "internal_server_error",
				Message: "Failed to process webhook",
			})
		}

	case "invoice.payment_failed":
		if err := h.handleInvoicePaymentFailed(c.Request().Context(), &event); err != nil {
			log.Printf("Error handling invoice.payment_failed: %v", err)
			return c.JSON(http.StatusInternalServerError, errorResponse{
				Error:   "internal_server_error",
				Message: "Failed to process webhook",
			})
		}

	default:
		log.Printf("Unhandled webhook event type: %s", event.Type)
	}

	// Mark event as processed (idempotency scaffold). If this fails, log and still acknowledge.
	if err := h.markStripeEventProcessed(c.Request().Context(), event.ID, event.Type, body); err != nil {
		log.Printf("Failed to mark Stripe event processed: %v", err)
	}

	// Return 200 to acknowledge receipt of the webhook
	return c.JSON(http.StatusOK, map[string]string{
		"status": "received",
	})
}

// ---------------------------- Event Handlers ----------------------------

// handleCheckoutSessionCompleted processes successful checkout sessions.
func (h *DefaultHandler) handleCheckoutSessionCompleted(ctx context.Context, event *StripeEvent) error {
	// Extract checkout session data
	sessionData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid checkout session data")
	}
	if h.db == nil {
		// No database configured (e.g., in tests). Skip persistence.
		return nil
	}

	customerID, _ := sessionData["customer"].(string)
	paymentStatus, _ := sessionData["payment_status"].(string)
	clientReferenceID, _ := sessionData["client_reference_id"].(string)

	log.Printf("Checkout completed - Customer: %s, Payment Status: %s, Reference: %s",
		customerID, paymentStatus, clientReferenceID)

	// Link Stripe customer to a user by client_reference_id (Auth0 sub or internal user ref)
	if clientReferenceID != "" && customerID != "" {
		userRepo := user.NewDefaultRepository(h.db)
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
func (h *DefaultHandler) handleSubscriptionCreated(ctx context.Context, event *StripeEvent) error {
	subscriptionData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data")
	}
	if h.db == nil {
		// No database configured (e.g., in tests). Skip persistence.
		return nil
	}

	customerID, _ := subscriptionData["customer"].(string)
	subscriptionID, _ := subscriptionData["id"].(string)
	status, _ := subscriptionData["status"].(string)

	log.Printf("Subscription created - Customer: %s, Subscription: %s, Status: %s",
		customerID, subscriptionID, status)

	// Persist subscription state
	subRepo := NewSubscriptionsRepository(h.db)
	userRepo := user.NewDefaultRepository(h.db)
	if customerID != "" && subscriptionID != "" {
		if u, err := userRepo.GetByStripeCustomerID(ctx, customerID); err == nil {
			// Extract optional subscription details
			var priceIDPtr *string
			if itemsRaw, ok := subscriptionData["items"].(map[string]interface{}); ok {
				if dataArr, ok := itemsRaw["data"].([]interface{}); ok && len(dataArr) > 0 {
					if firstItem, ok := dataArr[0].(map[string]interface{}); ok {
						if priceRaw, ok := firstItem["price"].(map[string]interface{}); ok {
							if pid, ok := priceRaw["id"].(string); ok && pid != "" {
								priceIDPtr = &pid
							}
						}
					}
				}
			}

			var (
				cpsPtr, cpePtr, cancelAtPtr, canceledAtPtr *time.Time
				cancelAtPeriodEnd                          bool
			)

			if v, ok := subscriptionData["current_period_start"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				cpsPtr = &t
			}
			if v, ok := subscriptionData["current_period_end"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				cpePtr = &t
			}
			if v, ok := subscriptionData["cancel_at"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				cancelAtPtr = &t
			}
			if v, ok := subscriptionData["canceled_at"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				canceledAtPtr = &t
			}
			if v, ok := subscriptionData["cancel_at_period_end"].(bool); ok {
				cancelAtPeriodEnd = v
			}

			if _, err := subRepo.UpsertByStripeID(ctx, u.ID.String(), subscriptionID, status, priceIDPtr, cpsPtr, cpePtr, cancelAtPtr, canceledAtPtr, cancelAtPeriodEnd); err != nil {
				log.Printf("Failed to upsert subscription (created): %v", err)
			}
		} else {
			log.Printf("No user found for Stripe customer on subscription.created: %s (err=%v)", customerID, err)
		}
	}
	return nil
}

// handleSubscriptionUpdated processes subscription update events.
func (h *DefaultHandler) handleSubscriptionUpdated(ctx context.Context, event *StripeEvent) error {
	subscriptionData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data")
	}
	if h.db == nil {
		// No database configured (e.g., in tests). Skip persistence.
		return nil
	}

	customerID, _ := subscriptionData["customer"].(string)
	subscriptionID, _ := subscriptionData["id"].(string)
	status, _ := subscriptionData["status"].(string)

	log.Printf("Subscription updated - Customer: %s, Subscription: %s, Status: %s",
		customerID, subscriptionID, status)

	// Persist subscription state
	subRepo := NewSubscriptionsRepository(h.db)
	userRepo := user.NewDefaultRepository(h.db)
	if customerID != "" && subscriptionID != "" {
		if u, err := userRepo.GetByStripeCustomerID(ctx, customerID); err == nil {
			// Extract optional subscription details
			var priceIDPtr *string
			if itemsRaw, ok := subscriptionData["items"].(map[string]interface{}); ok {
				if dataArr, ok := itemsRaw["data"].([]interface{}); ok && len(dataArr) > 0 {
					if firstItem, ok := dataArr[0].(map[string]interface{}); ok {
						if priceRaw, ok := firstItem["price"].(map[string]interface{}); ok {
							if pid, ok := priceRaw["id"].(string); ok && pid != "" {
								priceIDPtr = &pid
							}
						}
					}
				}
			}

			var (
				cpsPtr, cpePtr, cancelAtPtr, canceledAtPtr *time.Time
				cancelAtPeriodEnd                          bool
			)

			if v, ok := subscriptionData["current_period_start"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				cpsPtr = &t
			}
			if v, ok := subscriptionData["current_period_end"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				cpePtr = &t
			}
			if v, ok := subscriptionData["cancel_at"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				cancelAtPtr = &t
			}
			if v, ok := subscriptionData["canceled_at"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				canceledAtPtr = &t
			}
			if v, ok := subscriptionData["cancel_at_period_end"].(bool); ok {
				cancelAtPeriodEnd = v
			}

			if _, err := subRepo.UpsertByStripeID(ctx, u.ID.String(), subscriptionID, status, priceIDPtr, cpsPtr, cpePtr, cancelAtPtr, canceledAtPtr, cancelAtPeriodEnd); err != nil {
				log.Printf("Failed to upsert subscription (updated): %v", err)
			}
		} else {
			log.Printf("No user found for Stripe customer on subscription.updated: %s (err=%v)", customerID, err)
		}
	}
	return nil
}

// handleSubscriptionDeleted processes subscription cancellation events.
func (h *DefaultHandler) handleSubscriptionDeleted(ctx context.Context, event *StripeEvent) error {
	subscriptionData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data")
	}
	if h.db == nil {
		// No database configured (e.g., in tests). Skip persistence.
		return nil
	}

	customerID, _ := subscriptionData["customer"].(string)
	subscriptionID, _ := subscriptionData["id"].(string)

	log.Printf("Subscription deleted - Customer: %s, Subscription: %s",
		customerID, subscriptionID)

	// Mark subscription as canceled/deactivated
	subRepo := NewSubscriptionsRepository(h.db)
	userRepo := user.NewDefaultRepository(h.db)
	if customerID != "" && subscriptionID != "" {
		// Stripe sends a final status (typically "canceled"); persist it
		if u, err := userRepo.GetByStripeCustomerID(ctx, customerID); err == nil {
			// Extract optional subscription details even on deletion (Stripe often includes final state)
			var priceIDPtr *string
			if itemsRaw, ok := subscriptionData["items"].(map[string]interface{}); ok {
				if dataArr, ok := itemsRaw["data"].([]interface{}); ok && len(dataArr) > 0 {
					if firstItem, ok := dataArr[0].(map[string]interface{}); ok {
						if priceRaw, ok := firstItem["price"].(map[string]interface{}); ok {
							if pid, ok := priceRaw["id"].(string); ok && pid != "" {
								priceIDPtr = &pid
							}
						}
					}
				}
			}

			var (
				cpsPtr, cpePtr, cancelAtPtr, canceledAtPtr *time.Time
				cancelAtPeriodEnd                          bool
			)

			if v, ok := subscriptionData["current_period_start"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				cpsPtr = &t
			}
			if v, ok := subscriptionData["current_period_end"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				cpePtr = &t
			}
			if v, ok := subscriptionData["cancel_at"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				cancelAtPtr = &t
			}
			if v, ok := subscriptionData["canceled_at"].(float64); ok && v > 0 {
				t := time.Unix(int64(v), 0)
				canceledAtPtr = &t
			}
			if v, ok := subscriptionData["cancel_at_period_end"].(bool); ok {
				cancelAtPeriodEnd = v
			}

			if _, err := subRepo.UpsertByStripeID(ctx, u.ID.String(), subscriptionID, "canceled", priceIDPtr, cpsPtr, cpePtr, cancelAtPtr, canceledAtPtr, cancelAtPeriodEnd); err != nil {
				log.Printf("Failed to upsert subscription (deleted): %v", err)
			}
		} else {
			log.Printf("No user found for Stripe customer on subscription.deleted: %s (err=%v)", customerID, err)
		}
	}
	return nil
}

// handleInvoicePaymentSucceeded processes successful payment events.
func (h *DefaultHandler) handleInvoicePaymentSucceeded(ctx context.Context, event *StripeEvent) error {
	invoiceData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid invoice data")
	}
	if h.db == nil {
		// No database configured (e.g., in tests). Skip persistence.
		return nil
	}

	invoiceID, _ := invoiceData["id"].(string)
	customerID, _ := invoiceData["customer"].(string)
	subscriptionID, _ := invoiceData["subscription"].(string)
	status, _ := invoiceData["status"].(string)
	if status == "" {
		status = "paid"
	}

	var amountDueI, amountPaidI int32
	if v, ok := invoiceData["amount_due"].(float64); ok {
		amountDueI = int32(v)
	}
	if v, ok := invoiceData["amount_paid"].(float64); ok {
		amountPaidI = int32(v)
	}
	currency, _ := invoiceData["currency"].(string)
	invoiceNumber, _ := invoiceData["number"].(string)

	log.Printf("Invoice payment succeeded - Invoice: %s, Customer: %s, Subscription: %s, AmountPaid: %.2f",
		invoiceID, customerID, subscriptionID, float64(amountPaidI)/100)

	// Persist invoice
	if customerID != "" && invoiceID != "" {
		userRepo := user.NewDefaultRepository(h.db)
		if u, err := userRepo.GetByStripeCustomerID(ctx, customerID); err == nil {
			invRepo := NewInvoicesRepository(h.db)

			var subIDPtr, currencyPtr, invNumPtr *string
			if subscriptionID != "" {
				subIDPtr = &subscriptionID
			}
			if currency != "" {
				currencyPtr = &currency
			}
			if invoiceNumber != "" {
				invNumPtr = &invoiceNumber
			}

			if _, err := invRepo.Upsert(ctx, u.ID.String(), invoiceID, subIDPtr, status, amountDueI, amountPaidI, currencyPtr, invNumPtr); err != nil {
				log.Printf("Failed to upsert invoice (payment_succeeded): %v", err)
			}
		} else {
			log.Printf("No user found for Stripe customer on invoice.payment_succeeded: %s (err=%v)", customerID, err)
		}
	}

	return nil
}

// handleInvoicePaymentFailed processes failed payment events.
func (h *DefaultHandler) handleInvoicePaymentFailed(ctx context.Context, event *StripeEvent) error {
	invoiceData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid invoice data")
	}
	if h.db == nil {
		// No database configured (e.g., in tests). Skip persistence.
		return nil
	}

	invoiceID, _ := invoiceData["id"].(string)
	customerID, _ := invoiceData["customer"].(string)
	subscriptionID, _ := invoiceData["subscription"].(string)
	status, _ := invoiceData["status"].(string)
	if status == "" {
		status = "failed"
	}

	var amountDueI, amountPaidI int32
	if v, ok := invoiceData["amount_due"].(float64); ok {
		amountDueI = int32(v)
	}
	if v, ok := invoiceData["amount_paid"].(float64); ok {
		amountPaidI = int32(v)
	}
	currency, _ := invoiceData["currency"].(string)
	invoiceNumber, _ := invoiceData["number"].(string)

	log.Printf("Invoice payment failed - Invoice: %s, Customer: %s, Subscription: %s",
		invoiceID, customerID, subscriptionID)

	// Persist invoice with failed status
	if customerID != "" && invoiceID != "" {
		userRepo := user.NewDefaultRepository(h.db)
		if u, err := userRepo.GetByStripeCustomerID(ctx, customerID); err == nil {
			invRepo := NewInvoicesRepository(h.db)

			var subIDPtr, currencyPtr, invNumPtr *string
			if subscriptionID != "" {
				subIDPtr = &subscriptionID
			}
			if currency != "" {
				currencyPtr = &currency
			}
			if invoiceNumber != "" {
				invNumPtr = &invoiceNumber
			}

			if _, err := invRepo.Upsert(ctx, u.ID.String(), invoiceID, subIDPtr, status, amountDueI, amountPaidI, currencyPtr, invNumPtr); err != nil {
				log.Printf("Failed to upsert invoice (payment_failed): %v", err)
			}
		} else {
			log.Printf("No user found for Stripe customer on invoice.payment_failed: %s (err=%v)", customerID, err)
		}
	}

	return nil
}

// handleCustomerCreated processes new customer events.
func (h *DefaultHandler) handleCustomerCreated(_ context.Context, event *StripeEvent) error {
	customerData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid customer data")
	}

	customerID, _ := customerData["id"].(string)
	email, _ := customerData["email"].(string)

	log.Printf("Customer created: ID=%s, Email=%s", customerID, email)
	// TODO: Implement customer creation logic (optional)

	return nil
}

// handleCustomerUpdated processes customer update events.
func (h *DefaultHandler) handleCustomerUpdated(_ context.Context, event *StripeEvent) error {
	customerData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid customer data")
	}

	customerID, _ := customerData["id"].(string)
	email, _ := customerData["email"].(string)

	log.Printf("Customer updated: ID=%s, Email=%s", customerID, email)
	// TODO: Implement customer update logic (optional)

	return nil
}

// handleCustomerDeleted processes customer deletion events.
func (h *DefaultHandler) handleCustomerDeleted(_ context.Context, event *StripeEvent) error {
	customerData, ok := event.Data["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid customer data")
	}

	customerID, _ := customerData["id"].(string)

	log.Printf("Customer deleted: ID=%s", customerID)
	// TODO: Implement customer deletion logic (optional)

	return nil
}

// ---------------------------- Signature Helpers ----------------------------

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

// ---------------------------- Idempotency Helpers ----------------------------

// alreadyProcessedStripeEvent checks if the event was already processed (idempotency scaffold).
// Backed by processed_events; when db is nil (tests), treat as not processed.
func (h *DefaultHandler) alreadyProcessedStripeEvent(ctx context.Context, eventID string) (bool, error) {
	if h.db == nil {
		// No database configured (e.g., in tests). Treat as not processed.
		return false, nil
	}
	repo := NewProcessedEventsRepository(h.db)
	return repo.IsProcessed(ctx, eventID)
}

// markStripeEventProcessed records the processed event (idempotency scaffold).
// Stores the type and raw payload when available.
func (h *DefaultHandler) markStripeEventProcessed(ctx context.Context, eventID, eventType string, payload []byte) error {
	if h.db == nil {
		// No database configured (e.g., in tests). Skip persistence.
		return nil
	}
	repo := NewProcessedEventsRepository(h.db)
	_, err := repo.Upsert(ctx, eventID, &eventType, payload)
	return err
}
