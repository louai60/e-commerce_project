package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/louai60/e-commerce_project/backend/admin-service/handlers"
	"github.com/louai60/e-commerce_project/backend/admin-service/server"
	productpb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	userpb "github.com/louai60/e-commerce_project/backend/user-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type AdminServer struct {
    logger         *zap.Logger
    productClient  productpb.ProductServiceClient
    userClient     userpb.UserServiceClient
    // Add other service clients as needed
}

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        if !os.IsNotExist(err) {
            panic("Error loading .env file: " + err.Error())
        }
    }

    // Initialize logger
    logger, err := zap.NewDevelopment()
    if err != nil {
        panic("Failed to initialize logger: " + err.Error())
    }
    defer logger.Sync()

    // Get service addresses from environment variables
    productServiceAddr := os.Getenv("PRODUCT_SERVICE_ADDR")
    if productServiceAddr == "" {
        productServiceAddr = "localhost:50051" // Default
        logger.Warn("PRODUCT_SERVICE_ADDR not set, using default", zap.String("address", productServiceAddr))
    }
    userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
    if userServiceAddr == "" {
        userServiceAddr = "localhost:50052" // Default
        logger.Warn("USER_SERVICE_ADDR not set, using default", zap.String("address", userServiceAddr))
    }

    // Initialize gRPC connections to other services
    productConn, err := grpc.Dial(productServiceAddr, grpc.WithInsecure()) // Use environment variable
    if err != nil {
        logger.Fatal("Failed to connect to product service", zap.String("address", productServiceAddr), zap.Error(err))
    }
    defer productConn.Close()
    logger.Info("Connected to product service", zap.String("address", productServiceAddr))

    userConn, err := grpc.Dial(userServiceAddr, grpc.WithInsecure()) // Use environment variable
    if err != nil {
        logger.Fatal("Failed to connect to user service", zap.String("address", userServiceAddr), zap.Error(err))
    }
    defer userConn.Close()
    logger.Info("Connected to user service", zap.String("address", userServiceAddr))

    // Initialize admin server
    adminServer := &AdminServer{
        logger:         logger,
        productClient:  productpb.NewProductServiceClient(productConn),
        userClient:     userpb.NewUserServiceClient(userConn),
    }

    // Start HTTP server
    if err := adminServer.Start(); err != nil {
        logger.Fatal("Failed to start server", zap.Error(err))
    }
}

func (s *AdminServer) Start() error {
    // Initialize handler
    adminHandler := handlers.NewAdminHandler(
        s.logger,
        s.productClient,
        s.userClient,
    )

    // Initialize server
    server := server.NewServer(s.logger, adminHandler)
    
    // Setup routes
    server.SetupRoutes()

    // Start server
    port := os.Getenv("ADMIN_SERVICE_PORT")
    if port == "" {
        port = "8085" // Default port from config
    }
    
    s.logger.Info("Starting admin service", zap.String("port", port))
    return server.Start(port)
}
