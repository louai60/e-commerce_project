package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/louai60/e-commerce_project/backend/api-gateway/formatters"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DebugHandler handles debug-related requests
type DebugHandler struct {
	client pb.ProductServiceClient
	logger *zap.Logger
}

// NewDebugHandler creates a new debug handler
func NewDebugHandler(client pb.ProductServiceClient, logger *zap.Logger) *DebugHandler {
	return &DebugHandler{
		client: client,
		logger: logger,
	}
}

// GetProductRaw returns the raw product data from the product service
func (h *DebugHandler) GetProductRaw(c *gin.Context) {
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

	resp, err := h.client.GetProduct(c.Request.Context(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return both the raw proto response and the formatted response
	formattedProduct := formatters.FormatProduct(resp)
	
	c.JSON(http.StatusOK, gin.H{
		"raw_proto": map[string]interface{}{
			"id":                 resp.Id,
			"title":              resp.Title,
			"slug":               resp.Slug,
			"description":        resp.Description,
			"short_description":  resp.ShortDescription,
			"sku":                resp.Sku,
			"price":              resp.Price,
			"inventory_qty":      resp.InventoryQty,
			"inventory_status":   resp.InventoryStatus,
			"images_count":       len(resp.Images),
			"specifications_count": len(resp.Specifications),
			"tags_count":         len(resp.Tags),
			"has_brand":          resp.Brand != nil,
			"has_shipping":       resp.Shipping != nil,
			"has_seo":            resp.Seo != nil,
		},
		"formatted": formattedProduct,
	})
}
