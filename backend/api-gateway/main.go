package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // Added godotenv back
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/louai60/e-commerce_project/backend/api-gateway/handlers"
	"github.com/louai60/e-commerce_project/backend/api-gateway/middleware"
	adminpb "github.com/louai60/e-commerce_project/backend/admin-service/proto"
	productpb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

func main() {
	// Load .env file before anything else
	if err := godotenv.Load(); err != nil { // Use godotenv here
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
        grpc.WithBlock(),
        grpc.WithTimeout(2*time.Second), // Reduced timeout
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

    adminClient := adminpb.NewAdminServiceClient(adminConn) // Create admin client
    // The adminHandler needs to be defined in the handlers package for the gateway
    // For now, we comment this out until we create the gateway's admin handler
    adminHandler := handlers.NewAdminHandler(adminClient, logger) // Uncommented: Initialize the gateway's admin handler

    userHandler, err := handlers.NewUserHandler("localhost:50052", logger)
    if err != nil {
        logger.Fatal("Failed to initialize user handler", zap.Error(err))
    }

    // Initialize Gin router
    r := gin.Default()
    r.Use(middleware.Logger(logger), middleware.CORSMiddleware())

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

        // Brand routes
        brands := v1.Group("/brands")
        {
            brands.GET("", productHandler.ListBrands)
            brands.GET("/:id", productHandler.GetBrand)
            brands.POST("", middleware.AuthRequired(), middleware.AdminRequired(), productHandler.CreateBrand)
        }

        // Category routes
        categories := v1.Group("/categories")
        {
            categories.GET("", productHandler.ListCategories)
            categories.GET("/:id", productHandler.GetCategory)
            categories.POST("", middleware.AuthRequired(), middleware.AdminRequired(), productHandler.CreateCategory)
        }

        // User routes
        users := v1.Group("/users")
        {
            users.POST("/register", userHandler.Register)
            users.POST("/login", userHandler.Login)
            users.POST("/logout", userHandler.Logout) // Add logout route
            users.POST("/refresh", userHandler.RefreshToken)
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

        // Admin Dashboard routes (protected)
        adminDashboard := v1.Group("/admin/dashboard", middleware.AuthRequired(), middleware.AdminRequired())
        {
            adminDashboard.GET("/stats", adminHandler.GetDashboardStats)
            // Add more admin dashboard routes here later
        }
    }

    // Start server
    serverAddr := ":8080"
    logger.Info("Starting API Gateway", zap.String("address", serverAddr))
    if err := r.Run(serverAddr); err != nil {
        logger.Fatal("Failed to start server", zap.Error(err))
    }
}
