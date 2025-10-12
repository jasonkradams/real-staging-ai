package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/real-staging-ai/api/internal/storage"
)

// Table-driven tests for GetMySubscriptions
func TestGetMySubscriptions(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		query          string
		expectedStatus int
	}{
		{
			name:           "success",
			userID:         "auth0|testuser",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "with pagination",
			userID:         "auth0|testuser",
			query:          "limit=10&offset=5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no database",
			userID:         "auth0|testuser",
			query:          "",
			expectedStatus: http.StatusOK,
		},
	}

	handler := NewDefaultHandler(nil)
	runHandlerTableTest(t, handler.GetMySubscriptions, tests)
}

// Minimal pgx.Row stub used for DatabaseMock.QueryRow responses
type rowStub struct{ scan func(dest ...any) error }

func (r rowStub) Scan(dest ...any) error { return r.scan(dest...) }

// small error sentinel to avoid importing fmt everywhere
type tinyError string

func (e tinyError) Error() string { return string(e) }

func errBoom() error { return tinyError("boom") }

// rowsIterStub implements pgx.Rows for sqlc list scans in tests.
type rowsIterStub struct {
	idx   int
	scans []func(dest ...any) error
	err   error
}

func (r *rowsIterStub) Next() bool {
	if r.idx < len(r.scans) {
		r.idx++
		return true
	}
	return false
}

func (r *rowsIterStub) Scan(dest ...any) error {
	if r.idx == 0 || r.idx > len(r.scans) {
		return nil
	}
	return r.scans[r.idx-1](dest...)
}

func (r *rowsIterStub) Values() ([]any, error)                       { return nil, nil }
func (r *rowsIterStub) RawValues() [][]byte                          { return nil }
func (r *rowsIterStub) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *rowsIterStub) Err() error                                   { return r.err }
func (r *rowsIterStub) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *rowsIterStub) Close()                                       {}
func (r *rowsIterStub) Conn() *pgx.Conn                              { return nil }

// Helper to create a mock user row for tests
func mockUserRow(now time.Time) func(dest ...any) error {
	return func(dest ...any) error {
		dest[0].(*pgtype.UUID).Bytes = uuid.New()
		dest[0].(*pgtype.UUID).Valid = true
		*dest[1].(*string) = "auth0|testuser"
		*dest[2].(*pgtype.Text) = pgtype.Text{}
		*dest[3].(*string) = "user"
		*dest[4].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now, Valid: true}
		return nil
	}
}

// Helper to run table-driven tests for handler methods
func runHandlerTableTest(t *testing.T, handlerFunc func(echo.Context) error, tests []struct {
	name           string
	userID         string
	query          string
	expectedStatus int
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			target := "/"
			if tt.query != "" {
				target = "/?" + tt.query
			}
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.Header.Set("X-Test-User", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := handlerFunc(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

// --- Subscriptions: DB-backed tests ---

// Unauthorized when db != nil and JWT sub is empty (no X-Test-User header)
func TestGetMySubscriptions_DB_Unauthorized(t *testing.T) {
	h := NewDefaultHandler(&storage.DatabaseMock{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"sub": ""}})

	if err := h.GetMySubscriptions(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetMySubscriptions_DB_UserResolveError(t *testing.T) {
	db := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return rowStub{scan: func(dest ...any) error { return errBoom() }}
		},
	}
	h := NewDefaultHandler(db)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Test-User", "auth0|testuser")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetMySubscriptions(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// Helper to test DB list errors for both subscriptions and invoices
func testDBListError(t *testing.T, getHandler func(*storage.DatabaseMock) func(echo.Context) error, query string) {
	now := time.Now()
	db := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return rowStub{scan: mockUserRow(now)}
		},
		QueryFunc: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
			return nil, errBoom()
		},
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+query, nil)
	req.Header.Set("X-Test-User", "auth0|testuser")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := getHandler(db)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetMySubscriptions_DB_ListError(t *testing.T) {
	testDBListError(t, func(db *storage.DatabaseMock) func(echo.Context) error {
		return NewDefaultHandler(db).GetMySubscriptions
	}, "limit=9999&offset=-5")
}

func TestGetMySubscriptions_DB_SuccessMapping(t *testing.T) {
	now := time.Now()
	rows := &rowsIterStub{
		scans: []func(dest ...any) error{
			func(dest ...any) error {
				// First row
				dest[0].(*pgtype.UUID).Bytes = uuid.New()
				dest[0].(*pgtype.UUID).Valid = true
				dest[1].(*pgtype.UUID).Bytes = uuid.New()
				dest[1].(*pgtype.UUID).Valid = true
				*dest[2].(*string) = "sub_1"
				*dest[3].(*string) = "active"
				*dest[4].(*pgtype.Text) = pgtype.Text{String: "price_1", Valid: true}
				*dest[5].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true}
				*dest[6].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true}
				*dest[7].(*pgtype.Timestamptz) = pgtype.Timestamptz{}
				*dest[8].(*pgtype.Timestamptz) = pgtype.Timestamptz{}
				*dest[9].(*bool) = true
				*dest[10].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now, Valid: true}
				*dest[11].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now, Valid: true}
				return nil
			},
			func(dest ...any) error {
				// Second row (some NULL-able fields)
				*dest[0].(*pgtype.UUID) = pgtype.UUID{}
				dest[1].(*pgtype.UUID).Bytes = uuid.New()
				dest[1].(*pgtype.UUID).Valid = true
				*dest[2].(*string) = "sub_2"
				*dest[3].(*string) = "canceled"
				*dest[4].(*pgtype.Text) = pgtype.Text{}
				*dest[5].(*pgtype.Timestamptz) = pgtype.Timestamptz{}
				*dest[6].(*pgtype.Timestamptz) = pgtype.Timestamptz{}
				*dest[7].(*pgtype.Timestamptz) = pgtype.Timestamptz{}
				*dest[8].(*pgtype.Timestamptz) = pgtype.Timestamptz{}
				*dest[9].(*bool) = false
				*dest[10].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now, Valid: true}
				*dest[11].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now, Valid: true}
				return nil
			},
		},
	}

	db := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return rowStub{scan: mockUserRow(now)}
		},
		QueryFunc: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
			return rows, nil
		},
	}
	h := NewDefaultHandler(db)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?limit=100000&offset=-10", nil)
	req.Header.Set("X-Test-User", "auth0|testuser")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetMySubscriptions(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- Invoices: DB-backed tests ---

func TestGetMyInvoices_DB_Unauthorized(t *testing.T) {
	h := NewDefaultHandler(&storage.DatabaseMock{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"sub": ""}})

	if err := h.GetMyInvoices(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetMyInvoices_DB_UserResolveError(t *testing.T) {
	db := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return rowStub{scan: func(dest ...any) error { return errBoom() }}
		},
	}
	h := NewDefaultHandler(db)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Test-User", "auth0|testuser")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetMyInvoices(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetMyInvoices_DB_ListError(t *testing.T) {
	testDBListError(t, func(db *storage.DatabaseMock) func(echo.Context) error {
		return NewDefaultHandler(db).GetMyInvoices
	}, "limit=0&offset=-1")
}

func TestGetMyInvoices_DB_SuccessMapping(t *testing.T) {
	now := time.Now()
	rows := &rowsIterStub{
		scans: []func(dest ...any) error{
			func(dest ...any) error {
				// First row
				dest[0].(*pgtype.UUID).Bytes = uuid.New()
				dest[0].(*pgtype.UUID).Valid = true
				dest[1].(*pgtype.UUID).Bytes = uuid.New()
				dest[1].(*pgtype.UUID).Valid = true
				*dest[2].(*string) = "in_1"
				*dest[3].(*pgtype.Text) = pgtype.Text{String: "sub_1", Valid: true}
				*dest[4].(*string) = "paid"
				*dest[5].(*int32) = 1000
				*dest[6].(*int32) = 1000
				*dest[7].(*pgtype.Text) = pgtype.Text{String: "usd", Valid: true}
				*dest[8].(*pgtype.Text) = pgtype.Text{String: "INV-1", Valid: true}
				*dest[9].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now, Valid: true}
				*dest[10].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now, Valid: true}
				return nil
			},
			func(dest ...any) error {
				// Second row
				*dest[0].(*pgtype.UUID) = pgtype.UUID{}
				dest[1].(*pgtype.UUID).Bytes = uuid.New()
				dest[1].(*pgtype.UUID).Valid = true
				*dest[2].(*string) = "in_2"
				*dest[3].(*pgtype.Text) = pgtype.Text{}
				*dest[4].(*string) = "open"
				*dest[5].(*int32) = 0
				*dest[6].(*int32) = 0
				*dest[7].(*pgtype.Text) = pgtype.Text{}
				*dest[8].(*pgtype.Text) = pgtype.Text{}
				*dest[9].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now, Valid: true}
				*dest[10].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: now, Valid: true}
				return nil
			},
		},
	}

	db := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return rowStub{scan: mockUserRow(now)}
		},
		QueryFunc: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
			return rows, nil
		},
	}
	h := NewDefaultHandler(db)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?limit=3&offset=2", nil)
	req.Header.Set("X-Test-User", "auth0|testuser")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetMyInvoices(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- parseLimitOffset direct tests ---
func Test_parseLimitOffset(t *testing.T) {
	h := NewDefaultHandler(nil)
	cases := []struct {
		name       string
		query      string
		wantLimit  int32
		wantOffset int32
	}{
		{name: "success: defaults", query: "", wantLimit: DefaultLimit, wantOffset: 0},
		{name: "success: zero limit -> default", query: "limit=0", wantLimit: DefaultLimit, wantOffset: 0},
		{name: "success: negative limit -> default", query: "limit=-5", wantLimit: DefaultLimit, wantOffset: 0},
		{name: "success: below max limit", query: "limit=75", wantLimit: 75, wantOffset: 0},
		{name: "success: at max limit", query: "limit=100", wantLimit: MaxLimit, wantOffset: 0},
		{name: "success: above max -> default", query: "limit=150", wantLimit: DefaultLimit, wantOffset: 0},
		{name: "success: way above max -> default", query: "limit=100000", wantLimit: DefaultLimit, wantOffset: 0},
		{name: "success: neg offset ignored", query: "offset=-7", wantLimit: DefaultLimit, wantOffset: 0},
		{name: "success: pos offset set", query: "offset=12", wantLimit: DefaultLimit, wantOffset: 12},
		{name: "success: both set within limits", query: "limit=7&offset=3", wantLimit: 7, wantOffset: 3},
		{name: "success: above max with offset", query: "limit=200&offset=50", wantLimit: DefaultLimit, wantOffset: 50},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			target := "/"
			if tc.query != "" {
				target = "/?" + tc.query
			}
			req := httptest.NewRequest(http.MethodGet, target, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			gotL, gotO := h.parseLimitOffset(c)
			if gotL != tc.wantLimit || gotO != tc.wantOffset {
				t.Fatalf("limit/offset got (%d,%d) want (%d,%d)", gotL, gotO, tc.wantLimit, tc.wantOffset)
			}
		})
	}
}

// Table-driven tests for GetMyInvoices
func TestGetMyInvoices(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		query          string
		expectedStatus int
	}{
		{
			name:           "success",
			userID:         "auth0|testuser",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "with pagination",
			userID:         "auth0|testuser",
			query:          "limit=5&offset=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no database",
			userID:         "auth0|testuser",
			query:          "",
			expectedStatus: http.StatusOK,
		},
	}

	handler := NewDefaultHandler(nil)
	runHandlerTableTest(t, handler.GetMyInvoices, tests)
}
