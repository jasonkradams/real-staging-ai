-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_users_updated_at();

-- Drop index
DROP INDEX IF EXISTS idx_users_email;

-- Remove added columns
ALTER TABLE users
  DROP COLUMN IF EXISTS updated_at,
  DROP COLUMN IF EXISTS preferences,
  DROP COLUMN IF EXISTS profile_photo_url,
  DROP COLUMN IF EXISTS billing_address,
  DROP COLUMN IF EXISTS phone,
  DROP COLUMN IF EXISTS company_name,
  DROP COLUMN IF EXISTS full_name,
  DROP COLUMN IF EXISTS email;
