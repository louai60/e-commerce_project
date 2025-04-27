package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/louai60/e-commerce_project/backend/user-service/cache"
	"github.com/louai60/e-commerce_project/backend/user-service/config"
	"github.com/louai60/e-commerce_project/backend/user-service/db"
	"github.com/louai60/e-commerce_project/backend/user-service/handlers"
	pb "github.com/louai60/e-commerce_project/backend/user-service/proto"
	"github.com/louai60/e-commerce_project/backend/user-service/repository"
	"github.com/louai60/e-commerce_project/backend/user-service/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func initializeDatabase(ctx context.Context, db *sql.DB, logger *zap.Logger) error {
	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create users table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			user_id BIGSERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			hashed_password TEXT NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			phone_number VARCHAR(20),
			user_type VARCHAR(20) DEFAULT 'customer',
			role VARCHAR(20) DEFAULT 'user',
			account_status VARCHAR(20) DEFAULT 'active',
			email_verified BOOLEAN DEFAULT FALSE,
			phone_verified BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			last_login TIMESTAMP WITH TIME ZONE
		)`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create user_addresses table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_addresses (
			address_id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			address_type VARCHAR(20) NOT NULL,
			street_address1 VARCHAR(255) NOT NULL,
			street_address2 VARCHAR(255),
			city VARCHAR(100) NOT NULL,
			state VARCHAR(100) NOT NULL,
			postal_code VARCHAR(20) NOT NULL,
			country VARCHAR(100) NOT NULL,
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("failed to create user_addresses table: %w", err)
	}

	// Create payment_methods table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS payment_methods (
			payment_method_id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			payment_type VARCHAR(20) NOT NULL,
			card_last_four VARCHAR(4),
			card_brand VARCHAR(20),
			expiration_month SMALLINT,
			expiration_year SMALLINT,
			is_default BOOLEAN DEFAULT FALSE,
			billing_address_id BIGINT REFERENCES user_addresses(address_id),
			token TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("failed to create payment_methods table: %w", err)
	}

	// Create user_preferences table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_preferences (
			user_id BIGINT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
			language VARCHAR(10) DEFAULT 'en',
			currency VARCHAR(3) DEFAULT 'USD',
			notification_email BOOLEAN DEFAULT TRUE,
			notification_sms BOOLEAN DEFAULT FALSE,
			theme VARCHAR(20) DEFAULT 'light',
			timezone VARCHAR(50) DEFAULT 'UTC',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("failed to create user_preferences table: %w", err)
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_user_addresses_user_id ON user_addresses(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payment_methods_user_id ON payment_methods(user_id)`,
	}

	for _, idx := range indexes {
		_, err = tx.ExecContext(ctx, idx)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Database tables and indexes created successfully")
	return nil
}

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database configuration with master and replicas
	dbConfig, err := db.NewDBConfig(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database configuration", zap.Error(err))
	}
	defer dbConfig.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbConfig.Master.PingContext(ctx); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	// Initialize database tables
	if err := initializeDatabase(ctx, dbConfig.Master.DB, logger); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize repository
	logger.Info("Initializing repository...")
	repo := repository.NewPostgresRepository(dbConfig, logger)

	// Initialize rate limiter
	rateLimiter := service.NewSimpleRateLimiter(
		cfg.RateLimiter.Attempts,
		cfg.RateLimiter.Duration,
	)

	accessTokenDuration, err := time.ParseDuration(os.Getenv("JWT_ACCESS_TOKEN_DURATION"))
	if err != nil {
		accessTokenDuration = 24 * time.Hour // default to 24 hours
	}

	refreshTokenDuration, err := time.ParseDuration(os.Getenv("JWT_REFRESH_TOKEN_DURATION"))
	if err != nil {
		refreshTokenDuration = 7 * 24 * time.Hour // default to 7 days
	}

	// Initialize JWT manager
	privateKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
	if privateKeyPath == "" {
		privateKeyPath = "certificates/private_key.pem" // Default path
	}
	publicKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
	if publicKeyPath == "" {
		publicKeyPath = "certificates/public_key.pem" // Default path
	}

	jwtManager, err := service.NewJWTManager(
		privateKeyPath,
		publicKeyPath,
		accessTokenDuration,
		refreshTokenDuration,
		repo,   // Pass the repository
		logger, // Pass the logger
	)
	if err != nil {
		logger.Fatal("Failed to initialize JWT manager", zap.Error(err))
	}

	// Determine Redis host based on environment
	redisHost := cfg.Redis.Host
	if os.Getenv("APP_ENV") == "development" && os.Getenv("DOCKER_ENV") != "true" {
		redisHost = "localhost"
	}

	redisAddr := fmt.Sprintf("%s:%s", redisHost, cfg.Redis.Port)
	logger.Info("Connecting to Redis",
		zap.String("address", redisAddr))

	// Initialize tiered cache manager with circuit breaker
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	cacheManager, err := cache.NewTieredUserCacheManager(cache.TieredUserCacheOptions{
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDB:       redisDB,
		RedisPoolSize: 10,
		DefaultTTL:    30 * time.Minute,
		Logger:        logger,
		// Circuit breaker settings
		FailureThreshold:         5,
		ResetTimeout:             30 * time.Second,
		HalfOpenSuccessThreshold: 2,
	})

	// Warm up cache with critical data
	logger.Info("Starting cache warm-up")
	go func() {
		// Wait a bit for services to initialize
		time.Sleep(2 * time.Second)

		// Create context with timeout for warm-up
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Warm up cache with critical data
		result, err := cacheManager.WarmupCache(ctx)
		if err != nil {
			logger.Error("Cache warm-up failed", zap.Error(err))
			return
		}

		logger.Info("Cache warm-up completed",
			zap.Int("successCount", result.SuccessCount),
			zap.Int("errorCount", result.ErrorCount),
			zap.Duration("duration", result.Duration))
	}()
	if err != nil {
		logger.Fatal("Failed to initialize tiered cache manager",
			zap.Error(err),
			zap.String("redis_addr", redisAddr))
	}

	// Initialize service with all required dependencies
	userService := service.NewUserService(
		repo,
		cacheManager,
		logger,
		rateLimiter,
		jwtManager,
	)

	// Initialize handler
	userHandler := handlers.NewUserHandler(userService, logger, jwtManager)

	// Set up gRPC server
	var opts []grpc.ServerOption
	if cfg.Server.Environment == "production" {
		// Load TLS credentials
		creds, err := credentials.NewServerTLSFromFile(
			cfg.Server.TLS.CertPath,
			cfg.Server.TLS.KeyPath,
		)
		if err != nil {
			logger.Fatal("Failed to load TLS credentials", zap.Error(err))
		}
		opts = append(opts, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterUserServiceServer(grpcServer, userHandler)

	// Start the server
	lis, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	logger.Info("Starting gRPC server", zap.String("port", cfg.Server.Port))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}
}
