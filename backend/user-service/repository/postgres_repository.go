package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/louai60/e-commerce_project/backend/user-service/config"
	"github.com/louai60/e-commerce_project/backend/user-service/models"
)

type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository with connection pooling
func NewPostgresRepository(cfg *config.DatabaseConfig) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Run migrations
	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("error running migrations: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// Close releases database resources
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// WithTransaction executes a function within a database transaction
func (r *PostgresRepository) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// GetUser retrieves a single user by ID
func (r *PostgresRepository) GetUser(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, username, password, first_name, last_name, role, 
			   created_at, updated_at, last_login, is_active, is_verified
		FROM users 
		WHERE id = $1 AND is_active = true`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.Password,
		&user.FirstName, &user.LastName, &user.Role,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		&user.IsActive, &user.IsVerified,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a single user by email
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, username, password, first_name, last_name, role, 
			   created_at, updated_at, last_login, is_active, is_verified
		FROM users 
		WHERE email = $1 AND is_active = true`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.Password,
		&user.FirstName, &user.LastName, &user.Role,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		&user.IsActive, &user.IsVerified,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user by email: %w", err)
	}

	return user, nil
}

// ListUsers retrieves a paginated list of users
func (r *PostgresRepository) ListUsers(ctx context.Context, page, limit int32) ([]*models.User, int64, error) {
	offset := (page - 1) * limit

	countQuery := `SELECT COUNT(*) FROM users WHERE is_active = true`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error counting users: %w", err)
	}

	query := `
		SELECT id, email, username, password, first_name, last_name, role, 
			   created_at, updated_at, last_login, is_active, is_verified
		FROM users 
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.Password,
			&user.FirstName, &user.LastName, &user.Role,
			&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
			&user.IsActive, &user.IsVerified,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error after row iteration: %w", err)
	}

	return users, total, nil
}

// CreateUser inserts a new user into the database
func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, username, password, first_name, last_name, role,
			created_at, updated_at, is_active, is_verified
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.Password,
		user.FirstName, user.LastName, user.Role,
		user.CreatedAt, user.UpdatedAt, user.IsActive, user.IsVerified,
	)

	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user's information
func (r *PostgresRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = $1, username = $2, first_name = $3, last_name = $4,
			role = $5, updated_at = $6
		WHERE id = $7 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query,
		user.Email, user.Username, user.FirstName, user.LastName,
		user.Role, time.Now(), user.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("user not found or not active")
	}

	return nil
}

// DeleteUser soft-deletes a user by marking as inactive
func (r *PostgresRepository) DeleteUser(ctx context.Context, id string) error {
	query := `UPDATE users SET is_active = false, updated_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Ping verifies a connection to the database is still alive
func (r *PostgresRepository) Ping(ctx context.Context) error {
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}
