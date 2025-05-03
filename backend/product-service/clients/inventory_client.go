package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/louai60/e-commerce_project/backend/product-service/config"
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

// CreateInventoryItem creates a new inventory item for a product
func (c *InventoryClient) CreateInventoryItem(ctx context.Context, productID, sku string, variantID *string, initialQty, reorderPoint, reorderQty int) (*inventorypb.InventoryItem, error) {
	c.logger.Info("Creating inventory item",
		zap.String("product_id", productID),
		zap.String("sku", sku),
		zap.Int("initial_quantity", initialQty))

	// Convert variant ID to wrapper if provided
	var variantIDWrapper *wrappers.StringValue
	if variantID != nil && *variantID != "" {
		variantIDWrapper = &wrappers.StringValue{Value: *variantID}
	}

	// Create the request
	req := &inventorypb.CreateInventoryItemRequest{
		ProductId:       productID,
		VariantId:       variantIDWrapper,
		Sku:             sku,
		InitialQuantity: int32(initialQty),
		ReorderPoint:    int32(reorderPoint),
		ReorderQuantity: int32(reorderQty),
		// Warehouse allocations could be added here if needed
	}

	// Call the inventory service
	resp, err := c.client.CreateInventoryItem(ctx, req)
	if err != nil {
		c.logger.Error("Failed to create inventory item",
			zap.String("product_id", productID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}

	return resp.InventoryItem, nil
}

// GetInventoryItem retrieves an inventory item by product ID
func (c *InventoryClient) GetInventoryItem(ctx context.Context, productID string) (*inventorypb.InventoryItem, error) {
	c.logger.Info("Getting inventory item by product ID", zap.String("product_id", productID))

	// Create the request
	req := &inventorypb.GetInventoryItemRequest{
		Identifier: &inventorypb.GetInventoryItemRequest_ProductId{
			ProductId: productID,
		},
	}

	// Call the inventory service
	resp, err := c.client.GetInventoryItem(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get inventory item",
			zap.String("product_id", productID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	return resp.InventoryItem, nil
}

// UpdateInventoryItem updates an existing inventory item
func (c *InventoryClient) UpdateInventoryItem(ctx context.Context, id string, reorderPoint, reorderQty *int, status *string) (*inventorypb.InventoryItem, error) {
	c.logger.Info("Updating inventory item", zap.String("id", id))

	// Convert optional fields to wrappers
	var reorderPointWrapper *wrappers.Int32Value
	var reorderQtyWrapper *wrappers.Int32Value
	var statusWrapper *wrappers.StringValue

	if reorderPoint != nil {
		reorderPointWrapper = &wrappers.Int32Value{Value: int32(*reorderPoint)}
	}

	if reorderQty != nil {
		reorderQtyWrapper = &wrappers.Int32Value{Value: int32(*reorderQty)}
	}

	if status != nil && *status != "" {
		statusWrapper = &wrappers.StringValue{Value: *status}
	}

	// Create the request
	req := &inventorypb.UpdateInventoryItemRequest{
		Id:             id,
		ReorderPoint:   reorderPointWrapper,
		ReorderQuantity: reorderQtyWrapper,
		Status:         statusWrapper,
	}

	// Call the inventory service
	resp, err := c.client.UpdateInventoryItem(ctx, req)
	if err != nil {
		c.logger.Error("Failed to update inventory item",
			zap.String("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update inventory item: %w", err)
	}

	return resp.InventoryItem, nil
}

// CheckInventoryAvailability checks if a product is available in the requested quantity
func (c *InventoryClient) CheckInventoryAvailability(ctx context.Context, productID string, variantID *string, sku string, quantity int) (bool, error) {
	c.logger.Info("Checking inventory availability",
		zap.String("product_id", productID),
		zap.String("sku", sku),
		zap.Int("quantity", quantity))

	// Convert variant ID to wrapper if provided
	var variantIDWrapper *wrappers.StringValue
	if variantID != nil && *variantID != "" {
		variantIDWrapper = &wrappers.StringValue{Value: *variantID}
	}

	// Create the request
	req := &inventorypb.CheckInventoryAvailabilityRequest{
		Items: []*inventorypb.AvailabilityCheckItem{
			{
				ProductId: productID,
				VariantId: variantIDWrapper,
				Sku:       sku,
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
