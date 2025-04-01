-- First drop existing types if they exist
DROP TYPE IF EXISTS user_type CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;

-- Recreate the ENUM types
CREATE TYPE user_type AS ENUM ('customer', 'seller', 'admin');
CREATE TYPE user_role AS ENUM (
    'user', 'admin', 'super_admin',
    'basic_seller', 'verified_seller',
    'support_agent', 'warehouse_staff'
);

-- Create a function to generate random numbers within a range
CREATE OR REPLACE FUNCTION generate_random_id() 
RETURNS bigint AS $$
DECLARE
    random_id bigint;
BEGIN
    -- Generate a random number between 100000 and 999999999
    random_id := floor(random() * (999999999 - 100000 + 1) + 100000);
    RETURN random_id;
END;
$$ LANGUAGE plpgsql;

-- Create sequence with a random start
CREATE SEQUENCE IF NOT EXISTS users_id_seq
    START WITH 100000
    INCREMENT BY 1
    NO MAXVALUE;

-- Create the users table
CREATE TABLE IF NOT EXISTS users (
    user_id bigint PRIMARY KEY DEFAULT generate_random_id(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
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
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE
);

-- Create a trigger to ensure unique random IDs
CREATE OR REPLACE FUNCTION ensure_unique_user_id()
RETURNS TRIGGER AS $$
BEGIN
    WHILE EXISTS (SELECT 1 FROM users WHERE user_id = NEW.user_id) LOOP
        NEW.user_id := generate_random_id();
    END LOOP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ensure_unique_user_id
    BEFORE INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION ensure_unique_user_id();

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

-- Modify the users table
ALTER TABLE users 
    ALTER COLUMN user_type TYPE user_type USING user_type::user_type,
    ALTER COLUMN role TYPE user_role USING role::user_role;

