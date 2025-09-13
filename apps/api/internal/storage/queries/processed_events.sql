-- Processed Events (Stripe Idempotency)
-- sqlc queries for processed_events and subscriptions

-- name: GetProcessedEventByStripeID :one
SELECT
  id,
  stripe_event_id,
  type,
  payload,
  received_at
FROM processed_events
WHERE stripe_event_id = $1;

-- name: CreateProcessedEvent :one
INSERT INTO processed_events (stripe_event_id, type, payload)
VALUES ($1, $2, $3)
RETURNING
  id,
  stripe_event_id,
  type,
  payload,
  received_at;

-- Optional: single-statement upsert that returns the existing/new row.
-- Preserves existing values (no-op update) to obtain RETURNING without DO NOTHING.
-- name: UpsertProcessedEventByStripeID :one
INSERT INTO processed_events (stripe_event_id, type, payload)
VALUES ($1, $2, $3)
ON CONFLICT (stripe_event_id) DO UPDATE
SET type = processed_events.type
RETURNING
  id,
  stripe_event_id,
  type,
  payload,
  received_at;

-- Optional maintenance: delete older processed events by timestamp (retention)
-- name: DeleteOldProcessedEvents :exec
DELETE FROM processed_events
WHERE received_at < $1;



-- Subscriptions (Stripe subscription state)

-- Upsert by unique stripe_subscription_id. We do not modify user_id on conflict.
-- name: UpsertSubscriptionByStripeID :one
INSERT INTO subscriptions (
  user_id,
  stripe_subscription_id,
  status,
  price_id,
  current_period_start,
  current_period_end,
  cancel_at,
  canceled_at,
  cancel_at_period_end
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (stripe_subscription_id) DO UPDATE SET
  status = EXCLUDED.status,
  price_id = EXCLUDED.price_id,
  current_period_start = EXCLUDED.current_period_start,
  current_period_end = EXCLUDED.current_period_end,
  cancel_at = EXCLUDED.cancel_at,
  canceled_at = EXCLUDED.canceled_at,
  cancel_at_period_end = EXCLUDED.cancel_at_period_end,
  updated_at = now()
RETURNING
  id,
  user_id,
  stripe_subscription_id,
  status,
  price_id,
  current_period_start,
  current_period_end,
  cancel_at,
  canceled_at,
  cancel_at_period_end,
  created_at,
  updated_at;

-- name: GetSubscriptionByStripeID :one
SELECT
  id,
  user_id,
  stripe_subscription_id,
  status,
  price_id,
  current_period_start,
  current_period_end,
  cancel_at,
  canceled_at,
  cancel_at_period_end,
  created_at,
  updated_at
FROM subscriptions
WHERE stripe_subscription_id = $1;

-- name: ListSubscriptionsByUserID :many
SELECT
  id,
  user_id,
  stripe_subscription_id,
  status,
  price_id,
  current_period_start,
  current_period_end,
  cancel_at,
  canceled_at,
  cancel_at_period_end,
  created_at,
  updated_at
FROM subscriptions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteSubscriptionByStripeID :exec
DELETE FROM subscriptions
WHERE stripe_subscription_id = $1;
