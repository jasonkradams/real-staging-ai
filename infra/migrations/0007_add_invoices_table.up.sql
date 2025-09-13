-- Create Stripe invoices table for persistence and analytics
-- Stores basic invoice information and links to users and (optionally) subscriptions.

-- Invoices table
CREATE TABLE IF NOT EXISTS invoices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  stripe_invoice_id TEXT NOT NULL,
  stripe_subscription_id TEXT, -- optional linkage to subscriptions table (by Stripe ID)
  status TEXT NOT NULL,        -- e.g., draft, open, paid, uncollectible, void
  amount_due INTEGER NOT NULL DEFAULT 0,   -- in cents
  amount_paid INTEGER NOT NULL DEFAULT 0,  -- in cents
  currency TEXT,                           -- e.g., "usd"
  invoice_number TEXT,                     -- Stripe invoice number (human-readable), optional
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_invoices_stripe_invoice_id UNIQUE (stripe_invoice_id)
);

-- Helpful indexes
CREATE INDEX IF NOT EXISTS idx_invoices_user_id ON invoices (user_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices (status);
CREATE INDEX IF NOT EXISTS idx_invoices_created_at ON invoices (created_at DESC);

-- Optional: if you plan to frequently filter by subscription ID
CREATE INDEX IF NOT EXISTS idx_invoices_stripe_subscription_id ON invoices (stripe_subscription_id);
