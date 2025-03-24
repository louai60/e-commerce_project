package handlers

import (
    "context"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

type ProductHandler struct {
    client pb.ProductServiceClient
    logger *zap.Logger
}

func NewProductHandler(client pb.ProductServiceClient, logger *zap.Logger) *ProductHandler {
    return &ProductHandler{
        client: client,
        logger: logger,
    }
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
        return
    }

    req := &pb.GetProductRequest{Id: id}
    resp, err := h.client.GetProduct(context.Background(), req)
    if err != nil {
        st, ok := status.FromError(err)
        if ok {
            switch st.Code() {
            case codes.NotFound:
                c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
            case codes.InvalidArgument:
                c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
            default:
                c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
            }
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
    pageStr := c.DefaultQuery("page", "1")
    limitStr := c.DefaultQuery("limit", "10")

    page, err := strconv.Atoi(pageStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
        return
    }

    limit, err := strconv.Atoi(limitStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit number"})
        return
    }

    req := &pb.ListProductsRequest{
        Page:  int32(page),
        Limit: int32(limit),
    }

    resp, err := h.client.ListProducts(context.Background(), req)
    if err != nil {
        h.logger.Error("Failed to list products", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
    var req pb.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.client.CreateProduct(context.Background(), &req)
    if err != nil {
        st, ok := status.FromError(err)
        if ok && st.Code() == codes.InvalidArgument {
            c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        return
    }

    c.JSON(http.StatusCreated, resp)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    id := c.Param("id")
    var req pb.UpdateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    req.Id = id

    resp, err := h.client.UpdateProduct(context.Background(), &req)
    if err != nil {
        st, ok := status.FromError(err)
        if ok {
            switch st.Code() {
            case codes.NotFound:
                c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
            case codes.InvalidArgument:
                c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
            default:
                c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
            }
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
    id := c.Param("id")
    req := &pb.DeleteProductRequest{Id: id}

    _, err := h.client.DeleteProduct(context.Background(), req)
    if err != nil {
        st, ok := status.FromError(err)
        if ok && st.Code() == codes.NotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        return
    }

    c.Status(http.StatusNoContent)
}
