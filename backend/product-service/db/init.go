package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/louai60/e-commerce_project/backend/product-service/config"
)

// InitDatabase initializes the database and runs migrations
func InitDatabase(cfg *config.Config, logger *zap.Logger) (*DBConfig, error) {
	// First, connect to postgres to check if our database exists
	pgDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Secrets.DatabasePassword, cfg.Database.SSLMode,
	)

	logger.Info("Connecting to postgres to check if database exists")
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

	// Now create the DBConfig with master and replicas
	dbConfig, err := NewDBConfig(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database configuration: %w", err)
	}

	// Run migrations
	if err := runMigrations(dbConfig.Master, logger); err != nil {
		dbConfig.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return dbConfig, nil
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
	files, err := os.ReadDir(migrationsDir)
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
		content, err := os.ReadFile(filePath)
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

// GetTempDBConnection creates a temporary connection to the database for fixing migrations
func GetTempDBConnection(cfg *config.Config, logger *zap.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Secrets.DatabasePassword, cfg.Database.Name, cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// ForceFixMigration fixes a dirty migration by marking it as applied
func ForceFixMigration(db *sql.DB, version string, logger *zap.Logger) error {
	// First, create the schema_migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Check if the migration is already applied
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if migration exists: %w", err)
	}

	if exists {
		logger.Info("Migration already exists, removing it first", zap.String("version", version))
		_, err = db.Exec("DELETE FROM schema_migrations WHERE version = $1", version)
		if err != nil {
			return fmt.Errorf("failed to remove existing migration: %w", err)
		}
	}

	// Insert the migration as applied
	_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
	if err != nil {
		return fmt.Errorf("failed to force-apply migration: %w", err)
	}

	logger.Info("Migration force-fixed successfully", zap.String("version", version))
	return nil
}
