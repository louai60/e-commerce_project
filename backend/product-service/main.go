package main

import (
	"database/sql"
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
	"github.com/louai60/e-commerce_project/backend/product-service/handlers"
	"github.com/louai60/e-commerce_project/backend/product-service/middleware"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
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

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close() // Ensure db connection is closed when main exits

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

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
	productRepo := repository.NewProductRepository(db, log)
	brandRepo := repository.NewBrandRepository(db, log)
	categoryRepo := repository.NewCategoryRepository(db, log)

	// Initialize cache manager
	cacheManager, err := cache.NewCacheManager(cache.CacheOptions{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: 10,
		TTL:      15 * time.Minute,
	})
	if err != nil {
		log.Fatal("Failed to initialize cache manager", zap.Error(err))
	}
	defer cacheManager.Close()

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
