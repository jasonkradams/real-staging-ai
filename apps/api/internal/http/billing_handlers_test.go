package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestGetMySubscriptionsHandler_DBless_EmptyAndPaginationDefaults(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/subscriptions", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Server with nil db triggers the db-less early return path.
	s := &Server{}

	if err := s.getMySubscriptionsHandler(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Items  []any `json:"items"`
		Limit  int   `json:"limit"`
		Offset int   `json:"offset"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Items) != 0 {
		t.Fatalf("expected empty items, got %d", len(resp.Items))
	}
	if resp.Limit != 50 {
		t.Fatalf("expected default limit=50, got %d", resp.Limit)
	}
	if resp.Offset != 0 {
		t.Fatalf("expected default offset=0, got %d", resp.Offset)
	}
}

func TestGetMyInvoicesHandler_DBless_EmptyAndPaginationCap(t *testing.T) {
	e := echo.New()
	// Request with limit beyond cap and a non-zero offset
	req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/invoices?limit=200&offset=5", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Server with nil db triggers the db-less early return path.
	s := &Server{}

	if err := s.getMyInvoicesHandler(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Items  []any `json:"items"`
		Limit  int   `json:"limit"`
		Offset int   `json:"offset"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Items) != 0 {
		t.Fatalf("expected empty items, got %d", len(resp.Items))
	}
	// Limit should be capped to 100
	if resp.Limit != 100 {
		t.Fatalf("expected capped limit=100, got %d", resp.Limit)
	}
	if resp.Offset != 5 {
		t.Fatalf("expected offset=5, got %d", resp.Offset)
	}
}
