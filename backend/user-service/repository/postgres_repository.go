// postgres_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/louai60/e-commerce_project/backend/user-service/models"
)

import "go.uber.org/zap"

type PostgresRepository struct {
	db *sql.DB
	Logger *zap.Logger
}

func NewPostgresRepository(db *sql.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db: db,
		Logger: logger,
	}
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// User operations

func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			username, email, hashed_password, first_name, last_name, 
			phone_number, user_type, role, account_status, 
			email_verified, phone_verified
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING user_id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.HashedPassword,
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
		user.UserType,
		user.Role,
		user.AccountStatus,
		user.EmailVerified,
		user.PhoneVerified,
	).Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				if strings.Contains(pqErr.Constraint, "users_email_key") {
					return fmt.Errorf("email already exists")
				}
				if strings.Contains(pqErr.Constraint, "users_username_key") {
					return fmt.Errorf("username already exists")
				}
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	query := `
		SELECT user_id, username, email, hashed_password, first_name, last_name, 
			   phone_number, user_type, role, account_status, email_verified, 
			   phone_verified, created_at, updated_at, last_login
		FROM users
		WHERE user_id = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.HashedPassword,
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
		&user.UserType,
		&user.Role,
		&user.AccountStatus,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, first_name = $3, last_name = $4,
			phone_number = $5, user_type = $6, role = $7, account_status = $8,
			email_verified = $9, phone_verified = $10, updated_at = $11
		WHERE user_id = $12
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
		user.UserType,
		user.Role,
		user.AccountStatus,
		user.EmailVerified,
		user.PhoneVerified,
		time.Now(),
		user.UserID,
	).Scan(&user.UpdatedAt)
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, userID int64) error {
	query := `DELETE FROM users WHERE user_id = $1`
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	r.Logger.Info("Attempting to get user by email", zap.String("email", email))

	query := `
		SELECT 
			user_id, 
			username, 
			email, 
			hashed_password,  
			first_name, 
			last_name, 
			phone_number, 
			user_type, 
			role, 
			account_status, 
			email_verified, 
			phone_verified, 
			created_at, 
			updated_at, 
			COALESCE(last_login, created_at) as last_login
		FROM users 
		WHERE LOWER(email) = LOWER($1)`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.HashedPassword,  
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
		&user.UserType,
		&user.Role,
		&user.AccountStatus,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.Logger.Error("User not found", zap.String("email", email))
			return nil, fmt.Errorf("user not found")
		}
		r.Logger.Error("Database error", zap.Error(err))
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	r.Logger.Info("User found", 
		zap.String("email", email),
		zap.Int64("userId", user.UserID))
	return user, nil
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT user_id, username, email, hashed_password, first_name, last_name, 
			   phone_number, user_type, role, account_status, email_verified, 
			   phone_verified, created_at, updated_at, last_login
		FROM users
		WHERE username = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.HashedPassword,
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
		&user.UserType,
		&user.Role,
		&user.AccountStatus,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) ListUsers(ctx context.Context, page, limit int, where string, args ...interface{}) ([]*models.User, error) {
	offset := (page - 1) * limit
	query := `
		SELECT user_id, username, email, first_name, last_name, phone_number, 
			   user_type, role, account_status, created_at, updated_at, last_login
		FROM users
	`
	if where != "" {
		query += " " + where
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.PhoneNumber,
			&user.UserType,
			&user.Role,
			&user.AccountStatus,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLogin,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users rows: %w", err)
	}

	return users, nil
}

func (r *PostgresRepository) CountUsers(ctx context.Context, where string, args ...interface{}) (int64, error) {
	query := "SELECT COUNT(*) FROM users"
	if where != "" {
		query += " " + where
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// Address operations

func (r *PostgresRepository) CreateAddress(ctx context.Context, address *models.UserAddress) error {
	query := `
		INSERT INTO user_addresses (user_id, address_type, street_address1, 
								  street_address2, city, state, postal_code, 
								  country, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING address_id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		address.UserID, address.AddressType, address.StreetAddress1,
		address.StreetAddress2, address.City, address.State,
		address.PostalCode, address.Country, address.IsDefault,
	).Scan(&address.AddressID, &address.CreatedAt, &address.UpdatedAt)
}

func (r *PostgresRepository) GetAddresses(ctx context.Context, userID int64) ([]models.UserAddress, error) {
	query := `
		SELECT address_id, user_id, address_type, street_address1, street_address2,
			   city, state, postal_code, country, is_default, created_at, updated_at
		FROM user_addresses
		WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query addresses: %w", err)
	}
	defer rows.Close()

	var addresses []models.UserAddress
	for rows.Next() {
		var address models.UserAddress
		err := rows.Scan(
			&address.AddressID,
			&address.UserID,
			&address.AddressType,
			&address.StreetAddress1,
			&address.StreetAddress2,
			&address.City,
			&address.State,
			&address.PostalCode,
			&address.Country,
			&address.IsDefault,
			&address.CreatedAt,
			&address.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan address: %w", err)
		}
		addresses = append(addresses, address)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating addresses rows: %w", err)
	}

	return addresses, nil
}

func (r *PostgresRepository) UpdateAddress(ctx context.Context, address *models.UserAddress) error {
	query := `
		UPDATE user_addresses
		SET address_type = $1, street_address1 = $2, street_address2 = $3,
			city = $4, state = $5, postal_code = $6, country = $7, is_default = $8,
			updated_at = $9
		WHERE address_id = $10 AND user_id = $11
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		address.AddressType,
		address.StreetAddress1,
		address.StreetAddress2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.IsDefault,
		time.Now(),
		address.AddressID,
		address.UserID,
	).Scan(&address.UpdatedAt)
}

func (r *PostgresRepository) DeleteAddress(ctx context.Context, addressID, userID int64) error {
	query := `DELETE FROM user_addresses WHERE address_id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, addressID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("address not found or not owned by user")
	}

	return nil
}

func (r *PostgresRepository) GetDefaultAddress(ctx context.Context, userID int64) (*models.UserAddress, error) {
	query := `
		SELECT address_id, user_id, address_type, street_address1, street_address2,
			   city, state, postal_code, country, is_default, created_at, updated_at
		FROM user_addresses
		WHERE user_id = $1 AND is_default = true
		LIMIT 1`

	address := &models.UserAddress{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&address.AddressID,
		&address.UserID,
		&address.AddressType,
		&address.StreetAddress1,
		&address.StreetAddress2,
		&address.City,
		&address.State,
		&address.PostalCode,
		&address.Country,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("default address not found")
		}
		return nil, fmt.Errorf("failed to get default address: %w", err)
	}

	return address, nil
}

// Payment method operations

func (r *PostgresRepository) CreatePaymentMethod(ctx context.Context, payment *models.PaymentMethod) error {
	query := `
		INSERT INTO payment_methods (user_id, payment_type, card_last_four, 
								   card_brand, expiration_month, expiration_year,
								   is_default, billing_address_id, token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING payment_method_id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		payment.UserID, payment.PaymentType, payment.CardLastFour,
		payment.CardBrand, payment.ExpirationMonth, payment.ExpirationYear,
		payment.IsDefault, payment.BillingAddressID, payment.Token,
	).Scan(&payment.PaymentMethodID, &payment.CreatedAt, &payment.UpdatedAt)
}

func (r *PostgresRepository) GetPaymentMethods(ctx context.Context, userID int64) ([]models.PaymentMethod, error) {
	query := `
		SELECT payment_method_id, user_id, payment_type, card_last_four, 
			   card_brand, expiration_month, expiration_year, is_default,
			   billing_address_id, token, created_at, updated_at
		FROM payment_methods
		WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment methods: %w", err)
	}
	defer rows.Close()

	var methods []models.PaymentMethod
	for rows.Next() {
		var method models.PaymentMethod
		err := rows.Scan(
			&method.PaymentMethodID,
			&method.UserID,
			&method.PaymentType,
			&method.CardLastFour,
			&method.CardBrand,
			&method.ExpirationMonth,
			&method.ExpirationYear,
			&method.IsDefault,
			&method.BillingAddressID,
			&method.Token,
			&method.CreatedAt,
			&method.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment method: %w", err)
		}
		methods = append(methods, method)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment methods rows: %w", err)
	}

	return methods, nil
}

func (r *PostgresRepository) UpdatePaymentMethod(ctx context.Context, payment *models.PaymentMethod) error {
	query := `
		UPDATE payment_methods
		SET payment_type = $1, card_last_four = $2, card_brand = $3,
			expiration_month = $4, expiration_year = $5, is_default = $6,
			billing_address_id = $7, token = $8, updated_at = $9
		WHERE payment_method_id = $10 AND user_id = $11
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		payment.PaymentType,
		payment.CardLastFour,
		payment.CardBrand,
		payment.ExpirationMonth,
		payment.ExpirationYear,
		payment.IsDefault,
		payment.BillingAddressID,
		payment.Token,
		time.Now(),
		payment.PaymentMethodID,
		payment.UserID,
	).Scan(&payment.UpdatedAt)
}

func (r *PostgresRepository) DeletePaymentMethod(ctx context.Context, paymentID, userID int64) error {
	query := `DELETE FROM payment_methods WHERE payment_method_id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, paymentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment method not found or not owned by user")
	}

	return nil
}

func (r *PostgresRepository) GetDefaultPaymentMethod(ctx context.Context, userID int64) (*models.PaymentMethod, error) {
	query := `
		SELECT payment_method_id, user_id, payment_type, card_last_four, 
			   card_brand, expiration_month, expiration_year, is_default,
			   billing_address_id, token, created_at, updated_at
		FROM payment_methods
		WHERE user_id = $1 AND is_default = true
		LIMIT 1`

	payment := &models.PaymentMethod{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&payment.PaymentMethodID,
		&payment.UserID,
		&payment.PaymentType,
		&payment.CardLastFour,
		&payment.CardBrand,
		&payment.ExpirationMonth,
		&payment.ExpirationYear,
		&payment.IsDefault,
		&payment.BillingAddressID,
		&payment.Token,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("default payment method not found")
		}
		return nil, fmt.Errorf("failed to get default payment method: %w", err)
	}

	return payment, nil
}

// Preferences operations

func (r *PostgresRepository) CreatePreferences(ctx context.Context, prefs *models.UserPreferences) error {
	query := `
		INSERT INTO user_preferences (user_id, language, currency, 
									notification_email, notification_sms, 
									theme, timezone)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		prefs.UserID,
		prefs.Language,
		prefs.Currency,
		prefs.NotificationEmail,
		prefs.NotificationSMS,
		prefs.Theme,
		prefs.Timezone,
	).Scan(&prefs.CreatedAt, &prefs.UpdatedAt)
}

func (r *PostgresRepository) GetPreferences(ctx context.Context, userID int64) (*models.UserPreferences, error) {
	query := `
		SELECT user_id, language, currency, notification_email, 
			   notification_sms, theme, timezone, created_at, updated_at
		FROM user_preferences
		WHERE user_id = $1`

	prefs := &models.UserPreferences{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.UserID,
		&prefs.Language,
		&prefs.Currency,
		&prefs.NotificationEmail,
		&prefs.NotificationSMS,
		&prefs.Theme,
		&prefs.Timezone,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("preferences not found")
		}
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	return prefs, nil
}

func (r *PostgresRepository) UpdatePreferences(ctx context.Context, prefs *models.UserPreferences) error {
	query := `
		UPDATE user_preferences
		SET language = $1, currency = $2, notification_email = $3, 
			notification_sms = $4, theme = $5, timezone = $6, updated_at = $7
		WHERE user_id = $8
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		prefs.Language,
		prefs.Currency,
		prefs.NotificationEmail,
		prefs.NotificationSMS,
		prefs.Theme,
		prefs.Timezone,
		time.Now(),
		prefs.UserID,
	).Scan(&prefs.UpdatedAt)
}

// UpdateRefreshTokenID updates the refresh token ID for a given user.
func (r *PostgresRepository) UpdateRefreshTokenID(ctx context.Context, userID int64, refreshTokenID string) error {
	query := `
		UPDATE users
		SET refresh_token_id = $1, updated_at = $2
		WHERE user_id = $3`

	result, err := r.db.ExecContext(ctx, query, refreshTokenID, time.Now(), userID)
	if err != nil {
		r.Logger.Error("Failed to update refresh token ID", zap.Int64("userID", userID), zap.Error(err))
		return fmt.Errorf("failed to update refresh token ID: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.Logger.Error("Failed to get rows affected after updating refresh token ID", zap.Int64("userID", userID), zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.Logger.Warn("No user found to update refresh token ID", zap.Int64("userID", userID))
		return fmt.Errorf("user not found")
	}

	r.Logger.Info("Successfully updated refresh token ID", zap.Int64("userID", userID), zap.String("refreshTokenID", refreshTokenID))
	return nil
}
