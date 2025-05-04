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

// WarehouseService handles business logic for warehouse operations
type WarehouseService struct {
	warehouseRepo repository.WarehouseRepository
	logger        *zap.Logger
}

// NewWarehouseService creates a new warehouse service
func NewWarehouseService(
	warehouseRepo repository.WarehouseRepository,
	logger *zap.Logger,
) *WarehouseService {
	return &WarehouseService{
		warehouseRepo: warehouseRepo,
		logger:        logger,
	}
}

// CreateWarehouse creates a new warehouse
func (s *WarehouseService) CreateWarehouse(ctx context.Context, name, code, address, city, state, country, postalCode string, priority int) (*models.Warehouse, error) {
	// Validate inputs
	if name == "" || code == "" {
		return nil, models.ErrInvalidInput
	}

	// Check if warehouse with this code already exists
	_, err := s.warehouseRepo.GetWarehouseByCode(ctx, code)
	if err == nil {
		// Warehouse already exists
		return nil, models.ErrAlreadyExists
	} else if err != models.ErrNotFound {
		// Unexpected error
		s.logger.Error("Error checking for existing warehouse", zap.Error(err))
		return nil, fmt.Errorf("error checking for existing warehouse: %w", err)
	}

	// Create new warehouse
	now := time.Now().UTC()
	warehouse := &models.Warehouse{
		ID:         uuid.New().String(),
		Name:       name,
		Code:       code,
		Address:    address,
		City:       city,
		State:      state,
		Country:    country,
		PostalCode: postalCode,
		IsActive:   true,
		Priority:   priority,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Save the warehouse
	if err := s.warehouseRepo.CreateWarehouse(ctx, warehouse); err != nil {
		s.logger.Error("Failed to create warehouse", zap.Error(err))
		return nil, fmt.Errorf("failed to create warehouse: %w", err)
	}

	return warehouse, nil
}

// GetWarehouse retrieves a warehouse by ID or code
func (s *WarehouseService) GetWarehouse(ctx context.Context, id, code string) (*models.Warehouse, error) {
	var warehouse *models.Warehouse
	var err error

	if id != "" {
		warehouse, err = s.warehouseRepo.GetWarehouseByID(ctx, id)
	} else if code != "" {
		warehouse, err = s.warehouseRepo.GetWarehouseByCode(ctx, code)
	} else {
		return nil, models.ErrInvalidInput
	}

	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrNotFound
		}
		s.logger.Error("Failed to get warehouse", zap.Error(err))
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}

	return warehouse, nil
}

// UpdateWarehouse updates an existing warehouse
func (s *WarehouseService) UpdateWarehouse(ctx context.Context, id, name, address, city, state, country, postalCode string, priority *int, isActive *bool) (*models.Warehouse, error) {
	// Get the current warehouse
	warehouse, err := s.warehouseRepo.GetWarehouseByID(ctx, id)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrNotFound
		}
		s.logger.Error("Failed to get warehouse for update", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get warehouse for update: %w", err)
	}

	// Update fields if provided
	if name != "" {
		warehouse.Name = name
	}
	if address != "" {
		warehouse.Address = address
	}
	if city != "" {
		warehouse.City = city
	}
	if state != "" {
		warehouse.State = state
	}
	if country != "" {
		warehouse.Country = country
	}
	if postalCode != "" {
		warehouse.PostalCode = postalCode
	}
	if priority != nil {
		warehouse.Priority = *priority
	}
	if isActive != nil {
		warehouse.IsActive = *isActive
	}

	// Update the warehouse
	if err := s.warehouseRepo.UpdateWarehouse(ctx, warehouse); err != nil {
		s.logger.Error("Failed to update warehouse", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to update warehouse: %w", err)
	}

	return warehouse, nil
}

// ListWarehouses retrieves a paginated list of warehouses
func (s *WarehouseService) ListWarehouses(ctx context.Context, page, limit int, isActive *bool) ([]*models.Warehouse, int, error) {
	// Calculate offset from page and limit
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Get warehouses from repository
	warehouses, total, err := s.warehouseRepo.ListWarehouses(ctx, offset, limit, isActive)
	if err != nil {
		s.logger.Error("Failed to list warehouses", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list warehouses: %w", err)
	}

	return warehouses, total, nil
}
