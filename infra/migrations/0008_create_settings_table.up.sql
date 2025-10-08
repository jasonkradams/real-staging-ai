-- Create settings table to store system-wide configuration
CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_settings_updated_at ON settings(updated_at DESC);

-- Insert default model setting
INSERT INTO settings (key, value, description) 
VALUES (
    'active_model',
    'black-forest-labs/flux-kontext-max',
    'The active AI model used for virtual staging'
) ON CONFLICT (key) DO NOTHING;

-- Add comment to table
COMMENT ON TABLE settings IS 'System-wide configuration settings';
COMMENT ON COLUMN settings.key IS 'Unique setting identifier';
COMMENT ON COLUMN settings.value IS 'Setting value (stored as text, parsed by application)';
COMMENT ON COLUMN settings.description IS 'Human-readable description of the setting';
COMMENT ON COLUMN settings.updated_at IS 'Timestamp of last update';
COMMENT ON COLUMN settings.updated_by IS 'User who last updated the setting';
