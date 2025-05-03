package clients

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/louai60/e-commerce_project/backend/api-gateway/config"
	inventorypb "github.com/louai60/e-commerce_project/backend/inventory-service/proto"
)

// InventoryClient handles communication with the inventory service
type InventoryClient struct {
	client inventorypb.InventoryServiceClient
	conn   *grpc.ClientConn
	logger *zap.Logger
}

// NewInventoryClient creates a new inventory service client
func NewInventoryClient(cfg *config.Config, logger *zap.Logger) (*InventoryClient, error) {
	// Get inventory service address from config
	inventoryAddr := fmt.Sprintf("%s:%s", cfg.Services.Inventory.Host, cfg.Services.Inventory.Port)
	logger.Info("Connecting to inventory service", zap.String("address", inventoryAddr))

	// Set up connection with retry logic
	var conn *grpc.ClientConn
	var err error
	maxRetries := 3
	retryInterval := time.Second * 2

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		conn, err = grpc.DialContext(
			ctx,
			inventoryAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		cancel()

		if err == nil {
			break
		}

		logger.Warn("Failed to connect to inventory service, retrying...",
			zap.String("address", inventoryAddr),
			zap.Error(err),
			zap.Int("attempt", i+1),
			zap.Int("maxRetries", maxRetries))

		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	if err != nil {
		logger.Error("Failed to connect to inventory service after retries",
			zap.String("address", inventoryAddr),
			zap.Error(err))
		return nil, fmt.Errorf("failed to connect to inventory service: %w", err)
	}

	return &InventoryClient{
		client: inventorypb.NewInventoryServiceClient(conn),
		conn:   conn,
		logger: logger,
	}, nil
}

// Close closes the gRPC connection
func (c *InventoryClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetInventoryItem retrieves inventory information for a product
func (c *InventoryClient) GetInventoryItem(ctx context.Context, productID string) (*inventorypb.InventoryItem, error) {
	c.logger.Info("Getting inventory item by product ID", zap.String("product_id", productID))

	// Create the request
	req := &inventorypb.GetInventoryItemRequest{
		Identifier: &inventorypb.GetInventoryItemRequest_ProductId{
			ProductId: productID,
		},
	}

	c.logger.Debug("Sending GetInventoryItem request",
		zap.String("product_id", productID),
		zap.Any("request", req))

	// Call the inventory service
	resp, err := c.client.GetInventoryItem(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get inventory item",
			zap.String("product_id", productID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	if resp == nil {
		c.logger.Error("Received nil response from inventory service",
			zap.String("product_id", productID))
		return nil, fmt.Errorf("received nil response from inventory service")
	}

	if resp.InventoryItem == nil {
		c.logger.Error("Received response with nil inventory item",
			zap.String("product_id", productID))
		return nil, fmt.Errorf("received response with nil inventory item")
	}

	// Log the inventory item details for debugging
	c.logger.Info("Retrieved inventory item",
		zap.String("product_id", productID),
		zap.String("item_id", resp.InventoryItem.Id),
		zap.Int32("total_quantity", resp.InventoryItem.TotalQuantity),
		zap.Int32("available_quantity", resp.InventoryItem.AvailableQuantity),
		zap.String("status", resp.InventoryItem.Status))

	return resp.InventoryItem, nil
}

// CheckInventoryAvailability checks if a product is available in the requested quantity
func (c *InventoryClient) CheckInventoryAvailability(ctx context.Context, productID string, quantity int) (bool, error) {
	c.logger.Info("Checking inventory availability",
		zap.String("product_id", productID),
		zap.Int("quantity", quantity))

	// Create the request
	req := &inventorypb.CheckInventoryAvailabilityRequest{
		Items: []*inventorypb.AvailabilityCheckItem{
			{
				ProductId: productID,
				Quantity:  int32(quantity),
			},
		},
	}

	// Call the inventory service
	resp, err := c.client.CheckInventoryAvailability(ctx, req)
	if err != nil {
		c.logger.Error("Failed to check inventory availability",
			zap.String("product_id", productID),
			zap.Error(err))
		return false, fmt.Errorf("failed to check inventory availability: %w", err)
	}

	return resp.AllAvailable, nil
}

// ListInventoryItems retrieves a paginated list of inventory items
func (c *InventoryClient) ListInventoryItems(ctx context.Context, page, limit int, status, warehouseID string, lowStockOnly bool) ([]*inventorypb.InventoryItem, int, error) {
	c.logger.Info("Listing inventory items",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("status", status),
		zap.String("warehouse_id", warehouseID),
		zap.Bool("low_stock_only", lowStockOnly))

	// Create the request
	req := &inventorypb.ListInventoryItemsRequest{
		Page:         int32(page),
		Limit:        int32(limit),
		LowStockOnly: lowStockOnly,
	}

	// Add optional filters
	if status != "" {
		req.Status = &wrappers.StringValue{Value: status}
	}
	if warehouseID != "" {
		req.WarehouseId = &wrappers.StringValue{Value: warehouseID}
	}

	// Call the inventory service
	resp, err := c.client.ListInventoryItems(ctx, req)
	if err != nil {
		c.logger.Error("Failed to list inventory items", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list inventory items: %w", err)
	}

	return resp.InventoryItems, int(resp.Total), nil
}

// CreateInventoryItem creates a new inventory item for a product
func (c *InventoryClient) CreateInventoryItem(ctx context.Context, productID, sku string, variantID *string, initialQty, reorderPoint, reorderQty int) (*inventorypb.InventoryItem, error) {
	c.logger.Info("Creating inventory item",
		zap.String("product_id", productID),
		zap.String("sku", sku),
		zap.Int("initial_quantity", initialQty))

	// Create the request
	req := &inventorypb.CreateInventoryItemRequest{
		ProductId:       productID,
		Sku:             sku,
		InitialQuantity: int32(initialQty),
		ReorderPoint:    int32(reorderPoint),
		ReorderQuantity: int32(reorderQty),
		// Warehouse allocations could be added here if needed
	}

	// Add variant ID if provided
	if variantID != nil && *variantID != "" {
		req.VariantId = &wrappers.StringValue{Value: *variantID}
	}

	// Call the inventory service
	resp, err := c.client.CreateInventoryItem(ctx, req)
	if err != nil {
		c.logger.Error("Failed to create inventory item",
			zap.String("product_id", productID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}

	// Log the created inventory item details for debugging
	c.logger.Info("Created inventory item",
		zap.String("product_id", productID),
		zap.Int32("initial_quantity", req.InitialQuantity),
		zap.Int32("total_quantity", resp.InventoryItem.TotalQuantity),
		zap.Int32("available_quantity", resp.InventoryItem.AvailableQuantity),
		zap.String("status", resp.InventoryItem.Status))

	return resp.InventoryItem, nil
}

// ListWarehouses retrieves a paginated list of warehouses
func (c *InventoryClient) ListWarehouses(ctx context.Context, page, limit int, isActive *bool) ([]*inventorypb.Warehouse, int, error) {
	c.logger.Info("Listing warehouses",
		zap.Int("page", page),
		zap.Int("limit", limit))

	// Create the request
	req := &inventorypb.ListWarehousesRequest{
		Page:  int32(page),
		Limit: int32(limit),
	}

	// Add optional filters
	if isActive != nil {
		req.IsActive = &wrappers.BoolValue{Value: *isActive}
	}

	// Call the inventory service
	resp, err := c.client.ListWarehouses(ctx, req)
	if err != nil {
		c.logger.Error("Failed to list warehouses", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list warehouses: %w", err)
	}

	return resp.Warehouses, int(resp.Total), nil
}

// ListInventoryTransactions retrieves a paginated list of inventory transactions
func (c *InventoryClient) ListInventoryTransactions(ctx context.Context, page, limit int, transactionType, warehouseID, dateFrom, dateTo string) ([]*inventorypb.InventoryTransaction, int, error) {
	c.logger.Info("Listing inventory transactions",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("transaction_type", transactionType),
		zap.String("warehouse_id", warehouseID),
		zap.String("date_from", dateFrom),
		zap.String("date_to", dateTo))

	// Since the inventory service doesn't have a direct gRPC endpoint for listing transactions,
	// we'll implement a simplified version that returns mock data for now
	// TODO: Implement proper transaction listing when the inventory service supports it

	// Create mock transactions for testing
	transactions := make([]*inventorypb.InventoryTransaction, 0)
	total := 0

	// In a real implementation, this would call the inventory service
	// For now, we'll return an empty list
	c.logger.Warn("ListInventoryTransactions is not fully implemented yet")

	return transactions, total, nil
}
