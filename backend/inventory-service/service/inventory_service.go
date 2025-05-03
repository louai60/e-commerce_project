package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/louai60/e-commerce_project/backend/inventory-service/models"
	"github.com/louai60/e-commerce_project/backend/inventory-service/repository"
)

// InventoryService handles business logic for inventory operations
type InventoryService struct {
	inventoryRepo repository.InventoryRepository
	warehouseRepo repository.WarehouseRepository
	logger        *zap.Logger
}

// NewInventoryService creates a new inventory service
func NewInventoryService(
	inventoryRepo repository.InventoryRepository,
	warehouseRepo repository.WarehouseRepository,
	logger *zap.Logger,
) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		warehouseRepo: warehouseRepo,
		logger:        logger,
	}
}

// CreateInventoryItem creates a new inventory item with optional warehouse allocations
func (s *InventoryService) CreateInventoryItem(ctx context.Context, productID, sku string, variantID *string, initialQty, reorderPoint, reorderQty int, warehouseAllocations []models.WarehouseAllocation) (*models.InventoryItem, error) {
	// Check if inventory item already exists for this product/variant
	_, err := s.inventoryRepo.GetInventoryItemBySKU(ctx, sku)
	if err == nil {
		// Item already exists
		return nil, models.ErrAlreadyExists
	} else if err != models.ErrNotFound {
		// Unexpected error
		s.logger.Error("Error checking for existing inventory item", zap.Error(err))
		return nil, fmt.Errorf("error checking for existing inventory item: %w", err)
	}

	// Create new inventory item
	now := time.Now().UTC()
	item := &models.InventoryItem{
		ID:                uuid.New().String(),
		ProductID:         productID,
		VariantID:         variantID,
		SKU:               sku,
		TotalQuantity:     initialQty,
		AvailableQuantity: initialQty,
		ReservedQuantity:  0,
		ReorderPoint:      reorderPoint,
		ReorderQuantity:   reorderQty,
		Status:            models.DetermineInventoryStatus(initialQty, reorderPoint),
		LastUpdated:       now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// Create the inventory item
	if err := s.inventoryRepo.CreateInventoryItem(ctx, item); err != nil {
		s.logger.Error("Failed to create inventory item", zap.Error(err))
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}

	// If warehouse allocations are provided, distribute inventory
	if len(warehouseAllocations) > 0 {
		totalAllocated := 0
		for _, allocation := range warehouseAllocations {
			// Verify warehouse exists
			warehouse, err := s.warehouseRepo.GetWarehouseByID(ctx, allocation.WarehouseID)
			if err != nil {
				s.logger.Warn("Warehouse not found for allocation",
					zap.String("warehouse_id", allocation.WarehouseID),
					zap.Error(err))
				continue
			}

			if !warehouse.IsActive {
				s.logger.Warn("Cannot allocate to inactive warehouse",
					zap.String("warehouse_id", allocation.WarehouseID))
				continue
			}

			// Create inventory location
			location := &models.InventoryLocation{
				ID:                uuid.New().String(),
				InventoryItemID:   item.ID,
				WarehouseID:       allocation.WarehouseID,
				Quantity:          allocation.Quantity,
				AvailableQuantity: allocation.Quantity,
				ReservedQuantity:  0,
				CreatedAt:         now,
				UpdatedAt:         now,
			}

			if err := s.inventoryRepo.UpsertInventoryLocation(ctx, location); err != nil {
				s.logger.Error("Failed to create inventory location",
					zap.String("warehouse_id", allocation.WarehouseID),
					zap.Error(err))
				continue
			}

			totalAllocated += allocation.Quantity

			// Create transaction record
			transaction := &models.InventoryTransaction{
				ID:              uuid.New().String(),
				InventoryItemID: item.ID,
				WarehouseID:     &allocation.WarehouseID,
				TransactionType: models.TransactionStockAddition,
				Quantity:        allocation.Quantity,
				ReferenceID:     nil,
				ReferenceType:   nil,
				Notes:           nil,
				CreatedAt:       now,
			}

			if err := s.inventoryRepo.CreateInventoryTransaction(ctx, transaction); err != nil {
				s.logger.Warn("Failed to create transaction record", zap.Error(err))
				// Continue even if transaction record fails
			}
		}

		// If total allocated doesn't match initial quantity, log a warning
		if totalAllocated != initialQty {
			s.logger.Warn("Total allocated quantity doesn't match initial quantity",
				zap.Int("initial_qty", initialQty),
				zap.Int("total_allocated", totalAllocated))
		}
	} else if initialQty > 0 {
		// If no warehouse allocations but initial quantity > 0, create a transaction record
		transaction := &models.InventoryTransaction{
			ID:              uuid.New().String(),
			InventoryItemID: item.ID,
			WarehouseID:     nil,
			TransactionType: models.TransactionStockAddition,
			Quantity:        initialQty,
			ReferenceID:     nil,
			ReferenceType:   nil,
			Notes:           nil,
			CreatedAt:       now,
		}

		if err := s.inventoryRepo.CreateInventoryTransaction(ctx, transaction); err != nil {
			s.logger.Warn("Failed to create transaction record", zap.Error(err))
			// Continue even if transaction record fails
		}
	}

	// Retrieve the created item with locations
	createdItem, err := s.inventoryRepo.GetInventoryItemByID(ctx, item.ID)
	if err != nil {
		s.logger.Warn("Failed to retrieve created inventory item", zap.Error(err))
		return item, nil // Return the original item without locations
	}

	return createdItem, nil
}

// GetInventoryItem retrieves an inventory item by ID, product ID, or SKU
func (s *InventoryService) GetInventoryItem(ctx context.Context, id, productID, sku string) (*models.InventoryItem, error) {
	var item *models.InventoryItem
	var err error

	s.logger.Info("GetInventoryItem called",
		zap.String("id", id),
		zap.String("product_id", productID),
		zap.String("sku", sku))

	if id != "" {
		s.logger.Info("Getting inventory item by ID", zap.String("id", id))
		item, err = s.inventoryRepo.GetInventoryItemByID(ctx, id)
	} else if productID != "" {
		s.logger.Info("Getting inventory item by product ID", zap.String("product_id", productID))
		item, err = s.inventoryRepo.GetInventoryItemByProductID(ctx, productID)
	} else if sku != "" {
		s.logger.Info("Getting inventory item by SKU", zap.String("sku", sku))
		item, err = s.inventoryRepo.GetInventoryItemBySKU(ctx, sku)
	} else {
		s.logger.Warn("No identifier provided for GetInventoryItem")
		return nil, models.ErrInvalidInput
	}

	if err != nil {
		if err == models.ErrNotFound {
			s.logger.Warn("Inventory item not found",
				zap.String("id", id),
				zap.String("product_id", productID),
				zap.String("sku", sku))
			return nil, models.ErrNotFound
		}
		s.logger.Error("Failed to get inventory item",
			zap.Error(err),
			zap.String("id", id),
			zap.String("product_id", productID),
			zap.String("sku", sku))
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	s.logger.Info("Successfully retrieved inventory item",
		zap.String("item_id", item.ID),
		zap.String("product_id", item.ProductID),
		zap.String("sku", item.SKU),
		zap.Int("total_quantity", item.TotalQuantity),
		zap.Int("available_quantity", item.AvailableQuantity))

	return item, nil
}

// UpdateInventoryItem updates an inventory item's properties
func (s *InventoryService) UpdateInventoryItem(ctx context.Context, id string, reorderPoint, reorderQty *int, status *string) (*models.InventoryItem, error) {
	// Get the current item
	item, err := s.inventoryRepo.GetInventoryItemByID(ctx, id)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrNotFound
		}
		s.logger.Error("Failed to get inventory item for update", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get inventory item for update: %w", err)
	}

	// Update fields if provided
	if reorderPoint != nil {
		item.ReorderPoint = *reorderPoint
	}
	if reorderQty != nil {
		item.ReorderQuantity = *reorderQty
	}
	if status != nil {
		item.Status = *status
	} else {
		// Recalculate status based on current quantities and reorder point
		item.Status = models.DetermineInventoryStatus(item.AvailableQuantity, item.ReorderPoint)
	}

	// Update the item
	if err := s.inventoryRepo.UpdateInventoryItem(ctx, item); err != nil {
		s.logger.Error("Failed to update inventory item", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to update inventory item: %w", err)
	}

	return item, nil
}

// ListInventoryItems retrieves a paginated list of inventory items with optional filters
func (s *InventoryService) ListInventoryItems(ctx context.Context, page, limit int, status, warehouseID string, lowStockOnly bool) ([]*models.InventoryItem, int, error) {
	// Calculate offset from page and limit
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Build filters
	filters := make(map[string]interface{})
	if status != "" {
		filters["status"] = status
	}
	if warehouseID != "" {
		filters["warehouse_id"] = warehouseID
	}
	if lowStockOnly {
		filters["low_stock_only"] = true
	}

	// Get items from repository
	items, total, err := s.inventoryRepo.ListInventoryItems(ctx, offset, limit, filters)
	if err != nil {
		s.logger.Error("Failed to list inventory items", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list inventory items: %w", err)
	}

	// For each item, get its locations
	for _, item := range items {
		locations, err := s.inventoryRepo.GetInventoryLocations(ctx, item.ID)
		if err != nil {
			s.logger.Warn("Failed to get inventory locations",
				zap.Error(err),
				zap.String("inventory_item_id", item.ID))
			// Continue even if we can't get locations
		} else {
			item.Locations = locations
		}
	}

	return items, total, nil
}

// AddInventoryToLocation adds inventory to a specific warehouse location
func (s *InventoryService) AddInventoryToLocation(ctx context.Context, inventoryItemID, warehouseID string, quantity int, referenceID, referenceType, notes string) (*models.InventoryLocation, error) {
	// Validate inputs
	if quantity <= 0 {
		return nil, models.ErrInvalidQuantity
	}

	// Check if inventory item exists
	_, err := s.inventoryRepo.GetInventoryItemByID(ctx, inventoryItemID)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrNotFound
		}
		s.logger.Error("Failed to get inventory item", zap.Error(err), zap.String("id", inventoryItemID))
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	// Check if warehouse exists and is active
	warehouse, err := s.warehouseRepo.GetWarehouseByID(ctx, warehouseID)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrWarehouseNotFound
		}
		s.logger.Error("Failed to get warehouse", zap.Error(err), zap.String("id", warehouseID))
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}

	if !warehouse.IsActive {
		return nil, models.ErrWarehouseInactive
	}

	// Get existing location or create new one
	locations, err := s.inventoryRepo.GetInventoryLocations(ctx, inventoryItemID)
	if err != nil {
		s.logger.Error("Failed to get inventory locations", zap.Error(err), zap.String("inventory_item_id", inventoryItemID))
		return nil, fmt.Errorf("failed to get inventory locations: %w", err)
	}

	var location *models.InventoryLocation
	for i := range locations {
		if locations[i].WarehouseID == warehouseID {
			location = &locations[i]
			break
		}
	}

	now := time.Now().UTC()
	if location == nil {
		// Create new location
		location = &models.InventoryLocation{
			ID:                uuid.New().String(),
			InventoryItemID:   inventoryItemID,
			WarehouseID:       warehouseID,
			Quantity:          quantity,
			AvailableQuantity: quantity,
			ReservedQuantity:  0,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
	} else {
		// Update existing location
		location.Quantity += quantity
		location.AvailableQuantity += quantity
		location.UpdatedAt = now
	}

	// Update the location
	if err := s.inventoryRepo.UpsertInventoryLocation(ctx, location); err != nil {
		s.logger.Error("Failed to update inventory location", zap.Error(err))
		return nil, fmt.Errorf("failed to update inventory location: %w", err)
	}

	// Create transaction record
	var refID *string
	var refType *string
	var notePtr *string

	if referenceID != "" {
		refID = &referenceID
	}
	if referenceType != "" {
		refType = &referenceType
	}
	if notes != "" {
		notePtr = &notes
	}

	transaction := &models.InventoryTransaction{
		ID:              uuid.New().String(),
		InventoryItemID: inventoryItemID,
		WarehouseID:     &warehouseID,
		TransactionType: models.TransactionStockAddition,
		Quantity:        quantity,
		ReferenceID:     refID,
		ReferenceType:   refType,
		Notes:           notePtr,
		CreatedAt:       now,
	}

	if err := s.inventoryRepo.CreateInventoryTransaction(ctx, transaction); err != nil {
		s.logger.Warn("Failed to create transaction record", zap.Error(err))
		// Continue even if transaction record fails
	}

	// Set the warehouse in the location for the response
	location.Warehouse = warehouse

	return location, nil
}

// RemoveInventoryFromLocation removes inventory from a specific warehouse location
func (s *InventoryService) RemoveInventoryFromLocation(ctx context.Context, inventoryItemID, warehouseID string, quantity int, referenceID, referenceType, notes string) (*models.InventoryLocation, error) {
	// Validate inputs
	if quantity <= 0 {
		return nil, models.ErrInvalidQuantity
	}

	// Check if inventory item exists
	_, err := s.inventoryRepo.GetInventoryItemByID(ctx, inventoryItemID)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrNotFound
		}
		s.logger.Error("Failed to get inventory item", zap.Error(err), zap.String("id", inventoryItemID))
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	// Check if warehouse exists
	warehouse, err := s.warehouseRepo.GetWarehouseByID(ctx, warehouseID)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrWarehouseNotFound
		}
		s.logger.Error("Failed to get warehouse", zap.Error(err), zap.String("id", warehouseID))
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}

	// Get existing location
	locations, err := s.inventoryRepo.GetInventoryLocations(ctx, inventoryItemID)
	if err != nil {
		s.logger.Error("Failed to get inventory locations", zap.Error(err), zap.String("inventory_item_id", inventoryItemID))
		return nil, fmt.Errorf("failed to get inventory locations: %w", err)
	}

	var location *models.InventoryLocation
	for i := range locations {
		if locations[i].WarehouseID == warehouseID {
			location = &locations[i]
			break
		}
	}

	if location == nil {
		return nil, models.ErrNotFound
	}

	// Check if there's enough available inventory
	if location.AvailableQuantity < quantity {
		return nil, models.ErrInsufficientInventory
	}

	// Update the location
	now := time.Now().UTC()
	location.Quantity -= quantity
	location.AvailableQuantity -= quantity
	location.UpdatedAt = now

	if err := s.inventoryRepo.UpsertInventoryLocation(ctx, location); err != nil {
		s.logger.Error("Failed to update inventory location", zap.Error(err))
		return nil, fmt.Errorf("failed to update inventory location: %w", err)
	}

	// Create transaction record
	var refID *string
	var refType *string
	var notePtr *string

	if referenceID != "" {
		refID = &referenceID
	}
	if referenceType != "" {
		refType = &referenceType
	}
	if notes != "" {
		notePtr = &notes
	}

	transaction := &models.InventoryTransaction{
		ID:              uuid.New().String(),
		InventoryItemID: inventoryItemID,
		WarehouseID:     &warehouseID,
		TransactionType: models.TransactionStockRemoval,
		Quantity:        quantity,
		ReferenceID:     refID,
		ReferenceType:   refType,
		Notes:           notePtr,
		CreatedAt:       now,
	}

	if err := s.inventoryRepo.CreateInventoryTransaction(ctx, transaction); err != nil {
		s.logger.Warn("Failed to create transaction record", zap.Error(err))
		// Continue even if transaction record fails
	}

	// Set the warehouse in the location for the response
	location.Warehouse = warehouse

	return location, nil
}

// GetInventoryByLocation retrieves inventory items at a specific warehouse
func (s *InventoryService) GetInventoryByLocation(ctx context.Context, warehouseID string, page, limit int) ([]models.InventoryLocation, int, error) {
	// Check if warehouse exists
	warehouse, err := s.warehouseRepo.GetWarehouseByID(ctx, warehouseID)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, 0, models.ErrWarehouseNotFound
		}
		s.logger.Error("Failed to get warehouse", zap.Error(err), zap.String("id", warehouseID))
		return nil, 0, fmt.Errorf("failed to get warehouse: %w", err)
	}

	// Calculate offset from page and limit
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Get inventory locations
	locations, total, err := s.inventoryRepo.GetInventoryByWarehouse(ctx, warehouseID, offset, limit)
	if err != nil {
		s.logger.Error("Failed to get inventory by warehouse", zap.Error(err), zap.String("warehouse_id", warehouseID))
		return nil, 0, fmt.Errorf("failed to get inventory by warehouse: %w", err)
	}

	// Set the warehouse in each location
	for i := range locations {
		locations[i].Warehouse = warehouse
	}

	return locations, total, nil
}

// ReserveInventory creates temporary holds on inventory items
func (s *InventoryService) ReserveInventory(ctx context.Context, items []models.ReservationItem, referenceID, referenceType string, expirationMinutes int) (*models.InventoryReservation, error) {
	if len(items) == 0 {
		return nil, models.ErrInvalidInput
	}

	if referenceType == "" {
		return nil, models.ErrInvalidInput
	}

	// Default expiration time if not provided
	if expirationMinutes <= 0 {
		expirationMinutes = 30 // Default to 30 minutes
	}

	// We'll use the first item for the main reservation
	firstItem := items[0]

	// Check if inventory item exists
	_, err := s.inventoryRepo.GetInventoryItemByID(ctx, firstItem.InventoryItemID)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrNotFound
		}
		s.logger.Error("Failed to get inventory item", zap.Error(err), zap.String("id", firstItem.InventoryItemID))
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	// Create the reservation
	now := time.Now().UTC()
	expirationTime := now.Add(time.Duration(expirationMinutes) * time.Minute)

	var refID *string
	if referenceID != "" {
		refID = &referenceID
	}

	reservation := &models.InventoryReservation{
		ID:              uuid.New().String(),
		InventoryItemID: firstItem.InventoryItemID,
		WarehouseID:     firstItem.WarehouseID,
		Quantity:        firstItem.Quantity,
		Status:          models.ReservationPending,
		ExpirationTime:  expirationTime,
		ReferenceID:     refID,
		ReferenceType:   referenceType,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Create the reservation in the database
	if err := s.inventoryRepo.CreateReservation(ctx, reservation); err != nil {
		if err == models.ErrInsufficientInventory {
			return nil, models.ErrInsufficientInventory
		}
		s.logger.Error("Failed to create reservation", zap.Error(err))
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// If there are multiple items, create additional reservations
	// Note: In a real system, you might want to use a transaction to ensure all reservations succeed or fail together
	if len(items) > 1 {
		for i := 1; i < len(items); i++ {
			item := items[i]
			additionalReservation := &models.InventoryReservation{
				ID:              uuid.New().String(),
				InventoryItemID: item.InventoryItemID,
				WarehouseID:     item.WarehouseID,
				Quantity:        item.Quantity,
				Status:          models.ReservationPending,
				ExpirationTime:  expirationTime,
				ReferenceID:     refID,
				ReferenceType:   referenceType,
				CreatedAt:       now,
				UpdatedAt:       now,
			}

			if err := s.inventoryRepo.CreateReservation(ctx, additionalReservation); err != nil {
				// If one fails, we should ideally roll back all reservations
				// For simplicity, we'll just log the error and continue
				s.logger.Error("Failed to create additional reservation",
					zap.Error(err),
					zap.String("inventory_item_id", item.InventoryItemID))
			}
		}
	}

	return reservation, nil
}

// ConfirmReservation confirms a pending reservation
func (s *InventoryService) ConfirmReservation(ctx context.Context, reservationID string) (*models.InventoryReservation, error) {
	// Get the reservation
	reservation, err := s.inventoryRepo.GetReservationByID(ctx, reservationID)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrReservationNotFound
		}
		s.logger.Error("Failed to get reservation", zap.Error(err), zap.String("id", reservationID))
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	// Check if the reservation is in a valid state
	if reservation.Status != models.ReservationPending {
		return nil, models.ErrReservationInvalidState
	}

	// Check if the reservation has expired
	if time.Now().UTC().After(reservation.ExpirationTime) {
		return nil, models.ErrReservationExpired
	}

	// Update the reservation status
	reservation.Status = models.ReservationConfirmed
	reservation.UpdatedAt = time.Now().UTC()

	if err := s.inventoryRepo.UpdateReservation(ctx, reservation); err != nil {
		s.logger.Error("Failed to update reservation", zap.Error(err), zap.String("id", reservationID))
		return nil, fmt.Errorf("failed to update reservation: %w", err)
	}

	return reservation, nil
}

// CancelReservation cancels a pending reservation
func (s *InventoryService) CancelReservation(ctx context.Context, reservationID string) (*models.InventoryReservation, error) {
	// Get the reservation
	reservation, err := s.inventoryRepo.GetReservationByID(ctx, reservationID)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrReservationNotFound
		}
		s.logger.Error("Failed to get reservation", zap.Error(err), zap.String("id", reservationID))
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	// Check if the reservation is in a valid state
	if reservation.Status != models.ReservationPending {
		return nil, models.ErrReservationInvalidState
	}

	// Update the reservation status
	reservation.Status = models.ReservationCancelled
	reservation.UpdatedAt = time.Now().UTC()

	if err := s.inventoryRepo.UpdateReservation(ctx, reservation); err != nil {
		s.logger.Error("Failed to update reservation", zap.Error(err), zap.String("id", reservationID))
		return nil, fmt.Errorf("failed to update reservation: %w", err)
	}

	return reservation, nil
}

// CheckInventoryAvailability checks if requested quantities are available
func (s *InventoryService) CheckInventoryAvailability(ctx context.Context, items []models.AvailabilityCheckItem) ([]models.ItemAvailability, bool, error) {
	if len(items) == 0 {
		return nil, false, models.ErrInvalidInput
	}

	var results []models.ItemAvailability
	allAvailable := true

	for _, item := range items {
		// Get the inventory item
		var inventoryItem *models.InventoryItem
		var err error

		if item.SKU != "" {
			inventoryItem, err = s.inventoryRepo.GetInventoryItemBySKU(ctx, item.SKU)
		} else if item.ProductID != "" {
			inventoryItem, err = s.inventoryRepo.GetInventoryItemByProductID(ctx, item.ProductID)
		} else {
			// Skip items with no identifier
			continue
		}

		if err != nil {
			if err == models.ErrNotFound {
				// Item not found, mark as unavailable
				result := models.ItemAvailability{
					ProductID:         item.ProductID,
					VariantID:         item.VariantID,
					SKU:               item.SKU,
					RequestedQuantity: item.Quantity,
					AvailableQuantity: 0,
					IsAvailable:       false,
					Status:            "NOT_FOUND",
				}
				results = append(results, result)
				allAvailable = false
			} else {
				s.logger.Error("Failed to get inventory item", zap.Error(err))
				return nil, false, fmt.Errorf("failed to get inventory item: %w", err)
			}
			continue
		}

		// Check if there's enough available inventory
		isAvailable := inventoryItem.AvailableQuantity >= item.Quantity
		if !isAvailable {
			allAvailable = false
		}

		result := models.ItemAvailability{
			ProductID:         inventoryItem.ProductID,
			VariantID:         inventoryItem.VariantID,
			SKU:               inventoryItem.SKU,
			RequestedQuantity: item.Quantity,
			AvailableQuantity: inventoryItem.AvailableQuantity,
			IsAvailable:       isAvailable,
			Status:            inventoryItem.Status,
		}
		results = append(results, result)
	}

	return results, allAvailable, nil
}

// CleanExpiredReservations finds and cancels expired reservations
func (s *InventoryService) CleanExpiredReservations(ctx context.Context) (int, error) {
	count, err := s.inventoryRepo.CleanExpiredReservations(ctx)
	if err != nil {
		s.logger.Error("Failed to clean expired reservations", zap.Error(err))
		return 0, fmt.Errorf("failed to clean expired reservations: %w", err)
	}

	if count > 0 {
		s.logger.Info("Cleaned expired reservations", zap.Int("count", count))
	}

	return count, nil
}
