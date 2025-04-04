package handlers

import (
    "context"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    productpb "github.com/louai60/e-commerce_project/backend/product-service/proto"
    userpb "github.com/louai60/e-commerce_project/backend/user-service/proto"
)

type AdminHandler struct {
    logger         *zap.Logger
    productClient  productpb.ProductServiceClient
    userClient     userpb.UserServiceClient
}

func NewAdminHandler(
    logger *zap.Logger,
    productClient productpb.ProductServiceClient,
    userClient userpb.UserServiceClient,
) *AdminHandler {
    return &AdminHandler{
        logger:         logger,
        productClient:  productClient,
        userClient:     userClient,
    }
}

// Dashboard Stats
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
    ctx := context.Background()

    productsResp, err := h.productClient.GetProductsCount(ctx, &productpb.GetProductsCountRequest{})
    if err != nil {
        h.logger.Error("Failed to get products count", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard stats"})
        return
    }

    usersResp, err := h.userClient.GetUsersCount(ctx, &userpb.GetUsersCountRequest{})
    if err != nil {
        h.logger.Error("Failed to get users count", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard stats"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "products_count": productsResp.Count,
        "users_count":    usersResp.Count,
    })
}

// Product Management
func (h *AdminHandler) CreateProduct(c *gin.Context) {
    var req productpb.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx := context.Background()
    resp, err := h.productClient.CreateProduct(ctx, &req)
    if err != nil {
        h.logger.Error("Failed to create product", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
        return
    }

    c.JSON(http.StatusCreated, resp)
}

func (h *AdminHandler) UpdateProduct(c *gin.Context) {
    productID := c.Param("id")
    var req productpb.UpdateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    req.Id = productID

    ctx := context.Background()
    resp, err := h.productClient.UpdateProduct(ctx, &req)
    if err != nil {
        h.logger.Error("Failed to update product", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *AdminHandler) DeleteProduct(c *gin.Context) {
    productID := c.Param("id")
    ctx := context.Background()
    
    _, err := h.productClient.DeleteProduct(ctx, &productpb.DeleteProductRequest{Id: productID})
    if err != nil {
        h.logger.Error("Failed to delete product", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func (h *AdminHandler) GetProduct(c *gin.Context) {
    productID := c.Param("id")
    ctx := context.Background()
    
    resp, err := h.productClient.GetProduct(ctx, &productpb.GetProductRequest{Id: productID})
    if err != nil {
        h.logger.Error("Failed to get product", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *AdminHandler) ListProducts(c *gin.Context) {
    ctx := context.Background()
    resp, err := h.productClient.ListProducts(ctx, &productpb.ListProductsRequest{})
    if err != nil {
        h.logger.Error("Failed to list products", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list products"})
        return
    }
    c.JSON(http.StatusOK, resp)
}

// User Management
func (h *AdminHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    ctx := context.Background()
    
    userIDInt, err := strconv.ParseInt(userID, 10, 64)
    if err != nil {
        h.logger.Error("Invalid user ID format", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }
    resp, err := h.userClient.GetUser(ctx, &userpb.GetUserRequest{UserId: userIDInt})
    if err != nil {
        h.logger.Error("Failed to get user", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
    ctx := context.Background()
    resp, err := h.userClient.ListUsers(ctx, &userpb.ListUsersRequest{})
    if err != nil {
        h.logger.Error("Failed to list users", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
        return
    }
    c.JSON(http.StatusOK, resp)
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
    var req struct {
        UserID string `json:"user_id"`
        Role   string `json:"role"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx := context.Background()
    resp, err := h.userClient.UpdateUserRole(ctx, &userpb.UpdateUserRoleRequest{
        UserId: req.UserID,
        Role:   req.Role,
    })
    if err != nil {
        h.logger.Error("Failed to update user role", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Role updated successfully",
        "user": resp.User,
    })
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
    userID := c.Param("id")
    ctx := context.Background()
    
    userIDInt, err := strconv.ParseInt(userID, 10, 64)
    if err != nil {
        h.logger.Error("Invalid user ID format", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }
    _, err = h.userClient.DeleteUser(ctx, &userpb.DeleteUserRequest{UserId: userIDInt})
    if err != nil {
        h.logger.Error("Failed to delete user", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

