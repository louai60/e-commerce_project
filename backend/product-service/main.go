package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver (import driver for side effects)
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/louai60/e-commerce_project/backend/common/logger"
	"github.com/louai60/e-commerce_project/backend/product-service/cache"
	"github.com/louai60/e-commerce_project/backend/product-service/config"
	"github.com/louai60/e-commerce_project/backend/product-service/db"
	"github.com/louai60/e-commerce_project/backend/product-service/handlers"
	"github.com/louai60/e-commerce_project/backend/product-service/middleware"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
	"github.com/louai60/e-commerce_project/backend/product-service/repository/postgres"
	"github.com/louai60/e-commerce_project/backend/product-service/service"
)

func main() {
	// Load .env file before initializing logger
	if err := godotenv.Load(); err != nil {
		// Only log error if .env file exists but couldn't be loaded
		if !os.IsNotExist(err) {
			panic("Error loading .env file: " + err.Error())
		}
	}

	// Initialize logger first
	log := logger.GetLogger()
	defer log.Sync()

	// Load configuration
	cfg, err := config.LoadConfig(log)
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database configuration with master and replicas
	dbConfig, err := db.NewDBConfig(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize database configuration", zap.Error(err))
	}
	defer dbConfig.Close() // Ensure all db connections are closed when main exits

	// Create context with timeout for initialization
	// Commented out since we're not using it for migrations anymore
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	// Skip hardcoded migrations since we're using SQL migrations
	// Uncomment this if you need to run the hardcoded migrations
	// if err := repository.RunMigrations(ctx, db, log); err != nil {
	// 	log.Fatal("Failed to run database migrations", zap.Error(err))
	// }

	// Initialize repositories
	// Use the adapter to make the new repository compatible with the existing interface
	productRepo := postgres.NewProductRepositoryAdapter(dbConfig, log)
	// For now, use the master connection for other repositories
	brandRepo := repository.NewBrandRepository(dbConfig.Master, log)
	categoryRepo := repository.NewCategoryRepository(dbConfig.Master, log)

	// Initialize tiered cache manager with circuit breaker
	cacheManager, err := cache.NewTieredCacheManager(cache.TieredCacheOptions{
		RedisAddr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		RedisPassword: cfg.Redis.Password,
		RedisDB:       cfg.Redis.DB,
		RedisPoolSize: 10,
		DefaultTTL:    15 * time.Minute,
		Logger:        log,
		// Circuit breaker settings
		FailureThreshold:         5,
		ResetTimeout:             30 * time.Second,
		HalfOpenSuccessThreshold: 2,
	})
	if err != nil {
		log.Fatal("Failed to initialize tiered cache manager", zap.Error(err))
	}
	defer cacheManager.Close()

	// Warm up cache with critical data
	log.Info("Starting cache warm-up")
	go func() {
		// Wait a bit for services to initialize
		time.Sleep(2 * time.Second)

		// Create context with timeout for warm-up
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Warm up cache with critical data
		result, err := cacheManager.WarmupCache(ctx)
		if err != nil {
			log.Error("Cache warm-up failed", zap.Error(err))
			return
		}

		log.Info("Cache warm-up completed",
			zap.Int("successCount", result.SuccessCount),
			zap.Int("errorCount", result.ErrorCount),
			zap.Duration("duration", result.Duration))
	}()

	// Initialize service with all required repositories
	productService := service.NewProductService(
		productRepo,
		brandRepo,
		categoryRepo,
		cacheManager,
		log,
	)
	if productService == nil {
		log.Fatal("Failed to create product service")
	}

	// Initialize handler with the service
	productHandler := handlers.NewProductHandler(productService, log)
	if productHandler == nil {
		log.Fatal("Failed to create product handler")
	}

	// Set up gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.LoggingInterceptor(log)),
	)
	pb.RegisterProductServiceServer(grpcServer, productHandler)

	log.Info("Product service initialized",
		zap.String("environment", cfg.Server.Environment),
		zap.String("port", cfg.Server.Port),
	)

	// Start the server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve", zap.Error(err))
	}
}
