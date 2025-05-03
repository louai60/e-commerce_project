package models

import (
	"time"
)

// InventoryItem represents the main inventory record for a product or variant
type InventoryItem struct {
	ID                string              `json:"id" db:"id"`
	ProductID         string              `json:"product_id" db:"product_id"`
	VariantID         *string             `json:"variant_id,omitempty" db:"variant_id"`
	SKU               string              `json:"sku" db:"sku"`
	TotalQuantity     int                 `json:"total_quantity" db:"total_quantity"`
	AvailableQuantity int                 `json:"available_quantity" db:"available_quantity"`
	ReservedQuantity  int                 `json:"reserved_quantity" db:"reserved_quantity"`
	ReorderPoint      int                 `json:"reorder_point" db:"reorder_point"`
	ReorderQuantity   int                 `json:"reorder_quantity" db:"reorder_quantity"`
	Status            string              `json:"status" db:"status"`
	LastUpdated       time.Time           `json:"last_updated" db:"last_updated"`
	CreatedAt         time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at" db:"updated_at"`
	Locations         []InventoryLocation `json:"locations,omitempty" db:"-"`
}

// InventoryLocation represents inventory at a specific warehouse
type InventoryLocation struct {
	ID                string     `json:"id" db:"id"`
	InventoryItemID   string     `json:"inventory_item_id" db:"inventory_item_id"`
	WarehouseID       string     `json:"warehouse_id" db:"warehouse_id"`
	Quantity          int        `json:"quantity" db:"quantity"`
	AvailableQuantity int        `json:"available_quantity" db:"available_quantity"`
	ReservedQuantity  int        `json:"reserved_quantity" db:"reserved_quantity"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	Warehouse         *Warehouse `json:"warehouse,omitempty" db:"-"`
}

// InventoryTransaction represents a change in inventory
type InventoryTransaction struct {
	ID              string    `json:"id" db:"id"`
	InventoryItemID string    `json:"inventory_item_id" db:"inventory_item_id"`
	WarehouseID     *string   `json:"warehouse_id,omitempty" db:"warehouse_id"`
	TransactionType string    `json:"transaction_type" db:"transaction_type"`
	Quantity        int       `json:"quantity" db:"quantity"`
	ReferenceID     *string   `json:"reference_id,omitempty" db:"reference_id"`
	ReferenceType   *string   `json:"reference_type,omitempty" db:"reference_type"`
	Notes           *string   `json:"notes,omitempty" db:"notes"`
	CreatedBy       *string   `json:"created_by,omitempty" db:"created_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// InventoryReservation represents a temporary hold on inventory
type InventoryReservation struct {
	ID              string    `json:"id" db:"id"`
	InventoryItemID string    `json:"inventory_item_id" db:"inventory_item_id"`
	WarehouseID     *string   `json:"warehouse_id,omitempty" db:"warehouse_id"`
	Quantity        int       `json:"quantity" db:"quantity"`
	Status          string    `json:"status" db:"status"`
	ExpirationTime  time.Time `json:"expiration_time" db:"expiration_time"`
	ReferenceID     *string   `json:"reference_id,omitempty" db:"reference_id"`
	ReferenceType   string    `json:"reference_type" db:"reference_type"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// WarehouseAllocation represents a quantity allocation to a specific warehouse
type WarehouseAllocation struct {
	WarehouseID string `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
}

// ReservationItem represents an item to be reserved
type ReservationItem struct {
	InventoryItemID string  `json:"inventory_item_id"`
	Quantity        int     `json:"quantity"`
	WarehouseID     *string `json:"warehouse_id,omitempty"`
}

// AvailabilityCheckItem represents an item to check for availability
type AvailabilityCheckItem struct {
	ProductID string  `json:"product_id"`
	VariantID *string `json:"variant_id,omitempty"`
	SKU       string  `json:"sku"`
	Quantity  int     `json:"quantity"`
}

// ItemAvailability represents the availability status of an item
type ItemAvailability struct {
	ProductID         string  `json:"product_id"`
	VariantID         *string `json:"variant_id,omitempty"`
	SKU               string  `json:"sku"`
	RequestedQuantity int     `json:"requested_quantity"`
	AvailableQuantity int     `json:"available_quantity"`
	IsAvailable       bool    `json:"is_available"`
	Status            string  `json:"status"`
}

// Constants for inventory status
const (
	StatusInStock      = "IN_STOCK"
	StatusLowStock     = "LOW_STOCK"
	StatusOutOfStock   = "OUT_OF_STOCK"
	StatusDiscontinued = "DISCONTINUED"
	StatusBackordered  = "BACKORDERED"
)

// Constants for transaction types
const (
	TransactionStockAddition      = "STOCK_ADDITION"
	TransactionStockRemoval       = "STOCK_REMOVAL"
	TransactionReservation        = "RESERVATION"
	TransactionReservationRelease = "RESERVATION_RELEASE"
	TransactionAdjustment         = "ADJUSTMENT"
)

// Constants for reservation status
const (
	ReservationPending   = "PENDING"
	ReservationConfirmed = "CONFIRMED"
	ReservationCancelled = "CANCELLED"
	ReservationFulfilled = "FULFILLED"
	ReservationExpired   = "EXPIRED"
)

// DetermineInventoryStatus calculates the appropriate inventory status based on quantities
func DetermineInventoryStatus(availableQty, reorderPoint int) string {
	if availableQty <= 0 {
		return StatusOutOfStock
	} else if availableQty <= reorderPoint {
		return StatusLowStock
	}
	return StatusInStock
}
