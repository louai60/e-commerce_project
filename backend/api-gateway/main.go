package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "github.com/louai60/e-commerce_project/backend/api-gateway/handlers"
    "github.com/louai60/e-commerce_project/backend/api-gateway/middleware"
    productpb "github.com/louai60/e-commerce_project/backend/product-service/proto"
    userpb "github.com/louai60/e-commerce_project/backend/user-service/proto"
    // adminpb "github.com/louai60/e-commerce_project/backend/admin-service/proto"
    "github.com/joho/godotenv"
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

    // Initialize gRPC connections
    productConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        logger.Fatal("Failed to connect to product service", zap.Error(err))
    }
    defer productConn.Close()

    userConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        logger.Fatal("Failed to connect to user service", zap.Error(err))
    }
    defer userConn.Close()

    adminConn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        logger.Fatal("Failed to connect to admin service", zap.Error(err))
    }
    defer adminConn.Close()

    // Initialize gRPC clients
    productClient := productpb.NewProductServiceClient(productConn)
    userClient := userpb.NewUserServiceClient(userConn)
    // Remove unused adminClient
    
    // Initialize handlers
    productHandler := handlers.NewProductHandler(productClient, logger)
    userHandler, err := handlers.NewUserHandler("localhost:50052", logger) 
    if err != nil {
        logger.Fatal("Failed to create user handler", zap.Error(err))
    }
    adminHandler := handlers.NewAdminHandler(productClient, userClient, logger)

    // Initialize Gin router
    r := gin.Default()
    r.Use(middleware.Logger(logger), middleware.CORSMiddleware())

    // API routes
    v1 := r.Group("/api/v1")
    {
        // Product routes (public)
        products := v1.Group("/products")
        {
            products.GET("", productHandler.ListProducts)
            products.GET("/:id", productHandler.GetProduct)
        }

        // User routes
        users := v1.Group("/users")
        {
            users.POST("/register", userHandler.Register)
            users.POST("/login", userHandler.Login)     
            users.POST("/refresh", userHandler.RefreshToken)
            
            // Protected routes
            authenticated := users.Group("/", middleware.AuthRequired())
            {
                authenticated.GET("/profile", userHandler.GetProfile)
                authenticated.PUT("/profile", userHandler.UpdateProfile)
                
                // Address management
                authenticated.POST("/addresses", userHandler.AddAddress)
                // authenticated.GET("/addresses", userHandler.ListAddresses)
            }
        }

        // Admin routes
        admin := v1.Group("/admin", middleware.AuthRequired(), middleware.AdminRequired())
        {
            // Dashboard
            admin.GET("/dashboard/stats", adminHandler.GetDashboardStats)

            // Admin Product Management
            adminProducts := admin.Group("/products")
            {
                adminProducts.GET("", adminHandler.ListProducts)
                adminProducts.POST("", adminHandler.CreateProduct)
                adminProducts.GET("/:id", adminHandler.GetProduct)
                adminProducts.PUT("/:id", adminHandler.UpdateProduct)
                adminProducts.DELETE("/:id", adminHandler.DeleteProduct)
            }

            // Admin User Management
            adminUsers := admin.Group("/users")
            {
                adminUsers.GET("", adminHandler.ListUsers)
                adminUsers.GET("/:id", adminHandler.GetUser)
                // adminUsers.POST("/roles", adminHandler.UpdateUserRole)
                adminUsers.DELETE("/:id", adminHandler.DeleteUser)
            }
        }
    }

    // Start server
    serverAddr := ":8080"
    logger.Info("Starting API Gateway", zap.String("address", serverAddr))
    if err := r.Run(serverAddr); err != nil {
        logger.Fatal("Failed to start server", zap.Error(err))
    }
}





