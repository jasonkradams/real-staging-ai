package stripe

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

// rowStub implements pgx.Row and feeds a predetermined Invoice back to sqlc scan destinations.
type rowStub struct {
	inv queries.Invoice
	err error
}

func (r *rowStub) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	// Expect sqlc to scan in this exact order (see invoices.sql.go)
	// id, user_id, stripe_invoice_id, stripe_subscription_id, status,
	// amount_due, amount_paid, currency, invoice_number, created_at, updated_at
	if len(dest) < 11 {
		return fmt.Errorf("unexpected dest len: %d", len(dest))
	}

	*(dest[0].(*pgtype.UUID)) = r.inv.ID
	*(dest[1].(*pgtype.UUID)) = r.inv.UserID
	*(dest[2].(*string)) = r.inv.StripeInvoiceID
	*(dest[3].(*pgtype.Text)) = r.inv.StripeSubscriptionID
	*(dest[4].(*string)) = r.inv.Status
	*(dest[5].(*int32)) = r.inv.AmountDue
	*(dest[6].(*int32)) = r.inv.AmountPaid
	*(dest[7].(*pgtype.Text)) = r.inv.Currency
	*(dest[8].(*pgtype.Text)) = r.inv.InvoiceNumber
	*(dest[9].(*pgtype.Timestamptz)) = r.inv.CreatedAt
	*(dest[10].(*pgtype.Timestamptz)) = r.inv.UpdatedAt
	return nil
}

// using shared fakeDB from repository_subs_events_test.go

// using shared helpers from repository_subs_events_test.go

func TestInvoicesRepository_Upsert_Success(t *testing.T) {
	ctx := context.Background()

	userID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	invoiceID := "in_test_1"
	subID := "sub_123"
	status := "paid"
	amountDue := int32(3000)
	amountPaid := int32(3000)
	currency := "usd"
	invoiceNumber := "F-1001"
	now := time.Unix(1_700_000_000, 0)

	expected := queries.Invoice{
		ID:                   toUUID(uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")),
		UserID:               toUUID(userID),
		StripeInvoiceID:      invoiceID,
		StripeSubscriptionID: toText(subID),
		Status:               status,
		AmountDue:            amountDue,
		AmountPaid:           amountPaid,
		Currency:             toText(currency),
		InvoiceNumber:        toText(invoiceNumber),
		CreatedAt:            toTs(now),
		UpdatedAt:            toTs(now),
	}

	db := &fakeDB{row: &rowStub{inv: expected}}
	repo := NewInvoicesRepository(db)

	// Prepare pointer args as repository expects
	subPtr := subID
	curPtr := currency
	numPtr := invoiceNumber

	got, err := repo.Upsert(ctx, userID.String(), invoiceID, &subPtr, status, amountDue, amountPaid, &curPtr, &numPtr)
	if err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}
	if got == nil {
		t.Fatalf("Upsert returned nil invoice")
	}

	if got.StripeInvoiceID != invoiceID {
		t.Fatalf("StripeInvoiceID mismatch: got %q want %q", got.StripeInvoiceID, invoiceID)
	}
	if got.Status != status {
		t.Fatalf("Status mismatch: got %q want %q", got.Status, status)
	}
	if got.AmountPaid != amountPaid || got.AmountDue != amountDue {
		t.Fatalf("Amounts mismatch: got (paid=%d,due=%d) want (paid=%d,due=%d)",
			got.AmountPaid, got.AmountDue, amountPaid, amountDue)
	}
	if !got.UserID.Valid || got.UserID.Bytes != userID {
		t.Fatalf("UserID mismatch: got %v want %v", got.UserID.Bytes, userID)
	}
	if !got.Currency.Valid || got.Currency.String != currency {
		t.Fatalf("Currency mismatch: got %v want %v", got.Currency, currency)
	}
	if !got.InvoiceNumber.Valid || got.InvoiceNumber.String != invoiceNumber {
		t.Fatalf("InvoiceNumber mismatch: got %v want %v", got.InvoiceNumber, invoiceNumber)
	}
}

func TestInvoicesRepository_GetByStripeID_Success(t *testing.T) {
	ctx := context.Background()

	userID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	invoiceID := "in_test_2"
	now := time.Unix(1_700_000_100, 0)

	expected := queries.Invoice{
		ID:                   toUUID(uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")),
		UserID:               toUUID(userID),
		StripeInvoiceID:      invoiceID,
		StripeSubscriptionID: toText("sub_999"),
		Status:               "failed",
		AmountDue:            5000,
		AmountPaid:           0,
		Currency:             toText("usd"),
		InvoiceNumber:        toText("F-1002"),
		CreatedAt:            toTs(now),
		UpdatedAt:            toTs(now),
	}

	db := &fakeDB{row: &rowStub{inv: expected}}
	repo := NewInvoicesRepository(db)

	got, err := repo.GetByStripeID(ctx, invoiceID)
	if err != nil {
		t.Fatalf("GetByStripeID returned error: %v", err)
	}
	if got == nil {
		t.Fatalf("GetByStripeID returned nil invoice")
	}

	if got.StripeInvoiceID != invoiceID {
		t.Fatalf("StripeInvoiceID mismatch: got %q want %q", got.StripeInvoiceID, invoiceID)
	}
	if got.Status != "failed" {
		t.Fatalf("Status mismatch: got %q want %q", got.Status, "failed")
	}
	if got.AmountDue != 5000 || got.AmountPaid != 0 {
		t.Fatalf("Amounts mismatch: got (paid=%d,due=%d) want (paid=%d,due=%d)",
			got.AmountPaid, got.AmountDue, 0, 5000)
	}
}

func TestInvoicesRepository_GetByStripeID_NotFound(t *testing.T) {
	ctx := context.Background()

	// Simulate sqlc scan error translating to pgx.ErrNoRows
	db := &fakeDB{row: &rowStub{err: pgx.ErrNoRows}}
	repo := NewInvoicesRepository(db)

	_, err := repo.GetByStripeID(ctx, "in_missing")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != pgx.ErrNoRows {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}
