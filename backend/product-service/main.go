package main

import (
	"net"
	"os"
	"fmt"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"github.com/louai60/e-commerce_project/backend/common/logger"
	"github.com/louai60/e-commerce_project/backend/product-service/config"
	"github.com/louai60/e-commerce_project/backend/product-service/handlers"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
	"github.com/louai60/e-commerce_project/backend/product-service/service"
	"github.com/louai60/e-commerce_project/backend/product-service/cache"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"google.golang.org/grpc"
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

	// Initialize repository with PostgreSQL
	repo, err := repository.NewPostgresRepository(cfg.GetDSN())
	if err != nil {
		log.Fatal("Failed to initialize repository", zap.Error(err))
	}

	// Initialize cache manager
	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	cacheManager, err := cache.NewCacheManager(redisAddr)
	if err != nil {
		log.Fatal("Failed to initialize cache manager", zap.Error(err))
	}

	// Initialize service
	productService := service.NewProductService(repo, cacheManager, log)

	// Initialize handler
	productHandler := handlers.NewProductHandler(productService)

	// Set up gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
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
