package stripe

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

// ProcessedEventsRepository provides idempotency support for Stripe webhooks.
type ProcessedEventsRepository interface {
	// IsProcessed returns true if the given Stripe event ID has already been processed.
	IsProcessed(ctx context.Context, stripeEventID string) (bool, error)
	// Upsert marks the given event as processed; if it already exists, it is a no-op.
	// eventType may be nil. payload should be a JSON-encoded body (may be nil/empty).
	Upsert(ctx context.Context, stripeEventID string, eventType *string, payload []byte) (*queries.ProcessedEvent, error)
	// Get returns the processed event record by Stripe event ID.
	Get(ctx context.Context, stripeEventID string) (*queries.ProcessedEvent, error)
	// DeleteOlderThan removes processed events older than the given timestamp (for retention housekeeping).
	DeleteOlderThan(ctx context.Context, before time.Time) error
}

// SubscriptionsRepository manages user subscription state from Stripe.
type SubscriptionsRepository interface {
	// UpsertByStripeID upserts a subscription by its Stripe subscription ID.
	// cancelAtPeriodEnd is required; other time fields and priceID are optional.
	UpsertByStripeID(
		ctx context.Context,
		userID string,
		stripeSubscriptionID string,
		status string,
		priceID *string,
		currentPeriodStart *time.Time,
		currentPeriodEnd *time.Time,
		cancelAt *time.Time,
		canceledAt *time.Time,
		cancelAtPeriodEnd bool,
	) (*queries.Subscription, error)

	// GetByStripeID returns a subscription by Stripe subscription ID.
	GetByStripeID(ctx context.Context, stripeSubscriptionID string) (*queries.Subscription, error)

	// ListByUserID lists subscriptions for a user with pagination.
	ListByUserID(ctx context.Context, userID string, limit, offset int32) ([]*queries.Subscription, error)

	// DeleteByStripeID deletes a subscription row by Stripe subscription ID.
	DeleteByStripeID(ctx context.Context, stripeSubscriptionID string) error
}

/* ---------------------------- Implementations ---------------------------- */

type processedEventsRepo struct {
	q *queries.Queries
}

type subscriptionsRepo struct {
	q *queries.Queries
}

// NewProcessedEventsRepository returns a sqlc-backed ProcessedEventsRepository.
func NewProcessedEventsRepository(db storage.Database) ProcessedEventsRepository {
	return &processedEventsRepo{q: queries.New(db)}
}

// NewSubscriptionsRepository returns a sqlc-backed SubscriptionsRepository.
func NewSubscriptionsRepository(db storage.Database) SubscriptionsRepository {
	return &subscriptionsRepo{q: queries.New(db)}
}

/* ----------------------- ProcessedEventsRepository ----------------------- */

func (r *processedEventsRepo) IsProcessed(ctx context.Context, stripeEventID string) (bool, error) {
	_, err := r.q.GetProcessedEventByStripeID(ctx, stripeEventID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check processed event: %w", err)
	}
	return true, nil
}

func (r *processedEventsRepo) Upsert(ctx context.Context, stripeEventID string, eventType *string, payload []byte) (*queries.ProcessedEvent, error) {
	// sqlc model: UpsertProcessedEventByStripeID(ctx, $1, $2, $3) :one
	// eventType stored as TEXT (nullable), payload stored as JSONB (nullable).
	var et pgtype.Text
	if eventType != nil {
		et = pgtype.Text{String: *eventType, Valid: true}
	} else {
		et = pgtype.Text{} // NULL
	}

	// For JSONB payload, sqlc typically accepts []byte. If empty, pass nil-like.
	// When using pgtype.JSONB in your sqlc config, you can switch to pgtype.JSONB accordingly.
	var data []byte
	if len(payload) > 0 {
		data = payload
	}

	pe, err := r.q.UpsertProcessedEventByStripeID(ctx, queries.UpsertProcessedEventByStripeIDParams{
		StripeEventID: stripeEventID,
		Type:          et,
		Payload:       data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert processed event: %w", err)
	}
	return pe, nil
}

func (r *processedEventsRepo) Get(ctx context.Context, stripeEventID string) (*queries.ProcessedEvent, error) {
	pe, err := r.q.GetProcessedEventByStripeID(ctx, stripeEventID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get processed event: %w", err)
	}
	return pe, nil
}

func (r *processedEventsRepo) DeleteOlderThan(ctx context.Context, before time.Time) error {
	ts := pgtype.Timestamptz{Time: before, Valid: true}
	if err := r.q.DeleteOldProcessedEvents(ctx, ts); err != nil {
		return fmt.Errorf("failed to delete old processed events: %w", err)
	}
	return nil
}

/* -------------------------- SubscriptionsRepository -------------------------- */

func (r *subscriptionsRepo) UpsertByStripeID(
	ctx context.Context,
	userID string,
	stripeSubscriptionID string,
	status string,
	priceID *string,
	currentPeriodStart *time.Time,
	currentPeriodEnd *time.Time,
	cancelAt *time.Time,
	canceledAt *time.Time,
	cancelAtPeriodEnd bool,
) (*queries.Subscription, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	userUUID := pgtype.UUID{Bytes: uid, Valid: true}

	// Optional fields
	var priceText pgtype.Text
	if priceID != nil {
		priceText = pgtype.Text{String: *priceID, Valid: true}
	}

	var cps, cpe, ca, cz pgtype.Timestamptz
	if currentPeriodStart != nil {
		cps = pgtype.Timestamptz{Time: *currentPeriodStart, Valid: true}
	}
	if currentPeriodEnd != nil {
		cpe = pgtype.Timestamptz{Time: *currentPeriodEnd, Valid: true}
	}
	if cancelAt != nil {
		ca = pgtype.Timestamptz{Time: *cancelAt, Valid: true}
	}
	if canceledAt != nil {
		cz = pgtype.Timestamptz{Time: *canceledAt, Valid: true}
	}

	sub, err := r.q.UpsertSubscriptionByStripeID(ctx, queries.UpsertSubscriptionByStripeIDParams{
		UserID:               userUUID,
		StripeSubscriptionID: stripeSubscriptionID,
		Status:               status,
		PriceID:              priceText,
		CurrentPeriodStart:   cps,
		CurrentPeriodEnd:     cpe,
		CancelAt:             ca,
		CanceledAt:           cz,
		CancelAtPeriodEnd:    cancelAtPeriodEnd,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert subscription: %w", err)
	}

	return sub, nil
}

func (r *subscriptionsRepo) GetByStripeID(ctx context.Context, stripeSubscriptionID string) (*queries.Subscription, error) {
	sub, err := r.q.GetSubscriptionByStripeID(ctx, stripeSubscriptionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get subscription by stripe id: %w", err)
	}
	return sub, nil
}

func (r *subscriptionsRepo) ListByUserID(ctx context.Context, userID string, limit, offset int32) ([]*queries.Subscription, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	userUUID := pgtype.UUID{Bytes: uid, Valid: true}

	results, err := r.q.ListSubscriptionsByUserID(ctx, queries.ListSubscriptionsByUserIDParams{
		UserID: userUUID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions by user: %w", err)
	}
	return results, nil
}

func (r *subscriptionsRepo) DeleteByStripeID(ctx context.Context, stripeSubscriptionID string) error {
	if err := r.q.DeleteSubscriptionByStripeID(ctx, stripeSubscriptionID); err != nil {
		return fmt.Errorf("failed to delete subscription by stripe id: %w", err)
	}
	return nil
}
