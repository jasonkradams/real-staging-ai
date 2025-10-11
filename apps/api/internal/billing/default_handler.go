package billing

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/real-staging-ai/api/internal/auth"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/stripe"
	"github.com/real-staging-ai/api/internal/user"
)

// DefaultHandler implements the billing Handler by wrapping existing repositories
// and user resolution logic (Auth0 sub -> ensure users row).
type DefaultHandler struct {
	db storage.Database
}

// NewDefaultHandler constructs a DefaultHandler.
func NewDefaultHandler(db storage.Database) *DefaultHandler {
	return &DefaultHandler{db: db}
}

// ErrorResponse is a simple JSON error envelope for handler responses.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// GetMySubscriptions returns the current user's subscriptions (paginated).
func (h *DefaultHandler) GetMySubscriptions(c echo.Context) error {
	limit, offset := h.parseLimitOffset(c)

	// No DB configured (e.g., special test mode) — return empty list gracefully.
	if h.db == nil {
		return c.JSON(http.StatusOK, ListResponse[SubscriptionDTO]{Items: []SubscriptionDTO{}, Limit: limit, Offset: offset})
	}

	// Resolve current user (Auth0 sub or test header) and ensure a users row exists.
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	uRepo := user.NewDefaultRepository(h.db)
	u, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to resolve user",
		})
	}

	subRepo := stripe.NewSubscriptionsRepository(h.db)
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

	return c.JSON(http.StatusOK, ListResponse[SubscriptionDTO]{Items: items, Limit: limit, Offset: offset})
}

// GetMyInvoices returns the current user's invoices (paginated).
func (h *DefaultHandler) GetMyInvoices(c echo.Context) error {
	limit, offset := h.parseLimitOffset(c)

	// No DB configured (e.g., special test mode) — return empty list gracefully.
	if h.db == nil {
		return c.JSON(http.StatusOK, ListResponse[InvoiceDTO]{Items: []InvoiceDTO{}, Limit: limit, Offset: offset})
	}

	// Resolve current user (Auth0 sub or test header) and ensure a users row exists.
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil || auth0Sub == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Unable to resolve current user",
		})
	}

	uRepo := user.NewDefaultRepository(h.db)
	u, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to resolve user",
		})
	}

	invRepo := stripe.NewInvoicesRepository(h.db)
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

	return c.JSON(http.StatusOK, ListResponse[InvoiceDTO]{Items: items, Limit: limit, Offset: offset})
}

// parseLimitOffset reads limit/offset from query params and applies defaults/caps.
func (h *DefaultHandler) parseLimitOffset(c echo.Context) (int32, int32) {
	limit := DefaultLimit
	offset := int32(0)

	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}

	return limit, offset
}

// Helper mappers for sqlc/pgx types into DTO pointers.

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return pgUUIDToString(u)
}

// pgUUIDToString converts a pgtype.UUID to its canonical string form.
func pgUUIDToString(u pgtype.UUID) string {
	// pgtype.UUID.Bytes is a [16]byte; String() typically requires github.com/google/uuid.
	// Avoid importing another dep here by delegating to the repository’s string conversion,
	// but since we don't have it, reconstruct via the standard formatting.
	b := u.Bytes
	// Format as 8-4-4-4-12
	return formatUUIDBytes(b)
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

// formatUUIDBytes formats a 16-byte UUID into a canonical string.
// This avoids importing github.com/google/uuid just for formatting.
func formatUUIDBytes(b [16]byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, 36)

	writeByte := func(dst []byte, v byte) {
		dst[0] = hexdigits[v>>4]
		dst[1] = hexdigits[v&0x0f]
	}

	j := 0
	for i := 0; i < 16; i++ {
		switch i {
		case 4, 6, 8, 10:
			out[j] = '-'
			j++
		}
		writeByte(out[j:j+2], b[i])
		j += 2
	}
	return string(out)
}
