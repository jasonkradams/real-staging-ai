package stripe

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"

	"github.com/virtual-staging-ai/api/internal/storage"
)

// helper to build Stripe-Signature header string with given timestamp and one or more v1 signatures.
func makeSigHeader(ts int64, sigs ...[]byte) string {
	header := fmt.Sprintf("t=%d", ts)
	for _, s := range sigs {
		header += fmt.Sprintf(",v1=%s", hex.EncodeToString(s))
	}
	return header
}

func TestParseStripeSignatureHeader_Success(t *testing.T) {
	ts := int64(1700000000)
	// two valid hex signatures: 0x00 and 0xaa
	header := fmt.Sprintf("t=%d,v1=00,v1=aa", ts)

	gotTS, v1s, err := parseStripeSignatureHeader(header)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gotTS != ts {
		t.Fatalf("expected ts=%d, got %d", ts, gotTS)
	}
	if len(v1s) != 2 {
		t.Fatalf("expected 2 signatures, got %d", len(v1s))
	}
	if !bytes.Equal(v1s[0], []byte{0x00}) {
		t.Fatalf("expected first sig = 00, got %x", v1s[0])
	}
	if !bytes.Equal(v1s[1], []byte{0xaa}) {
		t.Fatalf("expected second sig = aa, got %x", v1s[1])
	}
}

func TestParseStripeSignatureHeader_MissingTimestamp(t *testing.T) {
	header := "v1=00"

	_, _, err := parseStripeSignatureHeader(header)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if want := "missing timestamp"; err == nil || !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestParseStripeSignatureHeader_InvalidTimestamp(t *testing.T) {
	header := "t=abc,v1=00"

	_, _, err := parseStripeSignatureHeader(header)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if want := "invalid timestamp"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestParseStripeSignatureHeader_InvalidV1Hex(t *testing.T) {
	header := "t=1700000000,v1=zz"

	_, _, err := parseStripeSignatureHeader(header)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if want := "invalid v1 signature"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestVerifyStripeSignature_Valid(t *testing.T) {
	body := []byte(`{"id":"evt_123"}`)
	secret := "whsec_test"
	ts := int64(1700000000)

	expected := computeStripeSignature(body, ts, secret)
	header := makeSigHeader(ts, expected)

	now := func() time.Time { return time.Unix(ts+60, 0) } // within 5 minutes
	if err := verifyStripeSignature(body, header, secret, 5*time.Minute, now); err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}
}

func TestVerifyStripeSignature_NoHeader(t *testing.T) {
	body := []byte(`{}`)
	secret := "whsec_test"
	now := func() time.Time { return time.Unix(1700000000, 0) }

	err := verifyStripeSignature(body, "", secret, 5*time.Minute, now)
	if err == nil {
		t.Fatalf("expected error for missing header, got nil")
	}
	if want := "missing Stripe-Signature header"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestVerifyStripeSignature_NoV1Provided(t *testing.T) {
	body := []byte(`{}`)
	secret := "whsec_test"
	ts := int64(1700000000)
	header := fmt.Sprintf("t=%d", ts)
	now := func() time.Time { return time.Unix(ts, 0) }

	err := verifyStripeSignature(body, header, secret, 5*time.Minute, now)
	if err == nil {
		t.Fatalf("expected error when no v1 signatures, got nil")
	}
	if want := "no v1 signatures found"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestVerifyStripeSignature_MismatchedSignature(t *testing.T) {
	body := []byte(`{"id":"evt_123"}`)
	secret := "whsec_test"
	badSecret := "whsec_other"
	ts := int64(1700000000)

	// signature computed with wrong secret
	wrong := computeStripeSignature(body, ts, badSecret)
	header := makeSigHeader(ts, wrong)

	now := func() time.Time { return time.Unix(ts, 0) }
	err := verifyStripeSignature(body, header, secret, 5*time.Minute, now)
	if err == nil {
		t.Fatalf("expected signature mismatch error, got nil")
	}
	if want := "no matching signature"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestVerifyStripeSignature_MultipleV1_OneMatches(t *testing.T) {
	body := []byte(`{"ok":true}`)
	secret := "whsec_test"
	ts := int64(1700000000)

	// first sig is bogus but valid hex, second is correct
	bogus := []byte("bogus-signature-bytes") // arbitrary
	good := computeStripeSignature(body, ts, secret)
	header := makeSigHeader(ts, bogus, good)

	now := func() time.Time { return time.Unix(ts, 0) }
	if err := verifyStripeSignature(body, header, secret, 5*time.Minute, now); err != nil {
		t.Fatalf("expected success when one of multiple signatures matches, got %v", err)
	}
}

func TestVerifyStripeSignature_TimestampOutsideTolerance(t *testing.T) {
	body := []byte(`{"id":"evt_999"}`)
	secret := "whsec_test"
	ts := int64(1700000000)

	sig := computeStripeSignature(body, ts, secret)
	header := makeSigHeader(ts, sig)

	// now is 10 minutes ahead, tolerance 5 minutes -> should fail
	now := func() time.Time { return time.Unix(ts, 0).Add(10 * time.Minute) }
	err := verifyStripeSignature(body, header, secret, 5*time.Minute, now)
	if err == nil {
		t.Fatalf("expected timestamp tolerance error, got nil")
	}
	if want := "timestamp outside tolerance window"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

// contains is a tiny helper to avoid pulling extra deps in tests.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		// naive substring search
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}

// ---------------------------- Direct DB-branch handler tests ----------------------------

// simpleDB returns okRow for all QueryRow calls to exercise DB-backed branches.
type simpleDB struct{}

func (s *simpleDB) Close() {}

func (s *simpleDB) Pool() storage.PgxPool { return nil }

func (s *simpleDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return okRow{}
}

func (s *simpleDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (s *simpleDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func Test_handleSubscriptionCreated_DB_Branch(t *testing.T) {
	h := NewDefaultHandler(&simpleDB{})

	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"customer": "cus_1",
				"id":       "sub_1",
				"status":   "active",
			},
		},
	}

	if err := h.handleSubscriptionCreated(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_handleSubscriptionUpdated_DB_Branch(t *testing.T) {
	h := NewDefaultHandler(&simpleDB{})

	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"customer": "cus_2",
				"id":       "sub_2",
				"status":   "past_due",
			},
		},
	}

	if err := h.handleSubscriptionUpdated(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_handleSubscriptionDeleted_DB_Branch(t *testing.T) {
	h := NewDefaultHandler(&simpleDB{})

	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"customer": "cus_3",
				"id":       "sub_3",
			},
		},
	}

	if err := h.handleSubscriptionDeleted(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_handleInvoicePaymentSucceeded_DB_Branch(t *testing.T) {
	h := NewDefaultHandler(&simpleDB{})

	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"id":           "in_1",
				"customer":     "cus_1",
				"subscription": "sub_1",
				"status":       "paid",
				"amount_due":   float64(1000),
				"amount_paid":  float64(1000),
				"currency":     "usd",
				"number":       "F-1001",
			},
		},
	}

	if err := h.handleInvoicePaymentSucceeded(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_handleInvoicePaymentFailed_DB_Branch(t *testing.T) {
	h := NewDefaultHandler(&simpleDB{})

	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"id":           "in_2",
				"customer":     "cus_2",
				"subscription": "sub_2",
				"status":       "failed",
				"amount_due":   float64(2000),
				"amount_paid":  float64(0),
				"currency":     "usd",
				"number":       "F-1002",
			},
		},
	}

	if err := h.handleInvoicePaymentFailed(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------- Webhook test helpers ----------------------------

// user-not-found DB: always returns ErrNoRows on QueryRow to exercise log-and-continue paths.
type userNotFoundDB struct{}

func (u *userNotFoundDB) Close()                {}
func (u *userNotFoundDB) Pool() storage.PgxPool { return nil }
func (u *userNotFoundDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return errRow{}
}
func (u *userNotFoundDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (u *userNotFoundDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func Test_handleCheckoutSessionCompleted_UserNotFound_OK(t *testing.T) {
	h := NewDefaultHandler(&userNotFoundDB{})
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"id":                  "cs_unf",
				"customer":            "cus_unf",
				"payment_status":      "paid",
				"client_reference_id": "auth0|missing",
			},
		},
	}
	if err := h.handleCheckoutSessionCompleted(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_handleSubscriptionCreated_UserNotFound_OK(t *testing.T) {
	h := NewDefaultHandler(&userNotFoundDB{})
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"customer": "cus_unf",
				"id":       "sub_unf",
				"status":   "active",
			},
		},
	}
	if err := h.handleSubscriptionCreated(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_handleInvoicePaymentSucceeded_UserNotFound_OK(t *testing.T) {
	h := NewDefaultHandler(&userNotFoundDB{})
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"id":           "in_unf",
				"customer":     "cus_unf",
				"subscription": "sub_unf",
				"status":       "paid",
				"amount_due":   float64(100),
				"amount_paid":  float64(100),
				"currency":     "usd",
				"number":       "F-UNF",
			},
		},
	}
	if err := h.handleInvoicePaymentSucceeded(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newEchoCtx(method string, body []byte, headers map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, "/api/v1/stripe/webhook", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func makeEvent(typ string, object map[string]any) []byte {
	body := map[string]any{
		"id":      "evt_test",
		"type":    typ,
		"created": time.Now().Unix(),
		"data": map[string]any{
			"object": object,
		},
	}
	b, _ := json.Marshal(body)
	return b
}

// okRow implements pgx.Row and returns success on Scan (used to simulate existing rows).
type okRow struct{}

func (okRow) Scan(dest ...any) error { return nil }

// errRow implements pgx.Row and returns pgx.ErrNoRows (used to simulate missing rows).
type errRow struct{}

func (errRow) Scan(dest ...any) error { return pgx.ErrNoRows }

// fakeDBAlreadyProcessed makes alreadyProcessedStripeEvent return true.
type fakeDBAlreadyProcessed struct{}

func (f *fakeDBAlreadyProcessed) Close()                {}
func (f *fakeDBAlreadyProcessed) Pool() storage.PgxPool { return nil }
func (f *fakeDBAlreadyProcessed) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return okRow{}
}
func (f *fakeDBAlreadyProcessed) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (f *fakeDBAlreadyProcessed) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

// fakeDBIdemFirstMissingThenUpsertOK simulates IsProcessed=false then Upsert success.
type fakeDBIdemFirstMissingThenUpsertOK struct {
	calls int
}

func (f *fakeDBIdemFirstMissingThenUpsertOK) Close()                {}
func (f *fakeDBIdemFirstMissingThenUpsertOK) Pool() storage.PgxPool { return nil }
func (f *fakeDBIdemFirstMissingThenUpsertOK) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	f.calls++
	if f.calls == 1 {
		// First call for GetProcessedEventByStripeID -> not found
		return errRow{}
	}
	// Second call for UpsertProcessedEventByStripeID -> ok
	return okRow{}
}
func (f *fakeDBIdemFirstMissingThenUpsertOK) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (f *fakeDBIdemFirstMissingThenUpsertOK) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

// ---------------------------- Webhook tests ----------------------------

func TestWebhook_EmptyBody_BadRequest(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)
	c, rec := newEchoCtx(http.MethodPost, []byte{}, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestWebhook_InvalidJSON_BadRequest(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)
	body := []byte("{invalid json")
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestWebhook_UnhandledType_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	body := makeEvent("unhandled.event", map[string]any{"x": 1})
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !contains(rec.Body.String(), "received") {
		t.Fatalf("expected response to contain 'received', got %s", rec.Body.String())
	}
}

func TestWebhook_Signature_MissingHeader_Unauthorized(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("STRIPE_WEBHOOK_SECRET", secret)
	h := NewDefaultHandler(nil)

	body := makeEvent("customer.created", map[string]any{"id": "cus_123"})
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	_ = h.Webhook(c)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if !contains(rec.Body.String(), "unauthorized") {
		t.Fatalf("expected 'unauthorized' in response, got %s", rec.Body.String())
	}
}

func TestWebhook_Signature_Valid_OK(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("STRIPE_WEBHOOK_SECRET", secret)
	h := NewDefaultHandler(nil)

	body := makeEvent("customer.created", map[string]any{"id": "cus_123", "email": "a@b"})
	ts := time.Now().Unix()
	sig := computeStripeSignature(body, ts, secret)
	headers := map[string]string{
		"Stripe-Signature": makeSigHeader(ts, sig),
	}
	c, rec := newEchoCtx(http.MethodPost, body, headers)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !contains(rec.Body.String(), "received") {
		t.Fatalf("expected response to contain 'received', got %s", rec.Body.String())
	}
}

// ---------------------------- Additional edge-case tests ----------------------------

type badReader struct{}

func (badReader) Read(p []byte) (int, error) { return 0, fmt.Errorf("read error") }

func TestWebhook_ReadBodyError_BadRequest(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stripe/webhook", badReader{})
	req.Header.Set(echo.HeaderContentType, "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

type failingRow struct{}

func (failingRow) Scan(dest ...any) error { return fmt.Errorf("boom") }

type fakeDBIdemError struct{}

func (f *fakeDBIdemError) Close() {}

func (f *fakeDBIdemError) Pool() storage.PgxPool { return nil }

func (f *fakeDBIdemError) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return failingRow{}
}

func (f *fakeDBIdemError) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (f *fakeDBIdemError) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func TestWebhook_IdempotencyError_500(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(&fakeDBIdemError{})

	body := makeEvent("unhandled.event", map[string]any{"x": "y"})
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func Test_handleCheckoutSessionCompleted_DB_Branch(t *testing.T) {
	h := NewDefaultHandler(&simpleDB{})
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"id":                  "cs_123",
				"customer":            "cus_link",
				"payment_status":      "paid",
				"client_reference_id": "auth0|user123",
			},
		},
	}
	if err := h.handleCheckoutSessionCompleted(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_handleCheckoutSessionCompleted_InvalidData(t *testing.T) {
	h := NewDefaultHandler(nil)
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": "not_a_map",
		},
	}
	if err := h.handleCheckoutSessionCompleted(context.Background(), &evt); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_handleSubscriptionCreated_InvalidData(t *testing.T) {
	h := NewDefaultHandler(nil)
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": "not_a_map",
		},
	}
	if err := h.handleSubscriptionCreated(context.Background(), &evt); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_handleInvoicePaymentSucceeded_InvalidData(t *testing.T) {
	h := NewDefaultHandler(nil)
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": "not_a_map",
		},
	}
	if err := h.handleInvoicePaymentSucceeded(context.Background(), &evt); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_handleCustomerCreated_InvalidData(t *testing.T) {
	h := NewDefaultHandler(nil)
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": "not_a_map",
		},
	}
	if err := h.handleCustomerCreated(context.Background(), &evt); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_handleCustomerUpdated_InvalidData(t *testing.T) {
	h := NewDefaultHandler(nil)
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": "not_a_map",
		},
	}
	if err := h.handleCustomerUpdated(context.Background(), &evt); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_handleCustomerDeleted_InvalidData(t *testing.T) {
	h := NewDefaultHandler(nil)
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": "not_a_map",
		},
	}
	if err := h.handleCustomerDeleted(context.Background(), &evt); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_handleSubscriptionUpdated_InvalidData(t *testing.T) {
	h := NewDefaultHandler(nil)
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": "not_a_map",
		},
	}
	if err := h.handleSubscriptionUpdated(context.Background(), &evt); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_handleSubscriptionDeleted_InvalidData(t *testing.T) {
	h := NewDefaultHandler(nil)
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": "not_a_map",
		},
	}
	if err := h.handleSubscriptionDeleted(context.Background(), &evt); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_handleInvoicePaymentFailed_InvalidData(t *testing.T) {
	h := NewDefaultHandler(nil)
	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": "not_a_map",
		},
	}
	if err := h.handleInvoicePaymentFailed(context.Background(), &evt); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestWebhook_CheckoutSessionCompleted_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	obj := map[string]any{"id": "cs_123", "customer": "cus_1", "payment_status": "paid", "client_reference_id": "auth0|u1"}
	body := makeEvent("checkout.session.completed", obj)
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_SubscriptionCreated_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	obj := map[string]any{"id": "sub_1", "customer": "cus_1", "status": "active"}
	body := makeEvent("customer.subscription.created", obj)
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_SubscriptionUpdated_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	obj := map[string]any{"id": "sub_2", "customer": "cus_2", "status": "past_due"}
	body := makeEvent("customer.subscription.updated", obj)
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_SubscriptionDeleted_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	obj := map[string]any{"id": "sub_3", "customer": "cus_3"}
	body := makeEvent("customer.subscription.deleted", obj)
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_InvoicePaymentSucceeded_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	obj := map[string]any{"id": "in_1", "customer": "cus_1", "subscription": "sub_1", "status": "paid", "amount_due": float64(1000), "amount_paid": float64(1000), "currency": "usd", "number": "F-1001"}
	body := makeEvent("invoice.payment_succeeded", obj)
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_InvoicePaymentFailed_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	obj := map[string]any{"id": "in_2", "customer": "cus_2", "subscription": "sub_2", "status": "failed", "amount_due": float64(2000), "amount_paid": float64(0), "currency": "usd", "number": "F-1002"}
	body := makeEvent("invoice.payment_failed", obj)
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_CustomerCreated_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	obj := map[string]any{"id": "cus_x", "email": "x@y"}
	body := makeEvent("customer.created", obj)
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_CustomerUpdated_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	obj := map[string]any{"id": "cus_y", "email": "y@z"}
	body := makeEvent("customer.updated", obj)
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_CustomerDeleted_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(nil)

	obj := map[string]any{"id": "cus_z"}
	body := makeEvent("customer.deleted", obj)
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_Idempotent_Duplicate(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(&fakeDBAlreadyProcessed{})

	body := makeEvent("customer.created", map[string]any{"id": "cus_dup"})
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !contains(rec.Body.String(), "duplicate") {
		t.Fatalf("expected response to contain 'duplicate', got %s", rec.Body.String())
	}
}

func TestWebhook_MarkProcessed_DB_Success(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(&fakeDBIdemFirstMissingThenUpsertOK{})

	body := makeEvent("unhandled.event", map[string]any{"x": "y"})
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !contains(rec.Body.String(), "received") {
		t.Fatalf("expected response to contain 'received', got %s", rec.Body.String())
	}
}

// Ensure we actually attempt Upsert (IsProcessed + Upsert -> 2 QueryRow calls)
func TestWebhook_MarkProcessed_Upsert_CalledTwice(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	f := &fakeDBIdemFirstMissingThenUpsertOK{}
	h := NewDefaultHandler(f)

	body := makeEvent("unhandled.event", map[string]any{"ok": true})
	c, _ := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.calls != 2 {
		t.Fatalf("expected 2 QueryRow calls (IsProcessed + Upsert), got %d", f.calls)
	}
}

// fakeDBIdemUpsertError simulates IsProcessed=false then Upsert failing (mark processed error)
// Webhook should still acknowledge with 200.
type fakeDBIdemUpsertError struct {
	calls int
}

func (f *fakeDBIdemUpsertError) Close()                {}
func (f *fakeDBIdemUpsertError) Pool() storage.PgxPool { return nil }
func (f *fakeDBIdemUpsertError) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	f.calls++
	if f.calls == 1 {
		// IsProcessed -> not found
		return errRow{}
	}
	// Upsert -> error on scan
	return failingRow{}
}
func (f *fakeDBIdemUpsertError) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (f *fakeDBIdemUpsertError) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func TestWebhook_MarkProcessed_DB_UpsertError_OK(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	h := NewDefaultHandler(&fakeDBIdemUpsertError{})

	body := makeEvent("unhandled.event", map[string]any{"x": "y"})
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	if err := h.Webhook(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 even if mark-processed fails, got %d", rec.Code)
	}
	if !contains(rec.Body.String(), "received") {
		t.Fatalf("expected response to contain 'received', got %s", rec.Body.String())
	}
}

// Enforce STRIPE_WEBHOOK_SECRET in non-dev envs (simulate non-dev by setting envs and tweaking os.Args[0])
func TestWebhook_MissingSecret_NonDev_ServiceUnavailable(t *testing.T) {
	// Force non-dev
	t.Setenv("APP_ENV", "production")
	t.Setenv("GO_ENV", "production")
	t.Setenv("ENV", "production")
	t.Setenv("NODE_ENV", "production")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")

	orig := os.Args[0]
	os.Args[0] = "api-server" // avoid ".test" heuristic
	defer func() { os.Args[0] = orig }()

	h := NewDefaultHandler(nil)
	body := makeEvent("customer.created", map[string]any{"id": "cus_nondev"})
	c, rec := newEchoCtx(http.MethodPost, body, nil)

	_ = h.Webhook(c)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when secret missing in non-dev, got %d", rec.Code)
	}
	if !contains(rec.Body.String(), "secret not configured") {
		t.Fatalf("expected response to mention 'secret not configured', got %s", rec.Body.String())
	}
}

// Mapping edge-case: ensure subscription handler tolerates nested price/timestamps presence
// even when user lookup fails (no-rows), exercising mapping paths.
func Test_handleSubscriptionCreated_Mapping_PriceAndTimes_OK(t *testing.T) {
	h := NewDefaultHandler(&userNotFoundDB{})
	now := float64(time.Now().Unix())

	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"customer": "cus_map",
				"id":       "sub_map",
				"status":   "active",
				"items": map[string]any{
					"data": []any{
						map[string]any{
							"price": map[string]any{"id": "price_123"},
						},
					},
				},
				"current_period_start": now - 3600,
				"current_period_end":   now + 3600,
				"cancel_at":            now + 7200,
				"canceled_at":          now - 7200,
				"cancel_at_period_end": true,
			},
		},
	}

	if err := h.handleSubscriptionCreated(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Mapping edge-case: invoice with partial data (no currency/number) should still process OK
func Test_handleInvoicePaymentSucceeded_PartialData_OK(t *testing.T) {
	h := NewDefaultHandler(&userNotFoundDB{})

	evt := StripeEvent{
		Data: map[string]interface{}{
			"object": map[string]interface{}{
				"id":           "in_partial",
				"customer":     "cus_partial",
				"subscription": "sub_partial",
				"status":       "paid",
				"amount_due":   float64(1234),
				"amount_paid":  float64(1234),
				// currency, number omitted intentionally
			},
		},
	}

	if err := h.handleInvoicePaymentSucceeded(context.Background(), &evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
