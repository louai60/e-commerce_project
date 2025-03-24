package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "github.com/louai60/e-commerce_project/backend/api-gateway/config"
    "github.com/louai60/e-commerce_project/backend/api-gateway/clients"
    "github.com/louai60/e-commerce_project/backend/api-gateway/handlers"
    "github.com/louai60/e-commerce_project/backend/api-gateway/middleware"
)

func main() {
    // Initialize logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        logger.Fatal("Failed to load configuration", zap.Error(err))
    }

    // Initialize service clients
    serviceClients, err := clients.NewServiceClients(cfg)
    if err != nil {
        logger.Fatal("Failed to initialize service clients", zap.Error(err))
    }
    defer serviceClients.Close()

    // Initialize Gin router
    router := gin.New()
    router.Use(gin.Recovery())
    router.Use(middleware.Logger(logger))
    router.Use(middleware.CORS())

    // Initialize handlers
    productHandler := handlers.NewProductHandler(serviceClients.ProductClient, logger)
    // userHandler := handlers.NewUserHandler(serviceClients.UserClient, logger)
    // orderHandler := handlers.NewOrderHandler(serviceClients.OrderClient, logger)
    // cartHandler := handlers.NewCartHandler(serviceClients.CartClient, logger)
    // Initialize other handlers

    // Register routes
    v1 := router.Group("/api/v1")
    {
        // Product routes
        products := v1.Group("/products")
        {
            products.GET("", productHandler.ListProducts)
            products.GET("/:id", productHandler.GetProduct)
            products.POST("", middleware.AuthRequired(), productHandler.CreateProduct)
            products.PUT("/:id", middleware.AuthRequired(), productHandler.UpdateProduct)
            products.DELETE("/:id", middleware.AuthRequired(), productHandler.DeleteProduct)
        }

        // User routes
        // users := v1.Group("/users")
        // {
        //     users.POST("/register", userHandler.Register)
        //     users.POST("/login", userHandler.Login)
        //     users.GET("/profile", middleware.AuthRequired(), userHandler.GetProfile)
        //     users.PUT("/profile", middleware.AuthRequired(), userHandler.UpdateProfile)
        // }

        // Order routes
        // orders := v1.Group("/orders", middleware.AuthRequired())
        // {
        //     orders.POST("", orderHandler.CreateOrder)
        //     orders.GET("", orderHandler.ListOrders)
        //     orders.GET("/:id", orderHandler.GetOrder)
        //     orders.PUT("/:id/status", orderHandler.UpdateOrderStatus)
        // }

        // Cart routes
        // cart := v1.Group("/cart", middleware.AuthRequired())
        // {
        //     cart.GET("", cartHandler.GetCart)
        //     cart.POST("/items", cartHandler.AddItem)
        //     cart.PUT("/items/:id", cartHandler.UpdateItem)
        //     cart.DELETE("/items/:id", cartHandler.RemoveItem)
        // }

        // Add other service routes
    }

    // Create HTTP server
    srv := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: router,
    }

    // Start server in a goroutine
    go func() {
        logger.Info("Starting server", zap.String("port", cfg.Server.Port))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("Failed to start server", zap.Error(err))
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // Graceful shutdown
    logger.Info("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Fatal("Server forced to shutdown", zap.Error(err))
    }

    logger.Info("Server exited properly")
}
