-- Down migration: drop Stripe-related tables and indexes added in 0006_add_stripe_tables.up.sql

BEGIN;

-- Drop indexes related to subscriptions
DROP INDEX IF EXISTS idx_subscriptions_user_status;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_user_id;

-- Drop subscriptions table
DROP TABLE IF EXISTS subscriptions;

-- Drop indexes related to processed_events
DROP INDEX IF EXISTS idx_processed_events_received_at;

-- Drop processed_events table
DROP TABLE IF EXISTS processed_events;

COMMIT;
