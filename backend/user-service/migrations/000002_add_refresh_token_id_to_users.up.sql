-- Add the refresh_token_id column to store the JTI of the current valid refresh token
ALTER TABLE users ADD COLUMN refresh_token_id VARCHAR(255);

-- Add an index for potentially faster lookups based on refresh token ID, though lookups might primarily be by user ID.
-- This might be more useful if implementing a global blacklist later.
CREATE INDEX idx_users_refresh_token_id ON users (refresh_token_id);
