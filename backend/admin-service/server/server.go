package server

import (
    // "context"
    // "net/http"
    // "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "github.com/louai60/e-commerce_project/backend/admin-service/handlers"
    "github.com/louai60/e-commerce_project/backend/admin-service/middleware"
)

type Server struct {
    router  *gin.Engine
    logger  *zap.Logger
    handler *handlers.AdminHandler
}

func NewServer(logger *zap.Logger, handler *handlers.AdminHandler) *Server {
    router := gin.Default()
    
    // Add middleware
    router.Use(middleware.AdminAuth())
    router.Use(middleware.LoggerMiddleware(logger))
    router.Use(middleware.CORSMiddleware())

    return &Server{
        router:  router,
        logger:  logger,
        handler: handler,
    }
}

func (s *Server) SetupRoutes() {
    api := s.router.Group("/api/v1")
    {
        // Dashboard
        api.GET("/dashboard/stats", s.handler.GetDashboardStats)

        // Products management
        products := api.Group("/products")
        {
            products.GET("", s.handler.ListProducts)
            products.POST("", s.handler.CreateProduct)
            products.PUT("/:id", s.handler.UpdateProduct)
            products.DELETE("/:id", s.handler.DeleteProduct)
            products.GET("/:id", s.handler.GetProduct)
        }

        // User management
        users := api.Group("/users")
        {
            users.GET("", s.handler.ListUsers)
            users.GET("/:id", s.handler.GetUser)
            users.POST("/roles", s.handler.UpdateUserRole)
            users.DELETE("/:id", s.handler.DeleteUser)
        }
    }
}

func (s *Server) Start(port string) error {
    return s.router.Run(":" + port)
}
