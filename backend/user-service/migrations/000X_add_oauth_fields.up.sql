ALTER TABLE users
ADD COLUMN provider VARCHAR(50),
ADD COLUMN provider_account_id VARCHAR(255),
ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;

CREATE INDEX idx_provider_account ON users(provider, provider_account_id);