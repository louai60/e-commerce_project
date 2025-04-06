-- Drop the index first
DROP INDEX IF EXISTS idx_users_refresh_token_id;

-- Remove the refresh_token_id column
ALTER TABLE users DROP COLUMN IF EXISTS refresh_token_id;
