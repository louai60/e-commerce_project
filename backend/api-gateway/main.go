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
            users.POST("/admin", middleware.AdminKeyRequired(), userHandler.CreateAdmin)
            
            // Protected routes
            authenticated := users.Group("/", middleware.AuthRequired())
            {
                authenticated.GET("/profile", userHandler.GetProfile)
                authenticated.PUT("/profile", userHandler.UpdateProfile)
                
                // Address management
                authenticated.POST("/addresses", userHandler.AddAddress)
                // authenticated.GET("/addresses", userHandler.ListAddresses)
                // authenticated.PUT("/addresses/:id", userHandler.UpdateAddress)
                // authenticated.DELETE("/addresses/:id", userHandler.DeleteAddress)
                
                // Payment methods
                authenticated.POST("/payment-methods", userHandler.AddPaymentMethod)
                // authenticated.GET("/payment-methods", userHandler.ListPaymentMethods)
                // authenticated.PUT("/payment-methods/:id", userHandler.UpdatePaymentMethod)
                // authenticated.DELETE("/payment-methods/:id", userHandler.DeletePaymentMethod)
                
                // Admin only routes
                admin := authenticated.Group("/", middleware.AdminRequired())
                {
                    admin.GET("", userHandler.ListUsers)
                    admin.GET("/:id", userHandler.GetUser)
                    admin.DELETE("/:id", userHandler.DeleteUser)
                }
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





