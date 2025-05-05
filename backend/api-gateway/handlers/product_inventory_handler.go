package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/louai60/e-commerce_project/backend/api-gateway/clients"
	"github.com/louai60/e-commerce_project/backend/api-gateway/formatters"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	productpb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

// ProductInventoryHandler handles product creation with inventory integration
func CreateProductWithInventory(
	c *gin.Context,
	productClient productpb.ProductServiceClient,
	inventoryClient *clients.InventoryClient,
	logger *zap.Logger,
) {
	// Parse the request
	var req struct {
		Product struct {
			Title            string                   `json:"title" binding:"required"`
			Slug             string                   `json:"slug" binding:"required"`
			Description      string                   `json:"description" binding:"required"`
			ShortDescription string                   `json:"short_description"`
			Price            float64                  `json:"price" binding:"required"`
			DiscountPrice    *float64                 `json:"discount_price,omitempty"`
			SKU              string                   `json:"sku" binding:"required"`
			IsPublished      bool                     `json:"is_published"`
			Weight           *float64                 `json:"weight,omitempty"`
			BrandID          string                   `json:"brand_id,omitempty"`
			CategoryIDs      []string                 `json:"category_ids,omitempty"`
			Images           []map[string]interface{} `json:"images,omitempty"`
			Specifications   []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
				Unit  string `json:"unit"`
			} `json:"specifications,omitempty"`
			Tags       []string `json:"tags,omitempty"`
			Attributes []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"attributes,omitempty"`
			Variants []struct {
				Title         string   `json:"title"`
				SKU           string   `json:"sku"`
				Price         float64  `json:"price"`
				DiscountPrice *float64 `json:"discount_price,omitempty"`
				InventoryQty  int      `json:"inventory_qty"`
				Attributes    []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"attributes,omitempty"`
				Images []map[string]interface{} `json:"images,omitempty"`
			} `json:"variants,omitempty"`
			Inventory *struct {
				InitialQuantity int `json:"initial_quantity"`
			} `json:"inventory,omitempty"`
			SEO *struct {
				MetaTitle       string   `json:"meta_title"`
				MetaDescription string   `json:"meta_description"`
				Keywords        []string `json:"keywords"`
			} `json:"seo,omitempty"`
			Shipping *struct {
				FreeShipping     bool `json:"free_shipping"`
				EstimatedDays    int  `json:"estimated_days"`
				ExpressAvailable bool `json:"express_available"`
			} `json:"shipping,omitempty"`
		} `json:"product" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request to proto message for product service
	product := &productpb.Product{
		Title:            req.Product.Title,
		Slug:             req.Product.Slug,
		Description:      req.Product.Description,
		ShortDescription: req.Product.ShortDescription,
		Price:            req.Product.Price,
		Sku:              req.Product.SKU,
		IsPublished:      req.Product.IsPublished,
	}

	// Handle optional fields
	if req.Product.Weight != nil {
		product.Weight = &wrapperspb.DoubleValue{Value: *req.Product.Weight}
	}

	if req.Product.BrandID != "" {
		product.BrandId = &wrapperspb.StringValue{Value: req.Product.BrandID}
	}

	if req.Product.DiscountPrice != nil {
		product.DiscountPrice = &wrapperspb.DoubleValue{Value: *req.Product.DiscountPrice}
	}

	// Handle images
	if len(req.Product.Images) > 0 {
		product.Images = make([]*productpb.ProductImage, 0, len(req.Product.Images))
		for _, img := range req.Product.Images {
			url, _ := img["url"].(string)
			altText, _ := img["alt_text"].(string)
			position, ok := img["position"].(float64)
			if !ok {
				position = 0
			}

			if url != "" {
				product.Images = append(product.Images, &productpb.ProductImage{
					Url:      url,
					AltText:  altText,
					Position: int32(position),
				})
			}
		}
	}

	// Handle category IDs
	if len(req.Product.CategoryIDs) > 0 {
		// Convert string IDs to Category objects
		product.Categories = make([]*productpb.Category, len(req.Product.CategoryIDs))
		for i, id := range req.Product.CategoryIDs {
			// Create a complete Category object with the ID
			// The repository needs the full Category object with ID to create the association
			product.Categories[i] = &productpb.Category{
				Id: id,
			}

			// Log the category ID being processed for debugging
			logger.Info("Adding category to product",
				zap.String("category_id", id),
				zap.String("product_title", req.Product.Title))
		}
	}

	// Handle specifications
	if len(req.Product.Specifications) > 0 {
		product.Specifications = make([]*productpb.ProductSpecification, len(req.Product.Specifications))
		for i, spec := range req.Product.Specifications {
			product.Specifications[i] = &productpb.ProductSpecification{
				Name:  spec.Name,
				Value: spec.Value,
				Unit:  spec.Unit,
			}
		}
	}

	// Handle tags
	if len(req.Product.Tags) > 0 {
		product.Tags = make([]*productpb.ProductTag, len(req.Product.Tags))
		for i, tag := range req.Product.Tags {
			product.Tags[i] = &productpb.ProductTag{
				Tag: tag,
			}
		}
	}

	// Handle attributes
	if len(req.Product.Attributes) > 0 {
		product.Attributes = make([]*productpb.ProductAttribute, len(req.Product.Attributes))
		for i, attr := range req.Product.Attributes {
			product.Attributes[i] = &productpb.ProductAttribute{
				Name:  attr.Name,
				Value: attr.Value,
			}
		}
	}

	// Handle variants
	if len(req.Product.Variants) > 0 {
		product.Variants = make([]*productpb.ProductVariant, len(req.Product.Variants))
		for i, variant := range req.Product.Variants {
			productVariant := &productpb.ProductVariant{
				Title: variant.Title,
				Sku:   variant.SKU,
				Price: variant.Price,
				// Note: inventory_qty is not in the proto definition
				// It will be handled by the inventory service separately
			}

			// Handle variant discount price
			if variant.DiscountPrice != nil {
				productVariant.DiscountPrice = &wrapperspb.DoubleValue{Value: *variant.DiscountPrice}
			}

			// Handle variant attributes
			if len(variant.Attributes) > 0 {
				productVariant.Attributes = make([]*productpb.VariantAttributeValue, len(variant.Attributes))
				for j, attr := range variant.Attributes {
					productVariant.Attributes[j] = &productpb.VariantAttributeValue{
						Name:  attr.Name,
						Value: attr.Value,
					}
				}
			}

			// Handle variant images
			if len(variant.Images) > 0 {
				productVariant.Images = make([]*productpb.VariantImage, 0, len(variant.Images))
				for _, img := range variant.Images {
					url, _ := img["url"].(string)
					altText, _ := img["alt_text"].(string)
					position, ok := img["position"].(float64)
					if !ok {
						position = 0
					}

					if url != "" {
						productVariant.Images = append(productVariant.Images, &productpb.VariantImage{
							Url:      url,
							AltText:  altText,
							Position: int32(position),
						})
					}
				}
			}

			product.Variants[i] = productVariant
		}
	}

	// Handle SEO
	if req.Product.SEO != nil {
		product.Seo = &productpb.ProductSEO{
			MetaTitle:       req.Product.SEO.MetaTitle,
			MetaDescription: req.Product.SEO.MetaDescription,
		}

		if len(req.Product.SEO.Keywords) > 0 {
			product.Seo.Keywords = req.Product.SEO.Keywords
		}
	}

	// Handle Shipping
	if req.Product.Shipping != nil {
		product.Shipping = &productpb.ProductShipping{
			FreeShipping:     req.Product.Shipping.FreeShipping,
			EstimatedDays:    int32(req.Product.Shipping.EstimatedDays),
			ExpressAvailable: req.Product.Shipping.ExpressAvailable,
		}
	}

	// Create the product
	grpcReq := &productpb.CreateProductRequest{
		Product: product,
	}

	// Call the product service to create the product
	resp, err := productClient.CreateProduct(c.Request.Context(), grpcReq)
	if err != nil {
		handleGRPCError(c, err, "Failed to create product", logger)
		return
	}

	// If inventory data is provided and inventory client is available, create inventory item
	var inventoryCreated bool = false
	if req.Product.Inventory != nil && inventoryClient != nil {
		initialQty := req.Product.Inventory.InitialQuantity

		// Log the initial quantity to verify it's correct
		logger.Info("Creating inventory item with initial quantity",
			zap.Int("initial_quantity", initialQty),
			zap.String("product_id", resp.Id))

		// Default reorder points
		reorderPoint := 5
		reorderQty := 20

		// Create inventory item for the main product
		var variantID *string
		inventoryItem, err := inventoryClient.CreateInventoryItem(
			c.Request.Context(),
			resp.Id,
			resp.Sku,
			variantID, // Pass nil for variant ID since this is a main product
			initialQty,
			reorderPoint,
			reorderQty,
		)

		if err != nil {
			logger.Warn("Failed to create inventory item in inventory service",
				zap.Error(err),
				zap.String("product_id", resp.Id))
			// Continue even if inventory creation fails - the product was created successfully
		} else {
			logger.Info("Successfully created inventory item in inventory service",
				zap.String("product_id", resp.Id),
				zap.Int("available_quantity", int(inventoryItem.AvailableQuantity)))
			inventoryCreated = true
		}

		// Create inventory items for variants if any
		if len(resp.Variants) > 0 && inventoryClient != nil {
			for _, variant := range resp.Variants {
				// Find the corresponding variant in the request to get the inventory quantity
				var variantInventoryQty int = 0
				for _, reqVariant := range req.Product.Variants {
					if reqVariant.SKU == variant.Sku {
						variantInventoryQty = reqVariant.InventoryQty
						break
					}
				}

				// Create a string pointer for variant ID
				variantIDPtr := &variant.Id

				// Create inventory item for the variant
				variantInventoryItem, err := inventoryClient.CreateInventoryItem(
					c.Request.Context(),
					resp.Id,
					variant.Sku,
					variantIDPtr,
					variantInventoryQty,
					reorderPoint,
					reorderQty,
				)

				if err != nil {
					logger.Warn("Failed to create inventory item for variant in inventory service",
						zap.Error(err),
						zap.String("product_id", resp.Id),
						zap.String("variant_id", variant.Id),
						zap.String("variant_sku", variant.Sku))
					// Continue even if inventory creation fails
				} else {
					logger.Info("Successfully created inventory item for variant in inventory service",
						zap.String("product_id", resp.Id),
						zap.String("variant_id", variant.Id),
						zap.String("variant_sku", variant.Sku),
						zap.Int("available_quantity", int(variantInventoryItem.AvailableQuantity)))
				}
			}
		}
	}

	// Format the product
	formattedProduct := formatters.FormatProduct(resp)

	// Try to fetch inventory data for the product if inventory was created
	if inventoryCreated || (req.Product.Inventory != nil && inventoryClient != nil) {
		// Add a delay to ensure inventory data is available
		// This helps with eventual consistency between services
		time.Sleep(500 * time.Millisecond)

		// Fetch inventory data
		inventoryItem, err := inventoryClient.GetInventoryItem(c.Request.Context(), resp.Id)
		if err == nil && inventoryItem != nil {
			logger.Info("Successfully fetched inventory data",
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
				LastUpdated:       inventoryItem.LastUpdated.AsTime().Format(time.RFC3339),
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
			logger.Warn("Failed to fetch inventory data for product",
				zap.Error(err),
				zap.String("product_id", resp.Id))

			// If we can't fetch the inventory data but we know it was created,
			// provide a default inventory object with the initial quantity
			if inventoryCreated && req.Product.Inventory != nil {
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

	c.JSON(http.StatusCreated, formattedProduct)
}

// Helper function to handle gRPC errors
func handleGRPCError(c *gin.Context, err error, message string, logger *zap.Logger) {
	st, ok := status.FromError(err)
	if !ok {
		logger.Error(message, zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %v", message, err)})
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
	case codes.PermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	default:
		logger.Error(message, zap.Error(err), zap.String("grpc_code", st.Code().String()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", message, st.Message())})
	}
}
