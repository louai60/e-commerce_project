package repository

import (
	"context"

	"github.com/louai60/e-commerce_project/backend/inventory-service/models"
)

// InventoryRepository defines the interface for inventory data operations
type InventoryRepository interface {
	// Inventory Item operations
	CreateInventoryItem(ctx context.Context, item *models.InventoryItem) error
	GetInventoryItemByID(ctx context.Context, id string) (*models.InventoryItem, error)
	GetInventoryItemByProductID(ctx context.Context, productID string) (*models.InventoryItem, error)
	GetInventoryItemBySKU(ctx context.Context, sku string) (*models.InventoryItem, error)
	UpdateInventoryItem(ctx context.Context, item *models.InventoryItem) error
	ListInventoryItems(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.InventoryItem, int, error)
	
	// Inventory Location operations
	GetInventoryLocations(ctx context.Context, inventoryItemID string) ([]models.InventoryLocation, error)
	UpsertInventoryLocation(ctx context.Context, location *models.InventoryLocation) error
	GetInventoryByWarehouse(ctx context.Context, warehouseID string, offset, limit int) ([]models.InventoryLocation, int, error)
	
	// Inventory Transaction operations
	CreateInventoryTransaction(ctx context.Context, transaction *models.InventoryTransaction) error
	GetInventoryTransactions(ctx context.Context, inventoryItemID string, limit int) ([]models.InventoryTransaction, error)
	
	// Inventory Reservation operations
	CreateReservation(ctx context.Context, reservation *models.InventoryReservation) error
	GetReservationByID(ctx context.Context, id string) (*models.InventoryReservation, error)
	UpdateReservation(ctx context.Context, reservation *models.InventoryReservation) error
	GetActiveReservations(ctx context.Context, inventoryItemID string) ([]models.InventoryReservation, error)
	CleanExpiredReservations(ctx context.Context) (int, error)
}

// WarehouseRepository defines the interface for warehouse data operations
type WarehouseRepository interface {
	CreateWarehouse(ctx context.Context, warehouse *models.Warehouse) error
	GetWarehouseByID(ctx context.Context, id string) (*models.Warehouse, error)
	GetWarehouseByCode(ctx context.Context, code string) (*models.Warehouse, error)
	UpdateWarehouse(ctx context.Context, warehouse *models.Warehouse) error
	ListWarehouses(ctx context.Context, offset, limit int, isActive *bool) ([]*models.Warehouse, int, error)
}
