BEGIN;

-- Drop invoice-related indexes
DROP INDEX IF EXISTS idx_invoices_stripe_subscription_id;
DROP INDEX IF EXISTS idx_invoices_created_at;
DROP INDEX IF EXISTS idx_invoices_status;
DROP INDEX IF EXISTS idx_invoices_user_id;

-- Drop invoices table
DROP TABLE IF EXISTS invoices;

COMMIT;
