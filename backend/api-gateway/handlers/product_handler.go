package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/louai60/e-commerce_project/backend/api-gateway/formatters"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

type ProductHandler struct {
	client pb.ProductServiceClient
	logger *zap.Logger
}

func NewProductHandler(client pb.ProductServiceClient, logger *zap.Logger) *ProductHandler {
	if client == nil {
		logger.Warn("Initializing ProductHandler with nil client - some functionality will be unavailable")
	}
	return &ProductHandler{
		client: client,
		logger: logger,
	}
}

// ImageUploadRequest represents the request body for image upload
type ImageUploadRequest struct {
	Folder   string `json:"folder" binding:"required"`
	AltText  string `json:"alt_text"`
	Position int32  `json:"position"`
}

// UploadImage handles image upload to Cloudinary
func (h *ProductHandler) UploadImage(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

	// Get the file from the form data
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Get form fields
	folder := c.DefaultPostForm("folder", "products")
	altText := c.PostForm("alt_text")
	positionStr := c.PostForm("position")
	position := int32(1) // Default position
	if positionStr != "" {
		pos, err := strconv.Atoi(positionStr)
		if err == nil {
			position = int32(pos)
		}
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		h.logger.Error("Failed to open file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process file"})
		return
	}
	defer src.Close()

	// Read the file content
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		h.logger.Error("Failed to read file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	// Create the gRPC request
	grpcReq := &pb.UploadImageRequest{
		File:     fileBytes,
		Folder:   folder,
		AltText:  altText,
		Position: position,
		Filename: file.Filename,
		MimeType: file.Header.Get("Content-Type"),
	}

	// Call the product service
	resp, err := h.client.UploadImage(c.Request.Context(), grpcReq)
	if err != nil {
		h.handleGRPCError(c, err, "Failed to upload image")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":       resp.Url,
		"public_id": resp.PublicId,
		"alt_text":  resp.AltText,
		"position":  resp.Position,
	})
}

// DeleteImage handles image deletion from Cloudinary
func (h *ProductHandler) DeleteImage(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

	publicID := c.Param("public_id")
	if publicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "public_id is required"})
		return
	}

	// Create the gRPC request
	grpcReq := &pb.DeleteImageRequest{
		PublicId: publicID,
	}

	// Call the product service
	resp, err := h.client.DeleteImage(c.Request.Context(), grpcReq)
	if err != nil {
		h.handleGRPCError(c, err, "Failed to delete image")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
	})
}

// GetProduct handles retrieving a product by ID or slug
func (h *ProductHandler) GetProduct(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

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
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

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
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

	// Log the raw request body for debugging
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Restore the request body for further processing
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Log the raw request
	h.logger.Info("Received product creation request", zap.String("body", string(bodyBytes)))

	var req pb.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err), zap.String("body", string(bodyBytes)))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that Product is not nil
	if req.Product == nil {
		h.logger.Error("Product data is nil")
		c.JSON(http.StatusBadRequest, gin.H{"error": "product data is required"})
		return
	}

	// Handle image uploads if present
	if req.Product.Images != nil && len(req.Product.Images) > 0 {
		for i, img := range req.Product.Images {
			if img == nil {
				continue
			}

			if img.Url != "" {
				// Image already has a URL, skip upload
				continue
			}

			// Get the file from the form data
			file, err := c.FormFile("images[" + strconv.Itoa(i) + "]")
			if err != nil {
				h.logger.Error("Failed to get file from form", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get image file"})
				return
			}

			// Open the file
			src, err := file.Open()
			if err != nil {
				h.logger.Error("Failed to open file", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process file"})
				return
			}
			defer src.Close()

			// Read the file content
			fileBytes, err := io.ReadAll(src)
			if err != nil {
				h.logger.Error("Failed to read file", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
				return
			}

			// Create the gRPC request for image upload
			uploadReq := &pb.UploadImageRequest{
				File:     fileBytes,
				Folder:   "products",
				AltText:  img.AltText,
				Position: img.Position,
				Filename: file.Filename,
				MimeType: file.Header.Get("Content-Type"),
			}

			// Upload the image
			uploadResp, err := h.client.UploadImage(c.Request.Context(), uploadReq)
			if err != nil {
				h.logger.Error("Failed to upload image", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image"})
				return
			}

			// Update the image URL
			img.Url = uploadResp.Url
		}
	}

	// Create the product
	h.logger.Info("Sending product creation request to product service",
		zap.Any("product", req.Product))

	resp, err := h.client.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))
		h.handleGRPCError(c, err, "Failed to create product")
		return
	}

	// Format the response
	formattedProduct := formatters.FormatProduct(resp)
	c.JSON(http.StatusCreated, formattedProduct)
}

// UpdateProduct handles updating an existing product
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

	var req pb.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle new image uploads if present
	if len(req.Product.Images) > 0 {
		for i, img := range req.Product.Images {
			if img.Url != "" {
				// Image already has a URL, skip upload
				continue
			}

			// Get the file from the form data
			file, err := c.FormFile("images[" + strconv.Itoa(i) + "]")
			if err != nil {
				h.logger.Error("Failed to get file from form", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get image file"})
				return
			}

			// Open the file
			src, err := file.Open()
			if err != nil {
				h.logger.Error("Failed to open file", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process file"})
				return
			}
			defer src.Close()

			// Read the file content
			fileBytes, err := io.ReadAll(src)
			if err != nil {
				h.logger.Error("Failed to read file", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
				return
			}

			// Create the gRPC request for image upload
			uploadReq := &pb.UploadImageRequest{
				File:     fileBytes,
				Folder:   "products",
				AltText:  img.AltText,
				Position: img.Position,
				Filename: file.Filename,
				MimeType: file.Header.Get("Content-Type"),
			}

			// Upload the image
			uploadResp, err := h.client.UploadImage(c.Request.Context(), uploadReq)
			if err != nil {
				h.logger.Error("Failed to upload image", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image"})
				return
			}

			// Update the image URL
			req.Product.Images[i].Url = uploadResp.Url
		}
	}

	// Update the product
	resp, err := h.client.UpdateProduct(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteProduct handles deleting a product
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

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
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

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

	// Handle nil response
	if resp == nil {
		h.logger.Error("Received nil response from GetBrand")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Format the response
	formattedBrand := formatters.FormatBrand(resp)
	c.JSON(http.StatusOK, formattedBrand)
}

// ListBrands handles retrieving a paginated list of brands
func (h *ProductHandler) ListBrands(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

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

	// Handle nil response
	if resp == nil {
		h.logger.Error("Received nil response from ListBrands")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Use the formatter to safely handle the response
	brands := make([]*pb.Brand, 0)
	if resp.Brands != nil {
		brands = resp.Brands
	}

	total := 0
	if resp.Total > 0 {
		total = int(resp.Total)
	}

	formattedResponse := formatters.FormatBrandList(brands, page, limit, total)
	c.JSON(http.StatusOK, formattedResponse)
}

// CreateBrand handles creating a new brand
func (h *ProductHandler) CreateBrand(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

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

	// Handle nil response
	if resp == nil {
		h.logger.Error("Received nil response from CreateBrand")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Format the response
	formattedBrand := formatters.FormatBrand(resp)
	c.JSON(http.StatusCreated, formattedBrand)
}

// GetCategory handles retrieving a category by ID or slug
func (h *ProductHandler) GetCategory(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

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

	// Handle nil response
	if resp == nil {
		h.logger.Error("Received nil response from GetCategory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Format the response
	formattedCategory := formatters.FormatCategory(resp)
	c.JSON(http.StatusOK, formattedCategory)
}

// ListCategories handles retrieving a paginated list of categories
func (h *ProductHandler) ListCategories(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

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

	// Handle nil response
	if resp == nil {
		h.logger.Error("Received nil response from ListCategories")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Use the formatter to safely handle the response
	categories := make([]*pb.Category, 0)
	if resp.Categories != nil {
		categories = resp.Categories
	}

	total := 0
	if resp.Total > 0 {
		total = int(resp.Total)
	}

	formattedResponse := formatters.FormatCategoryList(categories, page, limit, total)
	c.JSON(http.StatusOK, formattedResponse)
}

// CategoryRequest represents the JSON structure for category creation/update
type CategoryRequest struct {
	Category struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		ParentID    string `json:"parent_id"`
		ParentName  string `json:"parent_name"`
	} `json:"category"`
}

// CreateCategory handles creating a new category
func (h *ProductHandler) CreateCategory(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

	role, exists := c.Get("user_role")
	if !exists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	// Parse the request using our custom struct first
	var categoryReq CategoryRequest
	if err := c.ShouldBindJSON(&categoryReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the protobuf request
	protoCategory := &pb.Category{
		Name:        categoryReq.Category.Name,
		Slug:        categoryReq.Category.Slug,
		Description: categoryReq.Category.Description,
		ParentName:  categoryReq.Category.ParentName,
	}

	// Handle parent_id properly by converting to wrapperspb.StringValue
	if categoryReq.Category.ParentID != "" {
		protoCategory.ParentId = &wrapperspb.StringValue{Value: categoryReq.Category.ParentID}
	}

	// Create the final request
	req := &pb.CreateCategoryRequest{
		Category: protoCategory,
	}

	resp, err := h.client.CreateCategory(context.Background(), req)
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

	// Handle nil response
	if resp == nil {
		h.logger.Error("Received nil response from CreateCategory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Format the response
	formattedCategory := formatters.FormatCategory(resp)
	c.JSON(http.StatusCreated, formattedCategory)
}

// handleGRPCError handles gRPC errors and returns appropriate HTTP responses
func (h *ProductHandler) handleGRPCError(c *gin.Context, err error, message string) {
	// Check if client is nil (this shouldn't happen if we check in each handler, but just in case)
	if h.client == nil {
		h.logger.Error("Product service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product service unavailable"})
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		h.logger.Error(message, zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": message + ": " + err.Error()})
		return
	}

	// Log the detailed error
	h.logger.Error(message,
		zap.String("code", st.Code().String()),
		zap.String("message", st.Message()),
		zap.Any("details", st.Details()),
	)

	switch st.Code() {
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	case codes.PermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": message + ": " + st.Message()})
	}
}
