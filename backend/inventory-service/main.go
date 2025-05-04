package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/louai60/e-commerce_project/backend/common/logger"
	"github.com/louai60/e-commerce_project/backend/inventory-service/config"
	"github.com/louai60/e-commerce_project/backend/inventory-service/handlers"
	"github.com/louai60/e-commerce_project/backend/inventory-service/middleware"
	pb "github.com/louai60/e-commerce_project/backend/inventory-service/proto"
	"github.com/louai60/e-commerce_project/backend/inventory-service/repository/postgres"
	"github.com/louai60/e-commerce_project/backend/inventory-service/service"
)

func main() {
	// Initialize logger
	logger := initLogger()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Connect to database
	db, err := connectToDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	inventoryRepo := postgres.NewInventoryRepository(db, logger)
	warehouseRepo := postgres.NewWarehouseRepository(db, logger)

	// Initialize services
	inventoryService := service.NewInventoryService(inventoryRepo, warehouseRepo, logger)
	warehouseService := service.NewWarehouseService(warehouseRepo, logger)

	// Initialize gRPC handler
	inventoryHandler := handlers.NewInventoryHandler(inventoryService, warehouseService, logger)

	// Start gRPC server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.LoggingInterceptor(logger)),
	)
	pb.RegisterInventoryServiceServer(server, inventoryHandler)
	reflection.Register(server)

	// Start listening
	port := cfg.Server.Port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err), zap.String("port", port))
	}

	// Handle graceful shutdown
	go func() {
		logger.Info("Starting inventory service", zap.String("port", port))
		if err := server.Serve(lis); err != nil {
			logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down inventory service...")
	server.GracefulStop()
	logger.Info("Inventory service stopped")
}

func initLogger() *zap.Logger {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	logger.Initialize(env)
	return logger.GetLogger()
}

func connectToDatabase(cfg *config.Config, logger *zap.Logger) (*sql.DB, error) {
	dbConfig := cfg.Database

	// First, connect to postgres to check if our database exists
	pgDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password,
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
	err = pgDB.QueryRow(query, dbConfig.Name).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		logger.Info("Creating database", zap.String("name", dbConfig.Name))
		_, err = pgDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbConfig.Name))
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
		logger.Info("Database created successfully", zap.String("name", dbConfig.Name))
	} else {
		logger.Info("Database already exists", zap.String("name", dbConfig.Name))
	}

	// Connect to our database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Name,
	)

	// Try to connect with retries
	var db *sql.DB
	maxRetries := 5
	retryInterval := time.Second * 3

	for i := 0; i < maxRetries; i++ {
		logger.Info("Attempting to connect to database", zap.Int("attempt", i+1))
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			logger.Error("Failed to open database connection", zap.Error(err))
			time.Sleep(retryInterval)
			continue
		}

		// Test the connection
		err = db.Ping()
		if err == nil {
			logger.Info("Successfully connected to database")
			break
		}

		logger.Error("Failed to ping database", zap.Error(err))
		db.Close()
		time.Sleep(retryInterval)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeMinutes) * time.Minute)

	// Run migrations
	if err := runMigrations(db, logger); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify database connection: %w", err)
	}

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
