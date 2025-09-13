package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/virtual-staging-ai/api/internal/auth"
	"github.com/virtual-staging-ai/api/internal/stripe"
	"github.com/virtual-staging-ai/api/internal/user"
)

// SubscriptionDTO is the API-facing view of a subscription.
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

// InvoiceDTO is the API-facing view of an invoice.
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

// listResponse is a generic wrapper for paginated list endpoints.
type listResponse[T any] struct {
	Items  []T   `json:"items"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

// getMySubscriptionsHandler lists subscriptions for the current user (derived from JWT or test header).
func (s *Server) getMySubscriptionsHandler(c echo.Context) error {
	limit, offset := parseLimitOffset(c)

	// No DB configured (tests or special environments): return empty list gracefully.
	if s.db == nil {
		return c.JSON(http.StatusOK, listResponse[SubscriptionDTO]{Items: []SubscriptionDTO{}, Limit: limit, Offset: offset})
	}

	// Resolve current user (Auth0 sub or test header) and ensure a users row exists.
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing user identity",
		})
	}

	userRepo := user.NewUserRepository(s.db)
	u, err := userRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Create a user row on-the-fly to keep flows consistent.
			u, err = userRepo.Create(c.Request().Context(), auth0Sub, "", "user")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "internal_server_error",
					Message: "Failed to create user",
				})
			}
		} else {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to resolve user",
			})
		}
	}

	subRepo := stripe.NewSubscriptionsRepository(s.db)
	rows, err := subRepo.ListByUserID(c.Request().Context(), u.ID.String(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to list subscriptions",
		})
	}

	items := make([]SubscriptionDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, SubscriptionDTO{
			ID:                   uuidToString(r.ID),
			StripeSubscriptionID: r.StripeSubscriptionID,
			Status:               r.Status,
			PriceID:              textPtr(r.PriceID),
			CurrentPeriodStart:   timePtr(r.CurrentPeriodStart),
			CurrentPeriodEnd:     timePtr(r.CurrentPeriodEnd),
			CancelAt:             timePtr(r.CancelAt),
			CanceledAt:           timePtr(r.CanceledAt),
			CancelAtPeriodEnd:    r.CancelAtPeriodEnd,
			CreatedAt:            r.CreatedAt.Time,
			UpdatedAt:            r.UpdatedAt.Time,
		})
	}

	return c.JSON(http.StatusOK, listResponse[SubscriptionDTO]{Items: items, Limit: limit, Offset: offset})
}

// getMyInvoicesHandler lists invoices for the current user (derived from JWT or test header).
func (s *Server) getMyInvoicesHandler(c echo.Context) error {
	limit, offset := parseLimitOffset(c)

	// No DB configured (tests or special environments): return empty list gracefully.
	if s.db == nil {
		return c.JSON(http.StatusOK, listResponse[InvoiceDTO]{Items: []InvoiceDTO{}, Limit: limit, Offset: offset})
	}

	// Resolve current user (Auth0 sub or test header) and ensure a users row exists.
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing user identity",
		})
	}

	userRepo := user.NewUserRepository(s.db)
	u, err := userRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Create a user row on-the-fly to keep flows consistent.
			u, err = userRepo.Create(c.Request().Context(), auth0Sub, "", "user")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "internal_server_error",
					Message: "Failed to create user",
				})
			}
		} else {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to resolve user",
			})
		}
	}

	invRepo := stripe.NewInvoicesRepository(s.db)
	rows, err := invRepo.ListByUserID(c.Request().Context(), u.ID.String(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to list invoices",
		})
	}

	items := make([]InvoiceDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, InvoiceDTO{
			ID:                   uuidToString(r.ID),
			StripeInvoiceID:      r.StripeInvoiceID,
			StripeSubscriptionID: textPtr(r.StripeSubscriptionID),
			Status:               r.Status,
			AmountDue:            r.AmountDue,
			AmountPaid:           r.AmountPaid,
			Currency:             textPtr(r.Currency),
			InvoiceNumber:        textPtr(r.InvoiceNumber),
			CreatedAt:            r.CreatedAt.Time,
			UpdatedAt:            r.UpdatedAt.Time,
		})
	}

	return c.JSON(http.StatusOK, listResponse[InvoiceDTO]{Items: items, Limit: limit, Offset: offset})
}

// parseLimitOffset reads 'limit' and 'offset' query params with sane defaults and caps.
func parseLimitOffset(c echo.Context) (int32, int32) {
	const (
		defaultLimit = int32(50)
		maxLimit     = int32(100)
	)
	limit := defaultLimit
	offset := int32(0)

	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if int32(n) > maxLimit {
				limit = maxLimit
			} else {
				limit = int32(n)
			}
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	return limit, offset
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

func textPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func timePtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}
