package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	adminpb "github.com/louai60/e-commerce_project/backend/admin-service/proto"
	"github.com/louai60/e-commerce_project/backend/api-gateway/clients"
	"github.com/louai60/e-commerce_project/backend/api-gateway/config"
	"github.com/louai60/e-commerce_project/backend/api-gateway/handlers"
	"github.com/louai60/e-commerce_project/backend/api-gateway/internal/routes"
	"github.com/louai60/e-commerce_project/backend/api-gateway/middleware"
	productpb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

func main() {
	// Load .env file before anything else
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Load JWT public key for token validation
	if err := middleware.LoadPublicKey(); err != nil {
		logger.Fatal("Failed to load JWT public key", zap.Error(err))
	}

	// Initialize gRPC connections
	productServiceAddr := os.Getenv("PRODUCT_SERVICE_ADDR")
	if productServiceAddr == "" {
		productServiceAddr = "localhost:50051" // fallback to default
	}

	var productConn *grpc.ClientConn
	var productClient productpb.ProductServiceClient

	// Try to connect to product service but don't block startup
	productConn, err = grpc.Dial(
		productServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("Failed to connect to product service - some functionality will be unavailable",
			zap.String("address", productServiceAddr),
			zap.Error(err))
	} else {
		defer productConn.Close()
		productClient = productpb.NewProductServiceClient(productConn)
	}

	// Initialize product handler with potential nil client
	productHandler := handlers.NewProductHandler(productClient, logger)

	userConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to user service", zap.Error(err))
	}
	defer userConn.Close()

	// Connect to Admin Service
	adminServiceAddr := os.Getenv("ADMIN_SERVICE_ADDR")
	if adminServiceAddr == "" {
		logger.Fatal("ADMIN_SERVICE_ADDR environment variable is required")
	}
	adminConn, err := grpc.Dial(adminServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to admin service", zap.Error(err))
	}
	defer adminConn.Close()

	adminClient := adminpb.NewAdminServiceClient(adminConn)
	adminHandler := handlers.NewAdminHandler(adminClient, logger)

	userHandler, err := handlers.NewUserHandler("localhost:50052", logger)
	if err != nil {
		logger.Fatal("Failed to initialize user handler", zap.Error(err))
	}

	// Connect to Inventory Service
	inventoryServiceAddr := os.Getenv("INVENTORY_SERVICE_ADDR")
	if inventoryServiceAddr == "" {
		inventoryServiceAddr = "localhost:50055" // fallback to default
	}

	// Create a config object for the inventory client
	inventoryConfig := &config.Config{
		Services: config.ServicesConfig{
			Inventory: config.ServiceConfig{
				Host: "localhost",
				Port: "50055",
			},
		},
	}

	// Parse host and port from the address
	if inventoryServiceAddr != "" {
		parts := strings.Split(inventoryServiceAddr, ":")
		if len(parts) == 2 {
			inventoryConfig.Services.Inventory.Host = parts[0]
			inventoryConfig.Services.Inventory.Port = parts[1]
		}
	}

	// Initialize inventory client
	var inventoryClient *clients.InventoryClient
	inventoryClient, err = clients.NewInventoryClient(inventoryConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to inventory service - some functionality will be unavailable",
			zap.String("address", inventoryServiceAddr),
			zap.Error(err))
		// Set to nil to ensure proper error handling in the handler
		inventoryClient = nil
	}

	// Initialize inventory handler with potential nil client
	inventoryHandler := handlers.NewInventoryHandler(inventoryClient, logger)

	// Initialize GraphQL handler
	graphqlHandler, err := handlers.NewGraphQLHandler(logger, inventoryClient, productClient)
	if err != nil {
		logger.Error("Failed to initialize GraphQL handler", zap.Error(err))
	}

	// Initialize Gin router
	r := gin.New() // Use New() instead of Default() to avoid using the default logger and recovery
	r.Use(middleware.Logger(logger), middleware.CORSMiddleware(), middleware.Recovery(logger))

	// Setup all routes
	routes.SetupRoutes(r, productHandler, userHandler, adminHandler, inventoryHandler)

	// Setup GraphQL routes if handler was initialized successfully
	if graphqlHandler != nil {
		routes.SetupGraphQLRoutes(r, graphqlHandler)
		logger.Info("GraphQL endpoint configured at /api/v1/graphql")
	}

	// Setup static file server for uploaded images
	// Create uploads directory if it doesn't exist
	uploadsDir := os.Getenv("LOCAL_STORAGE_PATH")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		logger.Error("Failed to create uploads directory", zap.Error(err))
	}
	r.Static("/uploads", uploadsDir)
	logger.Info("Static file server configured", zap.String("path", uploadsDir))

	// Start server
	serverAddr := ":8080"
	logger.Info("Starting API Gateway", zap.String("address", serverAddr))
	if err := r.Run(serverAddr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
