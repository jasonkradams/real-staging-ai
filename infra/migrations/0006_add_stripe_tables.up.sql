-- Add Stripe-related tables: processed_events (idempotency) and subscriptions (user subscription state)

-- Processed events for idempotency of Stripe webhooks.
-- Ensures each Stripe event is processed at most once.
CREATE TABLE IF NOT EXISTS processed_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  stripe_event_id TEXT NOT NULL,
  type TEXT,
  payload JSONB,
  received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_processed_events_stripe_event_id UNIQUE (stripe_event_id)
);

-- Helpful index to query by time window if needed
CREATE INDEX IF NOT EXISTS idx_processed_events_received_at ON processed_events (received_at DESC);

-- Subscriptions table to track Stripe subscription state per user.
CREATE TABLE IF NOT EXISTS subscriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  stripe_subscription_id TEXT NOT NULL,
  status TEXT NOT NULL, -- e.g., active, trialing, past_due, canceled, incomplete, unpaid, etc.
  price_id TEXT,        -- Stripe price/plan identifier (optional but useful)
  current_period_start TIMESTAMPTZ,
  current_period_end TIMESTAMPTZ,
  cancel_at TIMESTAMPTZ,
  canceled_at TIMESTAMPTZ,
  cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_subscriptions_stripe_subscription_id UNIQUE (stripe_subscription_id)
);

-- Indexes to speed up common queries
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions (status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_status ON subscriptions (user_id, status);
