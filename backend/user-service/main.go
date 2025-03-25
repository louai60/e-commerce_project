package main

import (
	"context"
	"log"
	"net"
	// "os"
	// "os/signal"
	// "os/user"
	// "syscall"
	"time"

	"github.com/louai60/e-commerce_project/backend/user-service/config"
	"github.com/louai60/e-commerce_project/backend/user-service/handlers"
	pb "github.com/louai60/e-commerce_project/backend/user-service/proto"
	"github.com/louai60/e-commerce_project/backend/user-service/repository"
	"github.com/louai60/e-commerce_project/backend/user-service/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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

	// Initialize repository
	logger.Info("Initializing database connection and running migrations...")
	repo, err := repository.NewPostgresRepository(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize repository", zap.Error(err))
	}
	defer repo.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := repo.Ping(ctx); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Info("Successfully connected to database and initialized schema")

	// Initialize rate limiter
	rateLimiter := service.NewSimpleRateLimiter(
		cfg.RateLimiter.Attempts,
		cfg.RateLimiter.Duration,
	)

	// Initialize token manager
	tokenManager := service.NewJWTManager(
		cfg.Auth.SecretKey,
		cfg.Auth.AccessTokenDuration,
		// cfg.Auth.RefreshTokenDuration,
	)

	// Initialize service
	userService := service.NewUserService(
		repo,
		logger,
		rateLimiter,
		tokenManager,
	)

	// Initialize handler
	userHandler := handlers.NewUserHandler(userService, logger)

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
