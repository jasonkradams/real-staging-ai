package billing

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
)

// Package billing provides interfaces and DTOs for billing-related endpoints,
// mirroring the current HTTP handlers for subscriptions and invoices.
//
// It intentionally separates HTTP handler contracts (Handler) from business/data
// service contracts (Service), and defines transport-friendly DTOs decoupled
// from storage-layer types.

//go:generate go run github.com/matryer/moq@v0.5.3 -out billing_mock.go . Handler Service

// Handler defines the HTTP-level billing handlers, typically wired into Echo routes.
//
// Expected routes (as used by the API server):
// - GET /api/v1/billing/subscriptions
// - GET /api/v1/billing/invoices
type Handler interface {
	// GetMySubscriptions returns the current user's subscriptions with pagination.
	GetMySubscriptions(c echo.Context) error
	// GetMyInvoices returns the current user's invoices with pagination.
	GetMyInvoices(c echo.Context) error
}

// Service defines the data/logic contract used by billing handlers.
//
// Implementations typically resolve data from repositories (e.g., Stripe-related
// subscription and invoice tables) for a given user.
type Service interface {
	// ListSubscriptionsForUser lists subscriptions for a user with pagination.
	ListSubscriptionsForUser(ctx context.Context, userID string, p Pagination) ([]SubscriptionDTO, error)
	// ListInvoicesForUser lists invoices for a user with pagination.
	ListInvoicesForUser(ctx context.Context, userID string, p Pagination) ([]InvoiceDTO, error)
}

// SubscriptionDTO mirrors the shape exposed by the billing subscriptions endpoint.
type SubscriptionDTO struct {
	ID                   string     `json:"id"`
	StripeSubscriptionID string     `json:"stripe_subscription_id"`
	Status               string     `json:"status"`
	PriceID              *string    `json:"price_id,omitempty"`
	CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty"`
	CancelAt             *time.Time `json:"cancel_at,omitempty"`
	CanceledAt           *time.Time `json:"canceled_at,omitempty"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// InvoiceDTO mirrors the shape exposed by the billing invoices endpoint.
type InvoiceDTO struct {
	ID                   string    `json:"id"`
	StripeInvoiceID      string    `json:"stripe_invoice_id"`
	StripeSubscriptionID *string   `json:"stripe_subscription_id,omitempty"`
	Status               string    `json:"status"`
	AmountDue            int32     `json:"amount_due"`
	AmountPaid           int32     `json:"amount_paid"`
	Currency             *string   `json:"currency,omitempty"`
	InvoiceNumber        *string   `json:"invoice_number,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// ListResponse is a generic pagination wrapper for list endpoints.
type ListResponse[T any] struct {
	Items  []T   `json:"items"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

// Pagination captures common pagination inputs.
type Pagination struct {
	Limit  int32
	Offset int32
}

const (
	// DefaultLimit is the default number of items to return when not specified.
	DefaultLimit int32 = 50
	// MaxLimit is the maximum allowed number of items per page.
	MaxLimit int32 = 100
)

// NormalizeLimitOffset applies defaults and caps to the provided limit/offset.
func NormalizeLimitOffset(limit, offset int32) (int32, int32) {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
