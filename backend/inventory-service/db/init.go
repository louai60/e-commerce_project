package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/louai60/e-commerce_project/backend/inventory-service/config"
)

// InitDatabase initializes the database and runs migrations
func InitDatabase(cfg *config.Config, logger *zap.Logger) (*sql.DB, error) {
	// First, connect to postgres to check if our database exists
	pgDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password,
	)

	pgDB, err := sql.Open("postgres", pgDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pgDB.Close()

	// Check if our database exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = pgDB.QueryRow(query, cfg.Database.Name).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		logger.Info("Creating database", zap.String("name", cfg.Database.Name))
		_, err = pgDB.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Database.Name))
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
		logger.Info("Database created successfully", zap.String("name", cfg.Database.Name))
	} else {
		logger.Info("Database already exists", zap.String("name", cfg.Database.Name))
	}

	// Connect to our database
	dbDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name,
	)

	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	err = runMigrations(db, logger)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)

	return db, nil
}

// runMigrations runs all SQL migration files in the migrations directory
func runMigrations(db *sql.DB, logger *zap.Logger) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations[version] = true
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating migrations: %w", err)
	}

	// Get list of migration files
	migrationsDir := "migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Sort migration files by version
	sort.Strings(migrationFiles)

	// Apply migrations
	for _, file := range migrationFiles {
		// Extract version from filename (e.g., 000001_init_schema.up.sql -> 000001)
		parts := strings.Split(file, "_")
		if len(parts) < 2 {
			logger.Warn("Invalid migration filename", zap.String("file", file))
			continue
		}
		version := parts[0]

		// Skip if already applied
		if appliedMigrations[version] {
			logger.Info("Migration already applied", zap.String("version", version))
			continue
		}

		// Read migration file
		filePath := filepath.Join(migrationsDir, file)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Begin transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Execute migration
		logger.Info("Applying migration", zap.String("version", version), zap.String("file", file))
		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		// Record migration
		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		logger.Info("Migration applied successfully", zap.String("version", version))
	}

	return nil
}
