package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
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
	if webhookSecret != "" {
		// TODO: Implement signature verification
		// stripeSignature := c.Request().Header.Get("Stripe-Signature")
		// if !verifyStripeSignature(body, stripeSignature, webhookSecret) {
		//     return c.JSON(http.StatusUnauthorized, ErrorResponse{
		//         Error:   "unauthorized",
		//         Message: "Invalid webhook signature",
		//     })
		// }
	}

	// Parse the webhook event
	var event StripeEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing webhook JSON: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format")
	}

	log.Printf("Received Stripe webhook event: %s (ID: %s)", event.Type, event.ID)

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

	// TODO: Update user's subscription status in the database
	// This would typically involve:
	// 1. Finding the user by client_reference_id (user ID)
	// 2. Updating their Stripe customer ID
	// 3. Activating their subscription
	// 4. Sending a welcome email

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

	// TODO: Update user's subscription in the database
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

	// TODO: Update user's subscription status in the database
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

	// TODO: Deactivate user's subscription in the database
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
