package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/louai60/e-commerce_project/backend/api-gateway/clients"
	inventorypb "github.com/louai60/e-commerce_project/backend/inventory-service/proto"
)

// InventoryHandler handles HTTP requests related to inventory
type InventoryHandler struct {
	client *clients.InventoryClient
	logger *zap.Logger
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(client *clients.InventoryClient, logger *zap.Logger) *InventoryHandler {
	return &InventoryHandler{
		client: client,
		logger: logger,
	}
}

// GetClient returns the inventory client
func (h *InventoryHandler) GetClient() *clients.InventoryClient {
	return h.client
}

// GetInventoryItem retrieves inventory information for a product
func (h *InventoryHandler) GetInventoryItem(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Inventory service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "inventory service unavailable"})
		return
	}

	productID := c.Param("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	// Call the inventory service
	inventoryItem, err := h.client.GetInventoryItem(c.Request.Context(), productID)
	if err != nil {
		h.handleGRPCError(c, err, "Failed to get inventory item")
		return
	}

	// Format the response
	c.JSON(http.StatusOK, formatInventoryItem(inventoryItem))
}

// CheckInventoryAvailability checks if a product is available in the requested quantity
func (h *InventoryHandler) CheckInventoryAvailability(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Inventory service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "inventory service unavailable"})
		return
	}

	// Parse request parameters
	productID := c.Query("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	quantityStr := c.Query("quantity")
	if quantityStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity is required"})
		return
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quantity"})
		return
	}

	// Call the inventory service
	available, err := h.client.CheckInventoryAvailability(c.Request.Context(), productID, quantity)
	if err != nil {
		h.handleGRPCError(c, err, "Failed to check inventory availability")
		return
	}

	// Format the response
	c.JSON(http.StatusOK, gin.H{
		"product_id": productID,
		"quantity":   quantity,
		"available":  available,
	})
}

// ListInventoryItems retrieves a paginated list of inventory items
func (h *InventoryHandler) ListInventoryItems(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Inventory service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "inventory service unavailable"})
		return
	}

	// Parse pagination parameters
	page, limit := getPaginationParams(c)

	// Parse filters
	status := c.Query("status")
	warehouseID := c.Query("warehouse_id")
	lowStockOnlyStr := c.Query("low_stock_only")
	lowStockOnly := lowStockOnlyStr == "true"

	// Call the inventory service
	items, total, err := h.client.ListInventoryItems(
		c.Request.Context(),
		page,
		limit,
		status,
		warehouseID,
		lowStockOnly,
	)
	if err != nil {
		h.handleGRPCError(c, err, "Failed to list inventory items")
		return
	}

	// Format the response
	formattedItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		formattedItems[i] = formatInventoryItem(item)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": formattedItems,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// ListWarehouses retrieves a paginated list of warehouses
func (h *InventoryHandler) ListWarehouses(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Inventory service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "inventory service unavailable"})
		return
	}

	// Parse pagination parameters
	page, limit := getPaginationParams(c)

	// Parse filters
	isActiveStr := c.Query("is_active")
	var isActive *bool
	if isActiveStr != "" {
		active := isActiveStr == "true"
		isActive = &active
	}

	// Call the inventory service
	warehouses, total, err := h.client.ListWarehouses(
		c.Request.Context(),
		page,
		limit,
		isActive,
	)
	if err != nil {
		h.handleGRPCError(c, err, "Failed to list warehouses")
		return
	}

	// Format the response
	formattedWarehouses := make([]map[string]interface{}, len(warehouses))
	for i, warehouse := range warehouses {
		formattedWarehouses[i] = formatWarehouse(warehouse)
	}

	c.JSON(http.StatusOK, gin.H{
		"warehouses": formattedWarehouses,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// Helper function to handle gRPC errors
func (h *InventoryHandler) handleGRPCError(c *gin.Context, err error, message string) {
	h.logger.Error(message, zap.Error(err))

	// Check if it's a gRPC status error
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
		case codes.InvalidArgument:
			c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
		case codes.PermissionDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
		case codes.Unauthenticated:
			c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
		case codes.ResourceExhausted:
			c.JSON(http.StatusTooManyRequests, gin.H{"error": st.Message()})
		case codes.Unavailable:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": st.Message()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": message})
		}
		return
	}

	// Default error handling
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

// Helper function to get pagination parameters
func getPaginationParams(c *gin.Context) (int, int) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	return page, limit
}

// Helper function to format inventory item response
func formatInventoryItem(item *inventorypb.InventoryItem) map[string]interface{} {
	if item == nil {
		return nil
	}

	// Format locations
	locations := make([]map[string]interface{}, len(item.Locations))
	for i, loc := range item.Locations {
		locations[i] = map[string]interface{}{
			"id":                 loc.Id,
			"warehouse_id":       loc.WarehouseId,
			"quantity":           loc.Quantity,
			"available_quantity": loc.AvailableQuantity,
			"reserved_quantity":  loc.ReservedQuantity,
			"warehouse":          formatWarehouse(loc.Warehouse),
		}
	}

	return map[string]interface{}{
		"id":                 item.Id,
		"product_id":         item.ProductId,
		"variant_id":         item.VariantId.GetValue(),
		"sku":                item.Sku,
		"total_quantity":     item.TotalQuantity,
		"available_quantity": item.AvailableQuantity,
		"reserved_quantity":  item.ReservedQuantity,
		"reorder_point":      item.ReorderPoint,
		"reorder_quantity":   item.ReorderQuantity,
		"status":             item.Status,
		"locations":          locations,
		"last_updated":       item.LastUpdated.AsTime().Format(time.RFC3339),
		"created_at":         item.CreatedAt.AsTime().Format(time.RFC3339),
		"updated_at":         item.UpdatedAt.AsTime().Format(time.RFC3339),
	}
}

// Helper function to format warehouse response
func formatWarehouse(warehouse *inventorypb.Warehouse) map[string]interface{} {
	if warehouse == nil {
		return nil
	}

	return map[string]interface{}{
		"id":          warehouse.Id,
		"name":        warehouse.Name,
		"code":        warehouse.Code,
		"address":     warehouse.Address,
		"city":        warehouse.City,
		"state":       warehouse.State,
		"country":     warehouse.Country,
		"postal_code": warehouse.PostalCode,
		"is_active":   warehouse.IsActive,
		"priority":    warehouse.Priority,
		"created_at":  warehouse.CreatedAt.AsTime().Format(time.RFC3339),
		"updated_at":  warehouse.UpdatedAt.AsTime().Format(time.RFC3339),
	}
}

// ListInventoryTransactions retrieves a paginated list of inventory transactions
func (h *InventoryHandler) ListInventoryTransactions(c *gin.Context) {
	// Check if client is nil
	if h.client == nil {
		h.logger.Error("Inventory service client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "inventory service unavailable"})
		return
	}

	// Parse pagination parameters
	page, limit := getPaginationParams(c)

	// Parse filter parameters
	transactionType := c.Query("transaction_type")
	warehouseID := c.Query("warehouse_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	// Get transactions from inventory service
	transactions, total, err := h.client.ListInventoryTransactions(
		c.Request.Context(),
		page,
		limit,
		transactionType,
		warehouseID,
		dateFrom,
		dateTo,
	)
	if err != nil {
		h.handleGRPCError(c, err, "Failed to list inventory transactions")
		return
	}

	// Format the response
	formattedTransactions := make([]map[string]interface{}, len(transactions))
	for i, transaction := range transactions {
		formattedTransactions[i] = formatInventoryTransaction(transaction)
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": formattedTransactions,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// Helper function to format inventory transaction response
func formatInventoryTransaction(transaction *inventorypb.InventoryTransaction) map[string]interface{} {
	if transaction == nil {
		return nil
	}

	result := map[string]interface{}{
		"id":                transaction.Id,
		"inventory_item_id": transaction.InventoryItemId,
		"transaction_type":  transaction.TransactionType,
		"quantity":          transaction.Quantity,
		"created_at":        transaction.CreatedAt.AsTime().Format(time.RFC3339),
	}

	// Add optional fields if present
	if transaction.WarehouseId != nil {
		result["warehouse_id"] = transaction.WarehouseId.Value
	}
	if transaction.ReferenceId != nil {
		result["reference_id"] = transaction.ReferenceId.Value
	}
	if transaction.ReferenceType != nil {
		result["reference_type"] = transaction.ReferenceType.Value
	}
	if transaction.Notes != nil {
		result["notes"] = transaction.Notes.Value
	}
	if transaction.CreatedBy != nil {
		result["created_by"] = transaction.CreatedBy.Value
	}

	return result
}
