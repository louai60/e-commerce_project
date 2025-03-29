CREATE TYPE user_type AS ENUM ('customer', 'seller', 'admin');
CREATE TYPE user_role AS ENUM (
    'guest', 'registered', 'premium',
    'basic_seller', 'verified_seller',
    'support_agent', 'warehouse_staff', 'super_admin'
);

CREATE TABLE IF NOT EXISTS users (
    user_id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    phone_number VARCHAR(20),
    user_type user_type NOT NULL,
    role user_role NOT NULL,
    account_status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    
    CONSTRAINT valid_user_role CHECK (
        (user_type = 'customer' AND role IN ('guest', 'registered', 'premium')) OR
        (user_type = 'seller' AND role IN ('basic_seller', 'verified_seller')) OR
        (user_type = 'admin' AND role IN ('support_agent', 'warehouse_staff', 'super_admin'))
    )
);

CREATE INDEX idx_users_user_type ON users(user_type);
CREATE INDEX idx_users_role ON users(role);

-- User addresses table
CREATE TABLE IF NOT EXISTS user_addresses (
    address_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    address_type VARCHAR(20) NOT NULL,
    street_address1 VARCHAR(100) NOT NULL,
    street_address2 VARCHAR(100),
    city VARCHAR(50) NOT NULL,
    state VARCHAR(50) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(50) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Payment methods table
CREATE TABLE IF NOT EXISTS payment_methods (
    payment_method_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    payment_type VARCHAR(20) NOT NULL,
    card_last_four CHAR(4),
    card_brand VARCHAR(20),
    expiration_month SMALLINT,
    expiration_year SMALLINT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    billing_address_id BIGINT,
    token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (billing_address_id) REFERENCES user_addresses(address_id) ON DELETE SET NULL
);

-- User preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    preference_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    language VARCHAR(10) NOT NULL DEFAULT 'en-US',
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    marketing_emails BOOLEAN NOT NULL DEFAULT TRUE,
    sms_notifications BOOLEAN NOT NULL DEFAULT FALSE,
    dark_mode_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_user_addresses_user_id ON user_addresses(user_id);
CREATE INDEX idx_payment_methods_user_id ON payment_methods(user_id);

