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
)

func main() {
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

    // Initialize handlers
    productClient := productpb.NewProductServiceClient(productConn)
    productHandler := handlers.NewProductHandler(productClient, logger)

    userHandler, err := handlers.NewUserHandler("localhost:50052", logger)
    if err != nil {
        logger.Fatal("Failed to initialize user handler", zap.Error(err))
    }

    // Initialize Gin router
    r := gin.Default()
    r.Use(middleware.Logger(logger))

    // API routes
    v1 := r.Group("/api/v1")
    {
        // Product routes
        products := v1.Group("/products")
        {
            products.GET("", productHandler.ListProducts)
            products.GET("/:id", productHandler.GetProduct)
            products.POST("", middleware.AuthRequired(), middleware.AdminRequired(), productHandler.CreateProduct)
            products.PUT("/:id", middleware.AuthRequired(), middleware.AdminRequired(), productHandler.UpdateProduct)
            products.DELETE("/:id", middleware.AuthRequired(), middleware.AdminRequired(), productHandler.DeleteProduct)
        }

        // User routes
        users := v1.Group("/users")
        {
            users.POST("/register", userHandler.Register)
            users.POST("/login", userHandler.Login)
            users.GET("/profile", middleware.AuthRequired(), userHandler.GetProfile)
            users.PUT("/profile", middleware.AuthRequired(), userHandler.UpdateProfile)
            users.GET("", middleware.AuthRequired(), middleware.AdminRequired(), userHandler.ListUsers)
            users.GET("/:id", middleware.AuthRequired(), userHandler.GetUser)
            users.DELETE("/:id", middleware.AuthRequired(), middleware.AdminRequired(), userHandler.DeleteUser)
        }
    }

    // Start server
    serverAddr := ":8080"
    logger.Info("Starting API Gateway", zap.String("address", serverAddr))
    if err := r.Run(serverAddr); err != nil {
        logger.Fatal("Failed to start server", zap.Error(err))
    }
}


