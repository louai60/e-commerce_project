package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/louai60/e-commerce_project/backend/api-gateway/clients"
	"github.com/louai60/e-commerce_project/backend/api-gateway/formatters"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// GetClient returns the product service client
func (h *ProductHandler) GetClient() pb.ProductServiceClient {
	return h.client
}

// GetLogger returns the logger
func (h *ProductHandler) GetLogger() *zap.Logger {
	return h.logger
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

	// Format the product
	formattedProduct := formatters.FormatProduct(resp)

	// Try to fetch inventory data for the product
	inventoryClient, exists := c.Get("inventory_client")
	if exists && inventoryClient != nil {
		invClient, ok := inventoryClient.(*clients.InventoryClient)
		if ok {
			// Add a delay to ensure inventory data is available
			// This helps with eventual consistency between services
			time.Sleep(500 * time.Millisecond)

			// Fetch inventory data
			inventoryItem, err := invClient.GetInventoryItem(c.Request.Context(), resp.Id)
			if err == nil && inventoryItem != nil {
				h.logger.Info("Successfully fetched inventory data",
					zap.String("product_id", resp.Id),
					zap.Int("total_quantity", int(inventoryItem.TotalQuantity)),
					zap.Int("available_quantity", int(inventoryItem.AvailableQuantity)),
					zap.Int("reserved_quantity", int(inventoryItem.ReservedQuantity)),
					zap.String("status", inventoryItem.Status))

				// Update the inventory data in the response with comprehensive information
				formattedProduct.Inventory = &formatters.EnhancedInventoryInfo{
					Status:            inventoryItem.Status,
					Available:         inventoryItem.AvailableQuantity > 0,
					Quantity:          int(inventoryItem.AvailableQuantity), // For backward compatibility
					TotalQuantity:     int(inventoryItem.TotalQuantity),
					AvailableQuantity: int(inventoryItem.AvailableQuantity),
					ReservedQuantity:  int(inventoryItem.ReservedQuantity),
					ReorderPoint:      int(inventoryItem.ReorderPoint),
					ReorderQuantity:   int(inventoryItem.ReorderQuantity),
					LastUpdated:       formatTimestamp(inventoryItem.LastUpdated),
				}

				// Add location data if available
				if len(inventoryItem.Locations) > 0 {
					locations := make([]formatters.EnhancedLocationInfo, len(inventoryItem.Locations))
					for i, loc := range inventoryItem.Locations {
						locations[i] = formatters.EnhancedLocationInfo{
							WarehouseID: loc.WarehouseId,
							Quantity:    int(loc.Quantity),
						}
					}
					formattedProduct.Inventory.Locations = locations
				}
			} else {
				h.logger.Warn("Failed to fetch inventory data for product",
					zap.Error(err),
					zap.String("product_id", resp.Id))
			}
		}
	}

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
	debugStr := c.DefaultQuery("debug", "false")
	debug := debugStr == "true"

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

	// Log that we're retrieving products
	h.logger.Info("Retrieving product list", zap.Int("page", page), zap.Int("limit", limit))

	resp, err := h.client.ListProducts(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to list products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Log the number of products retrieved
	h.logger.Info("Retrieved products", zap.Int("count", len(resp.Products)), zap.Int32("total", resp.Total))

	// Calculate total pages based on the total count
	totalPages := (int(resp.Total) + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	// Add debug information if requested
	if debug {
		h.logger.Info("Debug mode enabled for product listing")
		debugInfo := map[string]interface{}{
			"requested_page":     page,
			"requested_limit":    limit,
			"actual_items_count": len(resp.Products),
			"reported_total":     resp.Total,
			"calculated_pages":   totalPages,
			"pagination_metadata": map[string]interface{}{
				"current_page": page,
				"total_pages":  totalPages,
				"per_page":     limit,
				"total_items":  resp.Total,
			},
		}
		c.JSON(http.StatusOK, debugInfo)
		return
	}

	// Format the response
	formattedResponse := formatters.FormatProductList(resp.Products, page, limit, int(resp.Total))

	// Try to fetch inventory data for each product
	inventoryClient, exists := c.Get("inventory_client")
	if exists && inventoryClient != nil {
		invClient, ok := inventoryClient.(*clients.InventoryClient)
		if ok {
			// Add a delay to ensure inventory data is available
			// This helps with eventual consistency between services
			time.Sleep(500 * time.Millisecond)

			for i, product := range formattedResponse.Products {
				// Fetch inventory data
				inventoryItem, err := invClient.GetInventoryItem(c.Request.Context(), product.ID)
				if err == nil && inventoryItem != nil {
					h.logger.Info("Successfully fetched inventory data for product in list",
						zap.String("product_id", product.ID),
						zap.Int("total_quantity", int(inventoryItem.TotalQuantity)),
						zap.Int("available_quantity", int(inventoryItem.AvailableQuantity)),
						zap.Int("reserved_quantity", int(inventoryItem.ReservedQuantity)),
						zap.String("status", inventoryItem.Status))

					// Update the inventory data in the response with comprehensive information
					formattedResponse.Products[i].Inventory = &formatters.EnhancedInventoryInfo{
						Status:            inventoryItem.Status,
						Available:         inventoryItem.AvailableQuantity > 0,
						Quantity:          int(inventoryItem.AvailableQuantity), // For backward compatibility
						TotalQuantity:     int(inventoryItem.TotalQuantity),
						AvailableQuantity: int(inventoryItem.AvailableQuantity),
						ReservedQuantity:  int(inventoryItem.ReservedQuantity),
						ReorderPoint:      int(inventoryItem.ReorderPoint),
						ReorderQuantity:   int(inventoryItem.ReorderQuantity),
						LastUpdated:       formatTimestamp(inventoryItem.LastUpdated),
					}

					// Add location data if available
					if len(inventoryItem.Locations) > 0 {
						locations := make([]formatters.EnhancedLocationInfo, len(inventoryItem.Locations))
						for j, loc := range inventoryItem.Locations {
							locations[j] = formatters.EnhancedLocationInfo{
								WarehouseID: loc.WarehouseId,
								Quantity:    int(loc.Quantity),
							}
						}
						formattedResponse.Products[i].Inventory.Locations = locations
					}
				} else {
					h.logger.Warn("Failed to fetch inventory data for product in list",
						zap.Error(err),
						zap.String("product_id", product.ID))
				}
			}
		}
	}

	// Log the number of formatted products
	h.logger.Info("Formatted products", zap.Int("count", len(formattedResponse.Products)))

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

	var req struct {
		Product struct {
			Title            string                           `json:"title" binding:"required"`
			Slug             string                           `json:"slug" binding:"required"`
			Description      string                           `json:"description" binding:"required"`
			ShortDescription string                           `json:"short_description"`
			Price            float64                          `json:"price" binding:"required"`
			DiscountPrice    *float64                         `json:"discount_price,omitempty"`
			SKU              string                           `json:"sku" binding:"required"`
			InventoryQty     int                              `json:"inventory_qty,omitempty"`
			Weight           *float64                         `json:"weight,omitempty"`
			IsPublished      bool                             `json:"is_published"`
			BrandID          *string                          `json:"brand_id,omitempty"`
			Images           []formatters.EnhancedImageInfo   `json:"images,omitempty"`
			Categories       []formatters.CategoryInfo        `json:"categories,omitempty"`
			Variants         []formatters.EnhancedVariantInfo `json:"variants,omitempty"`
			Tags             []string                         `json:"tags,omitempty"`
			Attributes       []formatters.AttributeInfo       `json:"attributes,omitempty"`
			Specifications   []formatters.SpecificationInfo   `json:"specifications,omitempty"`
			SEO              *formatters.EnhancedSEOInfo      `json:"seo,omitempty"`
			Shipping         *formatters.EnhancedShippingInfo `json:"shipping,omitempty"`
			Discount         *formatters.DiscountInfo         `json:"discount,omitempty"`
			Inventory        *struct {
				InitialQuantity int `json:"initial_quantity"`
			} `json:"inventory,omitempty"`
		} `json:"product" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request to proto message
	product := &pb.Product{
		Title:            req.Product.Title,
		Slug:             req.Product.Slug,
		Description:      req.Product.Description,
		ShortDescription: req.Product.ShortDescription,
		Price:            req.Product.Price,
		Sku:              req.Product.SKU,
		IsPublished:      req.Product.IsPublished,
	}

	// Set optional fields
	if req.Product.DiscountPrice != nil {
		product.DiscountPrice = &wrapperspb.DoubleValue{Value: *req.Product.DiscountPrice}
	}
	if req.Product.Weight != nil {
		product.Weight = &wrapperspb.DoubleValue{Value: *req.Product.Weight}
	}
	if req.Product.BrandID != nil {
		product.BrandId = &wrapperspb.StringValue{Value: *req.Product.BrandID}
	}

	// Convert variants
	if len(req.Product.Variants) > 0 {
		product.Variants = make([]*pb.ProductVariant, len(req.Product.Variants))
		for i, variant := range req.Product.Variants {
			product.Variants[i] = &pb.ProductVariant{
				Sku:              variant.SKU,
				Title:            variant.Title,
				Price:            variant.Price,
				Description:      variant.Description,
				ShortDescription: variant.ShortDescription,
			}

			// Set optional fields
			if variant.DiscountPrice != 0 {
				product.Variants[i].DiscountPrice = &wrapperspb.DoubleValue{Value: variant.DiscountPrice}
			}

			// Convert variant attributes
			if len(variant.Attributes) > 0 {
				product.Variants[i].Attributes = make([]*pb.VariantAttributeValue, len(variant.Attributes))
				for j, attr := range variant.Attributes {
					product.Variants[i].Attributes[j] = &pb.VariantAttributeValue{
						Name:  attr.Name,
						Value: attr.Value,
					}
				}
			}

			// Convert variant images
			if len(variant.Images) > 0 {
				product.Variants[i].Images = make([]*pb.VariantImage, len(variant.Images))
				for j, img := range variant.Images {
					product.Variants[i].Images[j] = &pb.VariantImage{
						Url:      img.URL,
						AltText:  img.AltText,
						Position: int32(img.Position),
					}
				}
			}
		}
	}

	// Convert images
	if len(req.Product.Images) > 0 {
		product.Images = make([]*pb.ProductImage, len(req.Product.Images))
		for i, img := range req.Product.Images {
			product.Images[i] = &pb.ProductImage{
				Url:      img.URL,
				AltText:  img.AltText,
				Position: int32(img.Position),
			}
		}
	}

	// Convert categories
	if len(req.Product.Categories) > 0 {
		product.Categories = make([]*pb.Category, len(req.Product.Categories))
		for i, cat := range req.Product.Categories {
			product.Categories[i] = &pb.Category{
				Id:   cat.ID,
				Name: cat.Name,
				Slug: cat.Slug,
			}
		}
	}

	// Convert tags
	if len(req.Product.Tags) > 0 {
		product.Tags = make([]*pb.ProductTag, len(req.Product.Tags))
		for i, tag := range req.Product.Tags {
			product.Tags[i] = &pb.ProductTag{
				Tag: tag,
			}
		}
	}

	// Convert specifications
	if len(req.Product.Specifications) > 0 {
		product.Specifications = make([]*pb.ProductSpecification, len(req.Product.Specifications))
		for i, spec := range req.Product.Specifications {
			product.Specifications[i] = &pb.ProductSpecification{
				Name:  spec.Name,
				Value: spec.Value,
				Unit:  spec.Unit,
			}
		}
	}

	// Convert SEO
	if req.Product.SEO != nil {
		product.Seo = &pb.ProductSEO{
			MetaTitle:       req.Product.SEO.MetaTitle,
			MetaDescription: req.Product.SEO.MetaDescription,
			Keywords:        req.Product.SEO.Keywords,
			Tags:            req.Product.SEO.MetaTags,
		}
	}

	// Convert shipping
	if req.Product.Shipping != nil {
		estimatedDays, err := strconv.Atoi(req.Product.Shipping.EstimatedDays)
		if err != nil {
			estimatedDays = 0 // Default value if conversion fails
		}
		product.Shipping = &pb.ProductShipping{
			FreeShipping:     req.Product.Shipping.FreeShipping,
			EstimatedDays:    int32(estimatedDays),
			ExpressAvailable: req.Product.Shipping.ExpressShippingAvailable,
		}
	}

	// Convert discount
	if req.Product.Discount != nil {
		product.Discount = &pb.ProductDiscount{
			Type:  req.Product.Discount.Type,
			Value: req.Product.Discount.Value,
		}
		if req.Product.Discount.ExpiresAt != "" {
			if t, err := time.Parse(time.RFC3339, req.Product.Discount.ExpiresAt); err == nil {
				product.Discount.ExpiresAt = timestamppb.New(t)
			}
		}
	}

	// Create the product
	grpcReq := &pb.CreateProductRequest{
		Product: product,
	}

	resp, err := h.client.CreateProduct(c.Request.Context(), grpcReq)
	if err != nil {
		h.handleGRPCError(c, err, "Failed to create product")
		return
	}

	// Note: Inventory creation is now handled by product_inventory_handler.go
	// to avoid duplicate requests

	// Format the product
	formattedProduct := formatters.FormatProduct(resp)

	// Try to fetch inventory data for the product
	inventoryClient, exists := c.Get("inventory_client")
	if exists && inventoryClient != nil {
		invClient, ok := inventoryClient.(*clients.InventoryClient)
		if ok {
			// Add a delay to ensure inventory data is available
			// This helps with eventual consistency between services
			time.Sleep(500 * time.Millisecond)

			// Fetch inventory data
			inventoryItem, err := invClient.GetInventoryItem(c.Request.Context(), resp.Id)
			if err == nil && inventoryItem != nil {
				h.logger.Info("Successfully fetched inventory data",
					zap.String("product_id", resp.Id),
					zap.Int("total_quantity", int(inventoryItem.TotalQuantity)),
					zap.Int("available_quantity", int(inventoryItem.AvailableQuantity)),
					zap.Int("reserved_quantity", int(inventoryItem.ReservedQuantity)),
					zap.String("status", inventoryItem.Status))

				// Update the inventory data in the response with comprehensive information
				formattedProduct.Inventory = &formatters.EnhancedInventoryInfo{
					Status:            inventoryItem.Status,
					Available:         inventoryItem.AvailableQuantity > 0,
					Quantity:          int(inventoryItem.AvailableQuantity), // For backward compatibility
					TotalQuantity:     int(inventoryItem.TotalQuantity),
					AvailableQuantity: int(inventoryItem.AvailableQuantity),
					ReservedQuantity:  int(inventoryItem.ReservedQuantity),
					ReorderPoint:      int(inventoryItem.ReorderPoint),
					ReorderQuantity:   int(inventoryItem.ReorderQuantity),
					LastUpdated:       formatTimestamp(inventoryItem.LastUpdated),
				}

				// Add location data if available
				if len(inventoryItem.Locations) > 0 {
					locations := make([]formatters.EnhancedLocationInfo, len(inventoryItem.Locations))
					for i, loc := range inventoryItem.Locations {
						locations[i] = formatters.EnhancedLocationInfo{
							WarehouseID: loc.WarehouseId,
							Quantity:    int(loc.Quantity),
						}
					}
					formattedProduct.Inventory.Locations = locations
				}
			} else {
				h.logger.Warn("Failed to fetch inventory data for product",
					zap.Error(err),
					zap.String("product_id", resp.Id))

				// If we can't fetch the inventory data but we know inventory was requested,
				// provide a default inventory object with the initial quantity
				if req.Product.Inventory != nil {
					initialQty := req.Product.Inventory.InitialQuantity
					formattedProduct.Inventory = &formatters.EnhancedInventoryInfo{
						Status:            "IN_STOCK",
						Available:         initialQty > 0,
						Quantity:          initialQty, // For backward compatibility
						TotalQuantity:     initialQty,
						AvailableQuantity: initialQty,
						ReservedQuantity:  0,
						ReorderPoint:      5,  // Default reorder point
						ReorderQuantity:   20, // Default reorder quantity
						LastUpdated:       time.Now().Format(time.RFC3339),
					}
				}
			}
		}
	}

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

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	var req struct {
		Product struct {
			Title            string                           `json:"title,omitempty"`
			Slug             string                           `json:"slug,omitempty"`
			Description      string                           `json:"description,omitempty"`
			ShortDescription string                           `json:"short_description,omitempty"`
			Price            float64                          `json:"price,omitempty"`
			DiscountPrice    *float64                         `json:"discount_price,omitempty"`
			SKU              string                           `json:"sku,omitempty"`
			Weight           *float64                         `json:"weight,omitempty"`
			IsPublished      bool                             `json:"is_published,omitempty"`
			BrandID          *string                          `json:"brand_id,omitempty"`
			Images           []formatters.EnhancedImageInfo   `json:"images,omitempty"`
			Categories       []formatters.CategoryInfo        `json:"categories,omitempty"`
			Variants         []formatters.EnhancedVariantInfo `json:"variants,omitempty"`
			Tags             []string                         `json:"tags,omitempty"`
			Attributes       []formatters.AttributeInfo       `json:"attributes,omitempty"`
			Specifications   []formatters.SpecificationInfo   `json:"specifications,omitempty"`
			SEO              *formatters.EnhancedSEOInfo      `json:"seo,omitempty"`
			Shipping         *formatters.EnhancedShippingInfo `json:"shipping,omitempty"`
			Discount         *formatters.DiscountInfo         `json:"discount,omitempty"`
		} `json:"product" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request to proto message
	product := &pb.Product{
		Id: id,
	}

	// Set fields that are present in the request
	if req.Product.Title != "" {
		product.Title = req.Product.Title
	}
	if req.Product.Slug != "" {
		product.Slug = req.Product.Slug
	}
	if req.Product.Description != "" {
		product.Description = req.Product.Description
	}
	if req.Product.ShortDescription != "" {
		product.ShortDescription = req.Product.ShortDescription
	}
	if req.Product.Price != 0 {
		product.Price = req.Product.Price
	}
	if req.Product.SKU != "" {
		product.Sku = req.Product.SKU
	}
	if req.Product.IsPublished {
		product.IsPublished = req.Product.IsPublished
	}

	// Set optional fields
	if req.Product.DiscountPrice != nil {
		product.DiscountPrice = &wrapperspb.DoubleValue{Value: *req.Product.DiscountPrice}
	}
	if req.Product.Weight != nil {
		product.Weight = &wrapperspb.DoubleValue{Value: *req.Product.Weight}
	}
	if req.Product.BrandID != nil {
		product.BrandId = &wrapperspb.StringValue{Value: *req.Product.BrandID}
	}

	// Convert variants
	if len(req.Product.Variants) > 0 {
		product.Variants = make([]*pb.ProductVariant, len(req.Product.Variants))
		for i, variant := range req.Product.Variants {
			product.Variants[i] = &pb.ProductVariant{
				Sku:              variant.SKU,
				Title:            variant.Title,
				Price:            variant.Price,
				Description:      variant.Description,
				ShortDescription: variant.ShortDescription,
			}

			// Set optional fields
			if variant.DiscountPrice != 0 {
				product.Variants[i].DiscountPrice = &wrapperspb.DoubleValue{Value: variant.DiscountPrice}
			}

			// Convert variant attributes
			if len(variant.Attributes) > 0 {
				product.Variants[i].Attributes = make([]*pb.VariantAttributeValue, len(variant.Attributes))
				for j, attr := range variant.Attributes {
					product.Variants[i].Attributes[j] = &pb.VariantAttributeValue{
						Name:  attr.Name,
						Value: attr.Value,
					}
				}
			}

			// Convert variant images
			if len(variant.Images) > 0 {
				product.Variants[i].Images = make([]*pb.VariantImage, len(variant.Images))
				for j, img := range variant.Images {
					product.Variants[i].Images[j] = &pb.VariantImage{
						Url:      img.URL,
						AltText:  img.AltText,
						Position: int32(img.Position),
					}
				}
			}
		}
	}

	// Convert images
	if len(req.Product.Images) > 0 {
		product.Images = make([]*pb.ProductImage, len(req.Product.Images))
		for i, img := range req.Product.Images {
			product.Images[i] = &pb.ProductImage{
				Url:      img.URL,
				AltText:  img.AltText,
				Position: int32(img.Position),
			}
		}
	}

	// Convert categories
	if len(req.Product.Categories) > 0 {
		product.Categories = make([]*pb.Category, len(req.Product.Categories))
		for i, cat := range req.Product.Categories {
			product.Categories[i] = &pb.Category{
				Id:   cat.ID,
				Name: cat.Name,
				Slug: cat.Slug,
			}
		}
	}

	// Convert tags
	if len(req.Product.Tags) > 0 {
		product.Tags = make([]*pb.ProductTag, len(req.Product.Tags))
		for i, tag := range req.Product.Tags {
			product.Tags[i] = &pb.ProductTag{
				Tag: tag,
			}
		}
	}

	// Convert specifications
	if len(req.Product.Specifications) > 0 {
		product.Specifications = make([]*pb.ProductSpecification, len(req.Product.Specifications))
		for i, spec := range req.Product.Specifications {
			product.Specifications[i] = &pb.ProductSpecification{
				Name:  spec.Name,
				Value: spec.Value,
				Unit:  spec.Unit,
			}
		}
	}

	// Convert SEO
	if req.Product.SEO != nil {
		product.Seo = &pb.ProductSEO{
			MetaTitle:       req.Product.SEO.MetaTitle,
			MetaDescription: req.Product.SEO.MetaDescription,
			Keywords:        req.Product.SEO.Keywords,
			Tags:            req.Product.SEO.MetaTags,
		}
	}

	// Convert shipping
	if req.Product.Shipping != nil {
		estimatedDays, err := strconv.Atoi(req.Product.Shipping.EstimatedDays)
		if err != nil {
			estimatedDays = 0 // Default value if conversion fails
		}
		product.Shipping = &pb.ProductShipping{
			FreeShipping:     req.Product.Shipping.FreeShipping,
			EstimatedDays:    int32(estimatedDays),
			ExpressAvailable: req.Product.Shipping.ExpressShippingAvailable,
		}
	}

	// Convert discount
	if req.Product.Discount != nil {
		product.Discount = &pb.ProductDiscount{
			Type:  req.Product.Discount.Type,
			Value: req.Product.Discount.Value,
		}
		if req.Product.Discount.ExpiresAt != "" {
			if t, err := time.Parse(time.RFC3339, req.Product.Discount.ExpiresAt); err == nil {
				product.Discount.ExpiresAt = timestamppb.New(t)
			}
		}
	}

	// Update the product
	grpcReq := &pb.UpdateProductRequest{
		Product: product,
	}

	resp, err := h.client.UpdateProduct(c.Request.Context(), grpcReq)
	if err != nil {
		h.handleGRPCError(c, err, "Failed to update product")
		return
	}

	formattedProduct := formatters.FormatProduct(resp)
	c.JSON(http.StatusOK, formattedProduct)
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
