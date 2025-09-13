package stripe

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

/* ----------------------------- test utilities ----------------------------- */

type fakeDB struct {
	row          pgx.Row
	execErr      error
	execCalled   bool
	lastExecSQL  string
	lastExecArgs []interface{}
}

func (f *fakeDB) Close() {}

func (f *fakeDB) Pool() storage.PgxPool { return nil }

func (f *fakeDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return f.row
}

func (f *fakeDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (f *fakeDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	f.execCalled = true
	f.lastExecSQL = sql
	f.lastExecArgs = args
	return pgconn.CommandTag{}, f.execErr
}

func toUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

func toText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func toTs(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

/* --------------------- stubs for pgx.Row scanning behavior --------------------- */

// processedEventRowStub stubs scanning a ProcessedEvent row.
type processedEventRowStub struct {
	pe  queries.ProcessedEvent
	err error
}

func (r *processedEventRowStub) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	// Expected order:
	// id, stripe_event_id, type, payload, received_at
	if len(dest) < 5 {
		return fmt.Errorf("unexpected dest len: %d", len(dest))
	}
	*(dest[0].(*pgtype.UUID)) = r.pe.ID
	*(dest[1].(*string)) = r.pe.StripeEventID
	*(dest[2].(*pgtype.Text)) = r.pe.Type
	*(dest[3].(*[]byte)) = r.pe.Payload
	*(dest[4].(*pgtype.Timestamptz)) = r.pe.ReceivedAt
	return nil
}

// subscriptionRowStub stubs scanning a Subscription row.
type subscriptionRowStub struct {
	sub queries.Subscription
	err error
}

func (r *subscriptionRowStub) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	// Expected order:
	// id, user_id, stripe_subscription_id, status, price_id,
	// current_period_start, current_period_end, cancel_at, canceled_at,
	// cancel_at_period_end, created_at, updated_at
	if len(dest) < 12 {
		return fmt.Errorf("unexpected dest len: %d", len(dest))
	}
	*(dest[0].(*pgtype.UUID)) = r.sub.ID
	*(dest[1].(*pgtype.UUID)) = r.sub.UserID
	*(dest[2].(*string)) = r.sub.StripeSubscriptionID
	*(dest[3].(*string)) = r.sub.Status
	*(dest[4].(*pgtype.Text)) = r.sub.PriceID
	*(dest[5].(*pgtype.Timestamptz)) = r.sub.CurrentPeriodStart
	*(dest[6].(*pgtype.Timestamptz)) = r.sub.CurrentPeriodEnd
	*(dest[7].(*pgtype.Timestamptz)) = r.sub.CancelAt
	*(dest[8].(*pgtype.Timestamptz)) = r.sub.CanceledAt
	*(dest[9].(*bool)) = r.sub.CancelAtPeriodEnd
	*(dest[10].(*pgtype.Timestamptz)) = r.sub.CreatedAt
	*(dest[11].(*pgtype.Timestamptz)) = r.sub.UpdatedAt
	return nil
}

/* --------------------- ProcessedEventsRepository tests --------------------- */

func TestProcessedEventsRepository_IsProcessed_True(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1_700_000_000, 0)

	expected := queries.ProcessedEvent{
		ID:            toUUID(uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")),
		StripeEventID: "evt_123",
		Type:          toText("invoice.payment_succeeded"),
		Payload:       []byte(`{"ok":true}`),
		ReceivedAt:    toTs(now),
	}
	db := &fakeDB{row: &processedEventRowStub{pe: expected}}

	repo := NewProcessedEventsRepository(db)
	ok, err := repo.IsProcessed(ctx, "evt_123")
	if err != nil {
		t.Fatalf("IsProcessed error: %v", err)
	}
	if !ok {
		t.Fatalf("IsProcessed = false, want true")
	}
}

func TestProcessedEventsRepository_IsProcessed_False(t *testing.T) {
	ctx := context.Background()
	db := &fakeDB{row: &processedEventRowStub{err: pgx.ErrNoRows}}

	repo := NewProcessedEventsRepository(db)
	ok, err := repo.IsProcessed(ctx, "evt_missing")
	if err != nil {
		t.Fatalf("IsProcessed error: %v", err)
	}
	if ok {
		t.Fatalf("IsProcessed = true, want false")
	}
}

func TestProcessedEventsRepository_Upsert_And_Get_Success(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1_700_000_100, 0)

	expected := queries.ProcessedEvent{
		ID:            toUUID(uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")),
		StripeEventID: "evt_456",
		Type:          toText("customer.subscription.created"),
		Payload:       []byte(`{"id":"evt_456"}`),
		ReceivedAt:    toTs(now),
	}

	// Upsert path
	dbUpsert := &fakeDB{row: &processedEventRowStub{pe: expected}}
	repoUpsert := NewProcessedEventsRepository(dbUpsert)

	etype := "customer.subscription.created"
	got, err := repoUpsert.Upsert(ctx, expected.StripeEventID, &etype, expected.Payload)
	if err != nil {
		t.Fatalf("Upsert error: %v", err)
	}
	if got == nil || got.StripeEventID != expected.StripeEventID {
		t.Fatalf("Upsert returned wrong event: got %+v", got)
	}
	if !got.Type.Valid || got.Type.String != etype {
		t.Fatalf("Upsert Type mismatch: got %v want %v", got.Type, etype)
	}

	// Get path
	dbGet := &fakeDB{row: &processedEventRowStub{pe: expected}}
	repoGet := NewProcessedEventsRepository(dbGet)

	got2, err := repoGet.Get(ctx, "evt_456")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got2 == nil || got2.StripeEventID != "evt_456" {
		t.Fatalf("Get returned wrong event: got %+v", got2)
	}
}

func TestProcessedEventsRepository_DeleteOlderThan_Exec(t *testing.T) {
	ctx := context.Background()
	db := &fakeDB{}
	repo := NewProcessedEventsRepository(db)

	before := time.Unix(1_600_000_000, 0)
	if err := repo.DeleteOlderThan(ctx, before); err != nil {
		t.Fatalf("DeleteOlderThan error: %v", err)
	}
	if !db.execCalled {
		t.Fatalf("expected Exec to be called")
	}
}

/* ----------------------- SubscriptionsRepository tests ---------------------- */

func TestSubscriptionsRepository_UpsertByStripeID_Success(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1_700_000_200, 0)

	userID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	subID := "sub_123"
	status := "active"
	priceID := "price_abc"
	cps := now.Add(-24 * time.Hour)
	cpe := now.Add(24 * time.Hour)
	cancelAt := now.Add(48 * time.Hour)
	canceledAt := now.Add(-48 * time.Hour)

	expected := queries.Subscription{
		ID:                   toUUID(uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")),
		UserID:               toUUID(userID),
		StripeSubscriptionID: subID,
		Status:               status,
		PriceID:              toText(priceID),
		CurrentPeriodStart:   toTs(cps),
		CurrentPeriodEnd:     toTs(cpe),
		CancelAt:             toTs(cancelAt),
		CanceledAt:           toTs(canceledAt),
		CancelAtPeriodEnd:    true,
		CreatedAt:            toTs(now),
		UpdatedAt:            toTs(now),
	}

	db := &fakeDB{row: &subscriptionRowStub{sub: expected}}
	repo := NewSubscriptionsRepository(db)

	got, err := repo.UpsertByStripeID(
		ctx,
		userID.String(),
		subID,
		status,
		&priceID,
		&cps,
		&cpe,
		&cancelAt,
		&canceledAt,
		true,
	)
	if err != nil {
		t.Fatalf("UpsertByStripeID error: %v", err)
	}
	if got == nil {
		t.Fatalf("UpsertByStripeID returned nil")
	}
	if got.StripeSubscriptionID != subID || got.Status != status {
		t.Fatalf("mismatch: got subID=%s status=%s", got.StripeSubscriptionID, got.Status)
	}
	if !got.PriceID.Valid || got.PriceID.String != priceID {
		t.Fatalf("price ID mismatch: got %v want %v", got.PriceID, priceID)
	}
	if !got.CancelAtPeriodEnd {
		t.Fatalf("CancelAtPeriodEnd mismatch: got false want true")
	}
}

func TestSubscriptionsRepository_GetByStripeID_Success(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1_700_000_300, 0)

	expected := queries.Subscription{
		ID:                   toUUID(uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")),
		UserID:               toUUID(uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")),
		StripeSubscriptionID: "sub_789",
		Status:               "past_due",
		PriceID:              toText("price_xyz"),
		CurrentPeriodStart:   toTs(now.Add(-12 * time.Hour)),
		CurrentPeriodEnd:     toTs(now.Add(12 * time.Hour)),
		CancelAt:             pgtype.Timestamptz{}, // null
		CanceledAt:           pgtype.Timestamptz{}, // null
		CancelAtPeriodEnd:    false,
		CreatedAt:            toTs(now),
		UpdatedAt:            toTs(now),
	}

	db := &fakeDB{row: &subscriptionRowStub{sub: expected}}
	repo := NewSubscriptionsRepository(db)

	got, err := repo.GetByStripeID(ctx, "sub_789")
	if err != nil {
		t.Fatalf("GetByStripeID error: %v", err)
	}
	if got == nil || got.StripeSubscriptionID != "sub_789" {
		t.Fatalf("GetByStripeID returned wrong subscription: %+v", got)
	}
	if got.Status != "past_due" {
		t.Fatalf("status mismatch: got %q want %q", got.Status, "past_due")
	}
}

func TestSubscriptionsRepository_UpsertByStripeID_InvalidUserID(t *testing.T) {
	ctx := context.Background()
	db := &fakeDB{row: &subscriptionRowStub{err: fmt.Errorf("should not be called")}}
	repo := NewSubscriptionsRepository(db)

	_, err := repo.UpsertByStripeID(ctx, "not-a-uuid", "sub_1", "active", nil, nil, nil, nil, nil, false)
	if err == nil {
		t.Fatalf("expected error for invalid user ID, got nil")
	}
	if want := "invalid user ID format"; err == nil || err.Error() == "" || !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestSubscriptionsRepository_DeleteByStripeID_Exec(t *testing.T) {
	ctx := context.Background()
	db := &fakeDB{} // Exec returns success by default
	repo := NewSubscriptionsRepository(db)

	if err := repo.DeleteByStripeID(ctx, "sub_to_delete"); err != nil {
		t.Fatalf("DeleteByStripeID error: %v", err)
	}
	if !db.execCalled {
		t.Fatalf("expected Exec to be called")
	}
}

/* ----------------------------- tiny helper ----------------------------- */

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
