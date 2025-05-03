package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/louai60/e-commerce_project/backend/api-gateway/handlers"
	"github.com/louai60/e-commerce_project/backend/api-gateway/middleware"
)

func SetupRoutes(r *gin.Engine, productHandler *handlers.ProductHandler, userHandler *handlers.UserHandler, adminHandler *handlers.AdminHandler, inventoryHandler *handlers.InventoryHandler) {
	// API routes
	v1 := r.Group("/api/v1")
	{
		// Create middleware to add inventory client to context
		inventoryClientMiddleware := func(c *gin.Context) {
			c.Set("inventory_client", inventoryHandler.GetClient())
			c.Next()
		}

		// Product routes
		products := v1.Group("/products", inventoryClientMiddleware)
		{
			products.GET("", productHandler.ListProducts)
			products.GET("/:id", productHandler.GetProduct)
			// Add inventory client to the context for product creation
			products.POST("", middleware.AuthRequired(), middleware.AdminRequired(), func(c *gin.Context) {
				// Use the product_inventory_handler to create product with inventory
				handlers.CreateProductWithInventory(c, productHandler.GetClient(), inventoryHandler.GetClient(), productHandler.GetLogger())
			})
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
			users.POST("/logout", userHandler.Logout)
			users.POST("/refresh", userHandler.RefreshToken)
			users.POST("/admin", middleware.AdminKeyRequired(), userHandler.CreateAdmin)

			// Protected routes
			authenticated := users.Group("/", middleware.AuthRequired())
			{
				authenticated.GET("/profile", userHandler.GetProfile)
				authenticated.PUT("/profile", userHandler.UpdateProfile)

				// Address management
				authenticated.POST("/addresses", userHandler.AddAddress)

				// Payment methods
				authenticated.POST("/payment-methods", userHandler.AddPaymentMethod)

				// Admin only routes
				admin := authenticated.Group("/", middleware.AdminRequired())
				{
					admin.GET("", userHandler.ListUsers)
					admin.GET("/:id", userHandler.GetUser)
					admin.DELETE("/:id", userHandler.DeleteUser)
				}
			}
		}

		// Image routes
		images := v1.Group("/images")
		{
			images.POST("/upload", middleware.AuthRequired(), middleware.AdminRequired(), productHandler.UploadImage)
			images.DELETE("/:public_id", middleware.AuthRequired(), middleware.AdminRequired(), productHandler.DeleteImage)
		}

		// Admin Dashboard routes (protected)
		adminDashboard := v1.Group("/admin/dashboard", middleware.AuthRequired(), middleware.AdminRequired())
		{
			adminDashboard.GET("/stats", adminHandler.GetDashboardStats)
		}

		// Inventory routes (most require admin access)
		inventory := v1.Group("/inventory")
		{
			// Public routes
			inventory.GET("/check", inventoryHandler.CheckInventoryAvailability)

			// Protected routes
			protected := inventory.Group("/", middleware.AuthRequired(), middleware.AdminRequired())
			{
				protected.GET("/items", inventoryHandler.ListInventoryItems)
				protected.GET("/items/:product_id", inventoryHandler.GetInventoryItem)
				protected.GET("/warehouses", inventoryHandler.ListWarehouses)
				protected.GET("/transactions", inventoryHandler.ListInventoryTransactions)
			}
		}
	}
}
