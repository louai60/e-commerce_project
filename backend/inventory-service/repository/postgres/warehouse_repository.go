package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/louai60/e-commerce_project/backend/inventory-service/models"
)

// WarehouseRepository implements the repository.WarehouseRepository interface
type WarehouseRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewWarehouseRepository creates a new PostgreSQL warehouse repository
func NewWarehouseRepository(db *sql.DB, logger *zap.Logger) *WarehouseRepository {
	return &WarehouseRepository{
		db:     db,
		logger: logger,
	}
}

// CreateWarehouse creates a new warehouse in the database
func (r *WarehouseRepository) CreateWarehouse(ctx context.Context, warehouse *models.Warehouse) error {
	// Generate ID if not provided
	if warehouse.ID == "" {
		warehouse.ID = uuid.New().String()
	}

	// Set timestamps if not provided
	now := time.Now().UTC()
	if warehouse.CreatedAt.IsZero() {
		warehouse.CreatedAt = now
	}
	if warehouse.UpdatedAt.IsZero() {
		warehouse.UpdatedAt = now
	}

	query := `
		INSERT INTO warehouses (
			id, name, code, address, city, state, country, postal_code,
			is_active, priority, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		warehouse.ID, warehouse.Name, warehouse.Code, warehouse.Address,
		warehouse.City, warehouse.State, warehouse.Country, warehouse.PostalCode,
		warehouse.IsActive, warehouse.Priority, warehouse.CreatedAt, warehouse.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create warehouse", zap.Error(err))
		return fmt.Errorf("failed to create warehouse: %w", err)
	}

	return nil
}

// GetWarehouseByID retrieves a warehouse by its ID
func (r *WarehouseRepository) GetWarehouseByID(ctx context.Context, id string) (*models.Warehouse, error) {
	query := `
		SELECT 
			id, name, code, address, city, state, country, postal_code,
			is_active, priority, created_at, updated_at
		FROM warehouses
		WHERE id = $1
	`

	var warehouse models.Warehouse
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&warehouse.ID, &warehouse.Name, &warehouse.Code, &warehouse.Address,
		&warehouse.City, &warehouse.State, &warehouse.Country, &warehouse.PostalCode,
		&warehouse.IsActive, &warehouse.Priority, &warehouse.CreatedAt, &warehouse.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFound
		}
		r.logger.Error("Failed to get warehouse by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get warehouse by ID: %w", err)
	}

	return &warehouse, nil
}

// GetWarehouseByCode retrieves a warehouse by its code
func (r *WarehouseRepository) GetWarehouseByCode(ctx context.Context, code string) (*models.Warehouse, error) {
	query := `
		SELECT 
			id, name, code, address, city, state, country, postal_code,
			is_active, priority, created_at, updated_at
		FROM warehouses
		WHERE code = $1
	`

	var warehouse models.Warehouse
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&warehouse.ID, &warehouse.Name, &warehouse.Code, &warehouse.Address,
		&warehouse.City, &warehouse.State, &warehouse.Country, &warehouse.PostalCode,
		&warehouse.IsActive, &warehouse.Priority, &warehouse.CreatedAt, &warehouse.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFound
		}
		r.logger.Error("Failed to get warehouse by code", zap.Error(err), zap.String("code", code))
		return nil, fmt.Errorf("failed to get warehouse by code: %w", err)
	}

	return &warehouse, nil
}

// UpdateWarehouse updates an existing warehouse
func (r *WarehouseRepository) UpdateWarehouse(ctx context.Context, warehouse *models.Warehouse) error {
	// Set updated timestamp
	now := time.Now().UTC()
	warehouse.UpdatedAt = now

	query := `
		UPDATE warehouses
		SET 
			name = $1,
			address = $2,
			city = $3,
			state = $4,
			country = $5,
			postal_code = $6,
			is_active = $7,
			priority = $8,
			updated_at = $9
		WHERE id = $10
	`

	result, err := r.db.ExecContext(
		ctx, query,
		warehouse.Name, warehouse.Address, warehouse.City, warehouse.State,
		warehouse.Country, warehouse.PostalCode, warehouse.IsActive, warehouse.Priority,
		warehouse.UpdatedAt, warehouse.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update warehouse", zap.Error(err), zap.String("id", warehouse.ID))
		return fmt.Errorf("failed to update warehouse: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrNotFound
	}

	return nil
}

// ListWarehouses retrieves a paginated list of warehouses with optional filters
func (r *WarehouseRepository) ListWarehouses(ctx context.Context, offset, limit int, isActive *bool) ([]*models.Warehouse, int, error) {
	// Build the WHERE clause based on filters
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if isActive != nil {
		whereClause = fmt.Sprintf("WHERE is_active = $%d", argIndex)
		args = append(args, *isActive)
		argIndex++
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM warehouses
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count warehouses", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count warehouses: %w", err)
	}

	// Get warehouses with pagination
	query := fmt.Sprintf(`
		SELECT 
			id, name, code, address, city, state, country, postal_code,
			is_active, priority, created_at, updated_at
		FROM warehouses
		%s
		ORDER BY priority DESC, name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list warehouses", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []*models.Warehouse
	for rows.Next() {
		var warehouse models.Warehouse
		if err := rows.Scan(
			&warehouse.ID, &warehouse.Name, &warehouse.Code, &warehouse.Address,
			&warehouse.City, &warehouse.State, &warehouse.Country, &warehouse.PostalCode,
			&warehouse.IsActive, &warehouse.Priority, &warehouse.CreatedAt, &warehouse.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan warehouse", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan warehouse: %w", err)
		}
		warehouses = append(warehouses, &warehouse)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating warehouses", zap.Error(err))
		return nil, 0, fmt.Errorf("error iterating warehouses: %w", err)
	}

	return warehouses, total, nil
}
