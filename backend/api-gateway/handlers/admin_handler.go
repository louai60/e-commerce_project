package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	productpb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	userpb "github.com/louai60/e-commerce_project/backend/user-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminHandler struct {
    productClient productpb.ProductServiceClient
    userClient    userpb.UserServiceClient
    logger        *zap.Logger
}

func NewAdminHandler(
    productClient productpb.ProductServiceClient,
    userClient userpb.UserServiceClient,
    logger *zap.Logger,
) *AdminHandler {
    return &AdminHandler{
        productClient: productClient,
        userClient:    userClient,
        logger:        logger,
    }
}

// Product Management
func (h *AdminHandler) CreateProduct(c *gin.Context) {
    var req productpb.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.productClient.CreateProduct(c.Request.Context(), &req)
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

    resp, err := h.productClient.UpdateProduct(c.Request.Context(), &req)
    if err != nil {
        h.logger.Error("Failed to update product", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *AdminHandler) DeleteProduct(c *gin.Context) {
    productID := c.Param("id")
    
    _, err := h.productClient.DeleteProduct(c.Request.Context(), &productpb.DeleteProductRequest{Id: productID})
    if err != nil {
        h.logger.Error("Failed to delete product", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func (h *AdminHandler) GetProduct(c *gin.Context) {
    productID := c.Param("id")
    
    resp, err := h.productClient.GetProduct(c.Request.Context(), &productpb.GetProductRequest{Id: productID})
    if err != nil {
        h.logger.Error("Failed to get product", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *AdminHandler) ListProducts(c *gin.Context) {
    resp, err := h.productClient.ListProducts(c.Request.Context(), &productpb.ListProductsRequest{})
    if err != nil {
        h.logger.Error("Failed to list products", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list products"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// Dashboard Stats
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
    ctx := c.Request.Context()

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

// User Management
func (h *AdminHandler) ListUsers(c *gin.Context) {
    page, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
    if err != nil || page < 1 {
        page = 1
    }

    limit, err := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 32)
    if err != nil || limit < 1 {
        limit = 10
    }

    resp, err := h.userClient.ListUsers(c.Request.Context(), &userpb.ListUsersRequest{
        Page:  int32(page),
        Limit: int32(limit),
    })
    if err != nil {
        h.logger.Error("Failed to list users", zap.Error(err))
        h.handleGRPCError(c, err, "Failed to list users")
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *AdminHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    id, err := strconv.ParseInt(userID, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    resp, err := h.userClient.GetUser(c.Request.Context(), &userpb.GetUserRequest{
        UserId: id,
    })
    if err != nil {
        h.logger.Error("Failed to get user", 
            zap.String("user_id", userID),
            zap.Error(err))
        h.handleGRPCError(c, err, "Failed to get user")
        return
    }

    c.JSON(http.StatusOK, resp)
}

// func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
//     userID := c.Param("id")
//     id, err := strconv.ParseInt(userID, 10, 64)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
//         return
//     }

//     var req userpb.UpdateUserRoleRequest
//     if err := c.ShouldBindJSON(&req); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//         return
//     }
//     req.UserId = id

//     resp, err := h.userClient.UpdateUserRole(c.Request.Context(), &req)

    
//     if err != nil {
//         h.logger.Error("Failed to update user role",
//             zap.String("user_id", userID),
//             zap.Error(err))
//         h.handleGRPCError(c, err, "Failed to update user role")
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{
//         "message": "User role updated successfully",
//         "user": resp.User,
//     })
// }

// func (h *AdminHandler) BlockUser(c *gin.Context) {
//     userID := c.Param("id")
//     id, err := strconv.ParseInt(userID, 10, 64)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
//         return
//     }

//     resp, err := h.userClient.UpdateUserStatus(c.Request.Context(), &userpb.UpdateUserStatusRequest{
//         UserId: id,
//         Status: "blocked",
//     })
//     if err != nil {
//         h.logger.Error("Failed to block user", 
//             zap.String("user_id", userID),
//             zap.Error(err))
//         h.handleGRPCError(c, err, "Failed to block user")
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{
//         "message": "User blocked successfully",
//         "user": resp.User,
//     })
// }

// func (h *AdminHandler) UnblockUser(c *gin.Context) {
//     userID := c.Param("id")
//     id, err := strconv.ParseInt(userID, 10, 64)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
//         return
//     }

//     resp, err := h.userClient.UpdateUserStatus(c.Request.Context(), &userpb.UpdateUserStatusRequest{
//         UserId: id,
//         Status: "active",
//     })
//     if err != nil {
//         h.logger.Error("Failed to unblock user", 
//             zap.String("user_id", userID),
//             zap.Error(err))
//         h.handleGRPCError(c, err, "Failed to unblock user")
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{
//         "message": "User unblocked successfully",
//         "user": resp.User,
//     })
// }

func (h *AdminHandler) DeleteUser(c *gin.Context) {
    userID := c.Param("id")
    id, err := strconv.ParseInt(userID, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    _, err = h.userClient.DeleteUser(c.Request.Context(), &userpb.DeleteUserRequest{
        UserId: id,
    })
    if err != nil {
        h.logger.Error("Failed to delete user", 
            zap.String("user_id", userID),
            zap.Error(err))
        h.handleGRPCError(c, err, "Failed to delete user")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// func (h *AdminHandler) SearchUsers(c *gin.Context) {
//     query := c.Query("q")
//     if query == "" {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
//         return
//     }

//     resp, err := h.userClient.SearchUsers(c.Request.Context(), &userpb.SearchUsersRequest{
//         Query: query,
//     })
//     if err != nil {
//         h.logger.Error("Failed to search users", 
//             zap.String("query", query),
//             zap.Error(err))
//         h.handleGRPCError(c, err, "Failed to search users")
//         return
//     }

//     c.JSON(http.StatusOK, resp)
// }

func (h *AdminHandler) handleGRPCError(c *gin.Context, err error, defaultMsg string) {
    st, ok := status.FromError(err)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": defaultMsg})
        return
    }

    switch st.Code() {
    case codes.NotFound:
        c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
    case codes.InvalidArgument:
        c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
    case codes.PermissionDenied:
        c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": defaultMsg})
    }
}

