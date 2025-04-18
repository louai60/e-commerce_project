package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/louai60/e-commerce_project/backend/api-gateway/formatters"
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

// GetProduct handles retrieving a product by ID or slug
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	req := &pb.GetProductRequest{
		Identifier: &pb.GetProductRequest_Id{
			Id: id,
		},
	}

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

	formattedProduct := formatters.FormatProduct(resp)
	// Wrap in a products array for consistent response format
	response := formatters.ProductListResponse{
		Products: []formatters.ProductResponse{formattedProduct},
		Total:    1,
		Pagination: formatters.PaginationInfo{
			CurrentPage: 1,
			TotalPages:  1,
			PerPage:     1,
			TotalItems:  1,
		},
	}
	c.JSON(http.StatusOK, response)
}

// ListProducts handles retrieving a paginated list of products
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

	formattedResponse := formatters.FormatProductList(resp.Products, page, limit, int(resp.Total))
	c.JSON(http.StatusOK, formattedResponse)
}

// CreateProduct handles creating a new product
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	role, exists := c.Get("user_role")
	if !exists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var req pb.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	resp, err := h.client.CreateProduct(ctx, &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			return
		}
		h.logger.Error("Failed to create product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	formattedProduct := formatters.FormatProduct(resp)
	// Wrap in a products array for consistent response format
	response := formatters.ProductListResponse{
		Products: []formatters.ProductResponse{formattedProduct},
		Total:    1,
		Pagination: formatters.PaginationInfo{
			CurrentPage: 1,
			TotalPages:  1,
			PerPage:     1,
			TotalItems:  1,
		},
	}
	c.JSON(http.StatusCreated, response)
}

// UpdateProduct handles updating an existing product
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	role, exists := c.Get("user_role")
	if !exists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var req pb.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	resp, err := h.client.UpdateProduct(ctx, &req)
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

	formattedProduct := formatters.FormatProduct(resp)
	// Wrap in a products array for consistent response format
	response := formatters.ProductListResponse{
		Products: []formatters.ProductResponse{formattedProduct},
		Total:    1,
		Pagination: formatters.PaginationInfo{
			CurrentPage: 1,
			TotalPages:  1,
			PerPage:     1,
			TotalItems:  1,
		},
	}
	c.JSON(http.StatusOK, response)
}

// DeleteProduct handles deleting a product
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	role, exists := c.Get("user_role")
	if !exists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	req := &pb.DeleteProductRequest{
		Id: id,
	}

	resp, err := h.client.DeleteProduct(context.Background(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// GetBrand handles retrieving a brand by ID or slug
func (h *ProductHandler) GetBrand(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "brand ID is required"})
		return
	}

	req := &pb.GetBrandRequest{
		Identifier: &pb.GetBrandRequest_Id{
			Id: id,
		},
	}

	resp, err := h.client.GetBrand(context.Background(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
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

// ListBrands handles retrieving a paginated list of brands
func (h *ProductHandler) ListBrands(c *gin.Context) {
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

	req := &pb.ListBrandsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	}

	resp, err := h.client.ListBrands(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to list brands", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"brands": resp.Brands,
		"total":  resp.Total,
		"page":   page,
		"limit":  limit,
	})
}

// CreateBrand handles creating a new brand
func (h *ProductHandler) CreateBrand(c *gin.Context) {
	role, exists := c.Get("user_role")
	if !exists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var req pb.CreateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.CreateBrand(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			return
		}
		h.logger.Error("Failed to create brand", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetCategory handles retrieving a category by ID or slug
func (h *ProductHandler) GetCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category ID is required"})
		return
	}

	req := &pb.GetCategoryRequest{
		Identifier: &pb.GetCategoryRequest_Id{
			Id: id,
		},
	}

	resp, err := h.client.GetCategory(context.Background(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
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

// ListCategories handles retrieving a paginated list of categories
func (h *ProductHandler) ListCategories(c *gin.Context) {
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

	req := &pb.ListCategoriesRequest{
		Page:  int32(page),
		Limit: int32(limit),
	}

	resp, err := h.client.ListCategories(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to list categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": resp.Categories,
		"total":      resp.Total,
		"page":       page,
		"limit":      limit,
	})
}

// CreateCategory handles creating a new category
func (h *ProductHandler) CreateCategory(c *gin.Context) {
	role, exists := c.Get("user_role")
	if !exists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var req pb.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.CreateCategory(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			return
		}
		h.logger.Error("Failed to create category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}
