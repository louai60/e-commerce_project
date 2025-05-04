package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/louai60/e-commerce_project/backend/inventory-service/models"
)

// InventoryRepository implements the repository.InventoryRepository interface
type InventoryRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewInventoryRepository creates a new PostgreSQL inventory repository
func NewInventoryRepository(db *sql.DB, logger *zap.Logger) *InventoryRepository {
	return &InventoryRepository{
		db:     db,
		logger: logger,
	}
}

// CreateInventoryItem creates a new inventory item in the database
func (r *InventoryRepository) CreateInventoryItem(ctx context.Context, item *models.InventoryItem) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set default values if not provided
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now
	item.LastUpdated = now

	// Insert the inventory item
	query := `
		INSERT INTO inventory_items (
			id, product_id, variant_id, sku, total_quantity, available_quantity,
			reserved_quantity, reorder_point, reorder_quantity, status,
			last_updated, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err = tx.ExecContext(
		ctx, query,
		item.ID, item.ProductID, item.VariantID, item.SKU, item.TotalQuantity,
		item.AvailableQuantity, item.ReservedQuantity, item.ReorderPoint,
		item.ReorderQuantity, item.Status, item.LastUpdated, item.CreatedAt, item.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create inventory item", zap.Error(err))
		return fmt.Errorf("failed to create inventory item: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetInventoryItemByID retrieves an inventory item by its ID
func (r *InventoryRepository) GetInventoryItemByID(ctx context.Context, id string) (*models.InventoryItem, error) {
	query := `
		SELECT
			id, product_id, variant_id, sku, total_quantity, available_quantity,
			reserved_quantity, reorder_point, reorder_quantity, status,
			last_updated, created_at, updated_at
		FROM inventory_items
		WHERE id = $1
	`

	var item models.InventoryItem
	var variantID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.ProductID, &variantID, &item.SKU, &item.TotalQuantity,
		&item.AvailableQuantity, &item.ReservedQuantity, &item.ReorderPoint,
		&item.ReorderQuantity, &item.Status, &item.LastUpdated, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFound
		}
		r.logger.Error("Failed to get inventory item by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get inventory item by ID: %w", err)
	}

	if variantID.Valid {
		item.VariantID = &variantID.String
	}

	// Get inventory locations
	locations, err := r.GetInventoryLocations(ctx, item.ID)
	if err != nil {
		r.logger.Warn("Failed to get inventory locations", zap.Error(err), zap.String("inventory_item_id", item.ID))
		// Continue even if we can't get locations
	} else {
		item.Locations = locations
	}

	return &item, nil
}

// GetInventoryItemByProductID retrieves an inventory item by product ID
func (r *InventoryRepository) GetInventoryItemByProductID(ctx context.Context, productID string) (*models.InventoryItem, error) {
	r.logger.Info("GetInventoryItemByProductID called", zap.String("product_id", productID))

	query := `
		SELECT
			id, product_id, variant_id, sku, total_quantity, available_quantity,
			reserved_quantity, reorder_point, reorder_quantity, status,
			last_updated, created_at, updated_at
		FROM inventory_items
		WHERE product_id = $1
	`

	r.logger.Debug("Executing query", zap.String("query", query), zap.String("product_id", productID))

	var item models.InventoryItem
	var variantID sql.NullString

	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&item.ID, &item.ProductID, &variantID, &item.SKU, &item.TotalQuantity,
		&item.AvailableQuantity, &item.ReservedQuantity, &item.ReorderPoint,
		&item.ReorderQuantity, &item.Status, &item.LastUpdated, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("No inventory item found for product ID", zap.String("product_id", productID))
			return nil, models.ErrNotFound
		}
		r.logger.Error("Failed to get inventory item by product ID", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get inventory item by product ID: %w", err)
	}

	r.logger.Info("Found inventory item by product ID",
		zap.String("product_id", productID),
		zap.String("item_id", item.ID),
		zap.Int("total_quantity", item.TotalQuantity),
		zap.Int("available_quantity", item.AvailableQuantity))

	if variantID.Valid {
		item.VariantID = &variantID.String
	}

	// Get inventory locations
	locations, err := r.GetInventoryLocations(ctx, item.ID)
	if err != nil {
		r.logger.Warn("Failed to get inventory locations", zap.Error(err), zap.String("inventory_item_id", item.ID))
		// Continue even if we can't get locations
	} else {
		item.Locations = locations
		r.logger.Info("Retrieved inventory locations",
			zap.String("item_id", item.ID),
			zap.Int("location_count", len(locations)))
	}

	return &item, nil
}

// GetInventoryItemBySKU retrieves an inventory item by SKU
func (r *InventoryRepository) GetInventoryItemBySKU(ctx context.Context, sku string) (*models.InventoryItem, error) {
	query := `
		SELECT
			id, product_id, variant_id, sku, total_quantity, available_quantity,
			reserved_quantity, reorder_point, reorder_quantity, status,
			last_updated, created_at, updated_at
		FROM inventory_items
		WHERE sku = $1
	`

	var item models.InventoryItem
	var variantID sql.NullString

	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&item.ID, &item.ProductID, &variantID, &item.SKU, &item.TotalQuantity,
		&item.AvailableQuantity, &item.ReservedQuantity, &item.ReorderPoint,
		&item.ReorderQuantity, &item.Status, &item.LastUpdated, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFound
		}
		r.logger.Error("Failed to get inventory item by SKU", zap.Error(err), zap.String("sku", sku))
		return nil, fmt.Errorf("failed to get inventory item by SKU: %w", err)
	}

	if variantID.Valid {
		item.VariantID = &variantID.String
	}

	// Get inventory locations
	locations, err := r.GetInventoryLocations(ctx, item.ID)
	if err != nil {
		r.logger.Warn("Failed to get inventory locations", zap.Error(err), zap.String("inventory_item_id", item.ID))
		// Continue even if we can't get locations
	} else {
		item.Locations = locations
	}

	return &item, nil
}

// UpdateInventoryItem updates an existing inventory item
func (r *InventoryRepository) UpdateInventoryItem(ctx context.Context, item *models.InventoryItem) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update the inventory item
	now := time.Now().UTC()
	item.UpdatedAt = now
	item.LastUpdated = now

	query := `
		UPDATE inventory_items
		SET
			total_quantity = $1,
			available_quantity = $2,
			reserved_quantity = $3,
			reorder_point = $4,
			reorder_quantity = $5,
			status = $6,
			last_updated = $7,
			updated_at = $8
		WHERE id = $9
	`

	result, err := tx.ExecContext(
		ctx, query,
		item.TotalQuantity, item.AvailableQuantity, item.ReservedQuantity,
		item.ReorderPoint, item.ReorderQuantity, item.Status,
		item.LastUpdated, item.UpdatedAt, item.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update inventory item", zap.Error(err), zap.String("id", item.ID))
		return fmt.Errorf("failed to update inventory item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrNotFound
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetInventoryLocations retrieves all inventory locations for an inventory item
func (r *InventoryRepository) GetInventoryLocations(ctx context.Context, inventoryItemID string) ([]models.InventoryLocation, error) {
	query := `
		SELECT
			id, inventory_item_id, warehouse_id, quantity, available_quantity,
			reserved_quantity, created_at, updated_at
		FROM inventory_locations
		WHERE inventory_item_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, inventoryItemID)
	if err != nil {
		r.logger.Error("Failed to get inventory locations", zap.Error(err), zap.String("inventory_item_id", inventoryItemID))
		return nil, fmt.Errorf("failed to get inventory locations: %w", err)
	}
	defer rows.Close()

	var locations []models.InventoryLocation
	for rows.Next() {
		var location models.InventoryLocation
		if err := rows.Scan(
			&location.ID, &location.InventoryItemID, &location.WarehouseID,
			&location.Quantity, &location.AvailableQuantity, &location.ReservedQuantity,
			&location.CreatedAt, &location.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan inventory location", zap.Error(err))
			return nil, fmt.Errorf("failed to scan inventory location: %w", err)
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating inventory locations", zap.Error(err))
		return nil, fmt.Errorf("error iterating inventory locations: %w", err)
	}

	return locations, nil
}

// UpsertInventoryLocation creates or updates an inventory location
func (r *InventoryRepository) UpsertInventoryLocation(ctx context.Context, location *models.InventoryLocation) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	location.UpdatedAt = now

	// If ID is not provided, generate one
	if location.ID == "" {
		location.ID = uuid.New().String()
		location.CreatedAt = now
	}

	query := `
		INSERT INTO inventory_locations (
			id, inventory_item_id, warehouse_id, quantity, available_quantity,
			reserved_quantity, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		ON CONFLICT (inventory_item_id, warehouse_id)
		DO UPDATE SET
			quantity = $4,
			available_quantity = $5,
			reserved_quantity = $6,
			updated_at = $8
		RETURNING id
	`

	var id string
	err = tx.QueryRowContext(
		ctx, query,
		location.ID, location.InventoryItemID, location.WarehouseID,
		location.Quantity, location.AvailableQuantity, location.ReservedQuantity,
		location.CreatedAt, location.UpdatedAt,
	).Scan(&id)

	if err != nil {
		r.logger.Error("Failed to upsert inventory location", zap.Error(err))
		return fmt.Errorf("failed to upsert inventory location: %w", err)
	}

	// Update the inventory item's total quantities
	updateItemQuery := `
		UPDATE inventory_items
		SET
			total_quantity = (
				SELECT COALESCE(SUM(quantity), 0)
				FROM inventory_locations
				WHERE inventory_item_id = $1
			),
			available_quantity = (
				SELECT COALESCE(SUM(available_quantity), 0)
				FROM inventory_locations
				WHERE inventory_item_id = $1
			),
			reserved_quantity = (
				SELECT COALESCE(SUM(reserved_quantity), 0)
				FROM inventory_locations
				WHERE inventory_item_id = $1
			),
			status = CASE
				WHEN (SELECT COALESCE(SUM(available_quantity), 0) FROM inventory_locations WHERE inventory_item_id = $1) <= 0 THEN 'OUT_OF_STOCK'
				WHEN (SELECT COALESCE(SUM(available_quantity), 0) FROM inventory_locations WHERE inventory_item_id = $1) <= reorder_point THEN 'LOW_STOCK'
				ELSE 'IN_STOCK'
			END,
			last_updated = $2,
			updated_at = $2
		WHERE id = $1
	`

	_, err = tx.ExecContext(ctx, updateItemQuery, location.InventoryItemID, now)
	if err != nil {
		r.logger.Error("Failed to update inventory item quantities", zap.Error(err))
		return fmt.Errorf("failed to update inventory item quantities: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetInventoryByWarehouse retrieves inventory items at a specific warehouse
func (r *InventoryRepository) GetInventoryByWarehouse(ctx context.Context, warehouseID string, offset, limit int) ([]models.InventoryLocation, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM inventory_locations
		WHERE warehouse_id = $1
	`

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, warehouseID).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count inventory locations", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count inventory locations: %w", err)
	}

	// Get inventory locations with pagination
	query := `
		SELECT
			l.id, l.inventory_item_id, l.warehouse_id, l.quantity, l.available_quantity,
			l.reserved_quantity, l.created_at, l.updated_at
		FROM inventory_locations l
		WHERE l.warehouse_id = $1
		ORDER BY l.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, warehouseID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get inventory by warehouse", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get inventory by warehouse: %w", err)
	}
	defer rows.Close()

	var locations []models.InventoryLocation
	for rows.Next() {
		var location models.InventoryLocation
		if err := rows.Scan(
			&location.ID, &location.InventoryItemID, &location.WarehouseID,
			&location.Quantity, &location.AvailableQuantity, &location.ReservedQuantity,
			&location.CreatedAt, &location.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan inventory location", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan inventory location: %w", err)
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating inventory locations", zap.Error(err))
		return nil, 0, fmt.Errorf("error iterating inventory locations: %w", err)
	}

	return locations, total, nil
}

// ListInventoryItems retrieves a paginated list of inventory items with optional filters
func (r *InventoryRepository) ListInventoryItems(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.InventoryItem, int, error) {
	// Build the WHERE clause based on filters
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if filters != nil {
		conditions := []string{}

		if status, ok := filters["status"]; ok {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, status)
			argIndex++
		}

		if warehouseID, ok := filters["warehouse_id"]; ok {
			conditions = append(conditions, fmt.Sprintf("id IN (SELECT inventory_item_id FROM inventory_locations WHERE warehouse_id = $%d)", argIndex))
			args = append(args, warehouseID)
			argIndex++
		}

		if lowStockOnly, ok := filters["low_stock_only"]; ok && lowStockOnly.(bool) {
			conditions = append(conditions, "(status = 'LOW_STOCK' OR status = 'OUT_OF_STOCK')")
		}

		if len(conditions) > 0 {
			whereClause = "WHERE " + strings.Join(conditions, " AND ")
		}
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM inventory_items
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count inventory items", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count inventory items: %w", err)
	}

	// Get inventory items with pagination
	query := fmt.Sprintf(`
		SELECT
			id, product_id, variant_id, sku, total_quantity, available_quantity,
			reserved_quantity, reorder_point, reorder_quantity, status,
			last_updated, created_at, updated_at
		FROM inventory_items
		%s
		ORDER BY last_updated DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list inventory items", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list inventory items: %w", err)
	}
	defer rows.Close()

	var items []*models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		var variantID sql.NullString

		if err := rows.Scan(
			&item.ID, &item.ProductID, &variantID, &item.SKU, &item.TotalQuantity,
			&item.AvailableQuantity, &item.ReservedQuantity, &item.ReorderPoint,
			&item.ReorderQuantity, &item.Status, &item.LastUpdated, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan inventory item", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan inventory item: %w", err)
		}

		if variantID.Valid {
			item.VariantID = &variantID.String
		}

		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating inventory items", zap.Error(err))
		return nil, 0, fmt.Errorf("error iterating inventory items: %w", err)
	}

	return items, total, nil
}

// CreateInventoryTransaction creates a new inventory transaction record
func (r *InventoryRepository) CreateInventoryTransaction(ctx context.Context, transaction *models.InventoryTransaction) error {
	// Generate ID if not provided
	if transaction.ID == "" {
		transaction.ID = uuid.New().String()
	}

	// Set created time if not provided
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO inventory_transactions (
			id, inventory_item_id, warehouse_id, transaction_type, quantity,
			reference_id, reference_type, notes, created_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		transaction.ID, transaction.InventoryItemID, transaction.WarehouseID,
		transaction.TransactionType, transaction.Quantity, transaction.ReferenceID,
		transaction.ReferenceType, transaction.Notes, transaction.CreatedBy,
		transaction.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create inventory transaction", zap.Error(err))
		return fmt.Errorf("failed to create inventory transaction: %w", err)
	}

	return nil
}

// GetInventoryTransactions retrieves transaction history for an inventory item
func (r *InventoryRepository) GetInventoryTransactions(ctx context.Context, inventoryItemID string, limit int) ([]models.InventoryTransaction, error) {
	query := `
		SELECT
			id, inventory_item_id, warehouse_id, transaction_type, quantity,
			reference_id, reference_type, notes, created_by, created_at
		FROM inventory_transactions
		WHERE inventory_item_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, inventoryItemID, limit)
	if err != nil {
		r.logger.Error("Failed to get inventory transactions", zap.Error(err))
		return nil, fmt.Errorf("failed to get inventory transactions: %w", err)
	}
	defer rows.Close()

	var transactions []models.InventoryTransaction
	for rows.Next() {
		var transaction models.InventoryTransaction
		var warehouseID, referenceID, referenceType, notes, createdBy sql.NullString

		if err := rows.Scan(
			&transaction.ID, &transaction.InventoryItemID, &warehouseID,
			&transaction.TransactionType, &transaction.Quantity, &referenceID,
			&referenceType, &notes, &createdBy, &transaction.CreatedAt,
		); err != nil {
			r.logger.Error("Failed to scan inventory transaction", zap.Error(err))
			return nil, fmt.Errorf("failed to scan inventory transaction: %w", err)
		}

		if warehouseID.Valid {
			transaction.WarehouseID = &warehouseID.String
		}
		if referenceID.Valid {
			transaction.ReferenceID = &referenceID.String
		}
		if referenceType.Valid {
			transaction.ReferenceType = &referenceType.String
		}
		if notes.Valid {
			transaction.Notes = &notes.String
		}
		if createdBy.Valid {
			transaction.CreatedBy = &createdBy.String
		}

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating inventory transactions", zap.Error(err))
		return nil, fmt.Errorf("error iterating inventory transactions: %w", err)
	}

	return transactions, nil
}

// CreateReservation creates a new inventory reservation
func (r *InventoryRepository) CreateReservation(ctx context.Context, reservation *models.InventoryReservation) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate ID if not provided
	if reservation.ID == "" {
		reservation.ID = uuid.New().String()
	}

	// Set timestamps if not provided
	now := time.Now().UTC()
	if reservation.CreatedAt.IsZero() {
		reservation.CreatedAt = now
	}
	if reservation.UpdatedAt.IsZero() {
		reservation.UpdatedAt = now
	}

	// Check if there's enough available inventory
	var availableQty int
	var checkQuery string

	if reservation.WarehouseID != nil {
		// Check specific warehouse
		checkQuery = `
			SELECT available_quantity
			FROM inventory_locations
			WHERE inventory_item_id = $1 AND warehouse_id = $2
		`
		err = tx.QueryRowContext(ctx, checkQuery, reservation.InventoryItemID, *reservation.WarehouseID).Scan(&availableQty)
	} else {
		// Check total available quantity
		checkQuery = `
			SELECT available_quantity
			FROM inventory_items
			WHERE id = $1
		`
		err = tx.QueryRowContext(ctx, checkQuery, reservation.InventoryItemID).Scan(&availableQty)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return models.ErrNotFound
		}
		r.logger.Error("Failed to check available quantity", zap.Error(err))
		return fmt.Errorf("failed to check available quantity: %w", err)
	}

	if availableQty < reservation.Quantity {
		return models.ErrInsufficientInventory
	}

	// Insert the reservation
	insertQuery := `
		INSERT INTO inventory_reservations (
			id, inventory_item_id, warehouse_id, quantity, status,
			expiration_time, reference_id, reference_type, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err = tx.ExecContext(
		ctx, insertQuery,
		reservation.ID, reservation.InventoryItemID, reservation.WarehouseID,
		reservation.Quantity, reservation.Status, reservation.ExpirationTime,
		reservation.ReferenceID, reservation.ReferenceType,
		reservation.CreatedAt, reservation.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create reservation", zap.Error(err))
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	// Update available quantity
	var updateQuery string
	if reservation.WarehouseID != nil {
		// Update specific warehouse
		updateQuery = `
			UPDATE inventory_locations
			SET
				available_quantity = available_quantity - $1,
				reserved_quantity = reserved_quantity + $1,
				updated_at = $2
			WHERE inventory_item_id = $3 AND warehouse_id = $4
		`
		_, err = tx.ExecContext(ctx, updateQuery, reservation.Quantity, now, reservation.InventoryItemID, *reservation.WarehouseID)
	} else {
		// Update inventory item (the triggers will handle updating the total)
		updateQuery = `
			UPDATE inventory_items
			SET
				available_quantity = available_quantity - $1,
				reserved_quantity = reserved_quantity + $1,
				updated_at = $2,
				last_updated = $2
			WHERE id = $3
		`
		_, err = tx.ExecContext(ctx, updateQuery, reservation.Quantity, now, reservation.InventoryItemID)
	}

	if err != nil {
		r.logger.Error("Failed to update available quantity", zap.Error(err))
		return fmt.Errorf("failed to update available quantity: %w", err)
	}

	// Create a transaction record
	transactionType := models.TransactionReservation
	// Convert ReferenceType from string to *string
	referenceType := reservation.ReferenceType

	transaction := &models.InventoryTransaction{
		ID:              uuid.New().String(),
		InventoryItemID: reservation.InventoryItemID,
		WarehouseID:     reservation.WarehouseID,
		TransactionType: transactionType,
		Quantity:        reservation.Quantity,
		ReferenceID:     reservation.ReferenceID,
		ReferenceType:   &referenceType,
		Notes:           nil, // Could add a note about the reservation
		CreatedAt:       now,
	}

	transactionQuery := `
		INSERT INTO inventory_transactions (
			id, inventory_item_id, warehouse_id, transaction_type, quantity,
			reference_id, reference_type, notes, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err = tx.ExecContext(
		ctx, transactionQuery,
		transaction.ID, transaction.InventoryItemID, transaction.WarehouseID,
		transaction.TransactionType, transaction.Quantity, transaction.ReferenceID,
		transaction.ReferenceType, transaction.Notes, transaction.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create transaction record", zap.Error(err))
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetReservationByID retrieves a reservation by its ID
func (r *InventoryRepository) GetReservationByID(ctx context.Context, id string) (*models.InventoryReservation, error) {
	query := `
		SELECT
			id, inventory_item_id, warehouse_id, quantity, status,
			expiration_time, reference_id, reference_type, created_at, updated_at
		FROM inventory_reservations
		WHERE id = $1
	`

	var reservation models.InventoryReservation
	var warehouseID, referenceID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&reservation.ID, &reservation.InventoryItemID, &warehouseID,
		&reservation.Quantity, &reservation.Status, &reservation.ExpirationTime,
		&referenceID, &reservation.ReferenceType, &reservation.CreatedAt, &reservation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFound
		}
		r.logger.Error("Failed to get reservation by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get reservation by ID: %w", err)
	}

	if warehouseID.Valid {
		reservation.WarehouseID = &warehouseID.String
	}
	if referenceID.Valid {
		reservation.ReferenceID = &referenceID.String
	}

	return &reservation, nil
}

// UpdateReservation updates an existing reservation
func (r *InventoryRepository) UpdateReservation(ctx context.Context, reservation *models.InventoryReservation) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the current reservation to compare changes
	var currentReservation models.InventoryReservation
	var warehouseID, referenceID sql.NullString

	getCurrentQuery := `
		SELECT
			id, inventory_item_id, warehouse_id, quantity, status,
			expiration_time, reference_id, reference_type, created_at, updated_at
		FROM inventory_reservations
		WHERE id = $1
		FOR UPDATE
	`

	err = tx.QueryRowContext(ctx, getCurrentQuery, reservation.ID).Scan(
		&currentReservation.ID, &currentReservation.InventoryItemID, &warehouseID,
		&currentReservation.Quantity, &currentReservation.Status, &currentReservation.ExpirationTime,
		&referenceID, &currentReservation.ReferenceType, &currentReservation.CreatedAt, &currentReservation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.ErrNotFound
		}
		r.logger.Error("Failed to get current reservation", zap.Error(err), zap.String("id", reservation.ID))
		return fmt.Errorf("failed to get current reservation: %w", err)
	}

	if warehouseID.Valid {
		currentReservation.WarehouseID = &warehouseID.String
	}
	if referenceID.Valid {
		currentReservation.ReferenceID = &referenceID.String
	}

	// Update the reservation
	now := time.Now().UTC()
	reservation.UpdatedAt = now

	updateQuery := `
		UPDATE inventory_reservations
		SET
			status = $1,
			updated_at = $2
		WHERE id = $3
	`

	result, err := tx.ExecContext(
		ctx, updateQuery,
		reservation.Status, reservation.UpdatedAt, reservation.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update reservation", zap.Error(err), zap.String("id", reservation.ID))
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrNotFound
	}

	// Handle inventory updates based on status change
	if currentReservation.Status != reservation.Status {
		var transactionType string
		var updateInventoryQuery string
		var notes *string

		switch reservation.Status {
		case models.ReservationCancelled, models.ReservationExpired:
			// Release the reserved inventory
			transactionType = models.TransactionReservationRelease
			noteStr := fmt.Sprintf("Reservation %s: %s", reservation.Status, reservation.ID)
			notes = &noteStr

			if currentReservation.WarehouseID != nil {
				// Update specific warehouse
				updateInventoryQuery = `
					UPDATE inventory_locations
					SET
						available_quantity = available_quantity + $1,
						reserved_quantity = reserved_quantity - $1,
						updated_at = $2
					WHERE inventory_item_id = $3 AND warehouse_id = $4
				`
				_, err = tx.ExecContext(
					ctx, updateInventoryQuery,
					currentReservation.Quantity, now, currentReservation.InventoryItemID, *currentReservation.WarehouseID,
				)
			} else {
				// Update inventory item
				updateInventoryQuery = `
					UPDATE inventory_items
					SET
						available_quantity = available_quantity + $1,
						reserved_quantity = reserved_quantity - $1,
						updated_at = $2,
						last_updated = $2
					WHERE id = $3
				`
				_, err = tx.ExecContext(
					ctx, updateInventoryQuery,
					currentReservation.Quantity, now, currentReservation.InventoryItemID,
				)
			}

			if err != nil {
				r.logger.Error("Failed to release reserved inventory", zap.Error(err))
				return fmt.Errorf("failed to release reserved inventory: %w", err)
			}

		case models.ReservationConfirmed, models.ReservationFulfilled:
			// For confirmed/fulfilled reservations, we keep the inventory reserved
			// but we might want to create a transaction record for tracking
			transactionType = models.TransactionStockRemoval
			noteStr := fmt.Sprintf("Reservation %s: %s", reservation.Status, reservation.ID)
			notes = &noteStr
		}

		// Create a transaction record if needed
		if transactionType != "" {
			// Convert ReferenceType from string to *string
			referenceType := currentReservation.ReferenceType

			transaction := &models.InventoryTransaction{
				ID:              uuid.New().String(),
				InventoryItemID: currentReservation.InventoryItemID,
				WarehouseID:     currentReservation.WarehouseID,
				TransactionType: transactionType,
				Quantity:        currentReservation.Quantity,
				ReferenceID:     currentReservation.ReferenceID,
				ReferenceType:   &referenceType,
				Notes:           notes,
				CreatedAt:       now,
			}

			transactionQuery := `
				INSERT INTO inventory_transactions (
					id, inventory_item_id, warehouse_id, transaction_type, quantity,
					reference_id, reference_type, notes, created_at
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9
				)
			`

			_, err = tx.ExecContext(
				ctx, transactionQuery,
				transaction.ID, transaction.InventoryItemID, transaction.WarehouseID,
				transaction.TransactionType, transaction.Quantity, transaction.ReferenceID,
				transaction.ReferenceType, transaction.Notes, transaction.CreatedAt,
			)

			if err != nil {
				r.logger.Error("Failed to create transaction record", zap.Error(err))
				return fmt.Errorf("failed to create transaction record: %w", err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetActiveReservations retrieves all active reservations for an inventory item
func (r *InventoryRepository) GetActiveReservations(ctx context.Context, inventoryItemID string) ([]models.InventoryReservation, error) {
	query := `
		SELECT
			id, inventory_item_id, warehouse_id, quantity, status,
			expiration_time, reference_id, reference_type, created_at, updated_at
		FROM inventory_reservations
		WHERE inventory_item_id = $1 AND status = 'PENDING'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, inventoryItemID)
	if err != nil {
		r.logger.Error("Failed to get active reservations", zap.Error(err))
		return nil, fmt.Errorf("failed to get active reservations: %w", err)
	}
	defer rows.Close()

	var reservations []models.InventoryReservation
	for rows.Next() {
		var reservation models.InventoryReservation
		var warehouseID, referenceID sql.NullString

		if err := rows.Scan(
			&reservation.ID, &reservation.InventoryItemID, &warehouseID,
			&reservation.Quantity, &reservation.Status, &reservation.ExpirationTime,
			&referenceID, &reservation.ReferenceType, &reservation.CreatedAt, &reservation.UpdatedAt,
		); err != nil {
			r.logger.Error("Failed to scan reservation", zap.Error(err))
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}

		if warehouseID.Valid {
			reservation.WarehouseID = &warehouseID.String
		}
		if referenceID.Valid {
			reservation.ReferenceID = &referenceID.String
		}

		reservations = append(reservations, reservation)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating reservations", zap.Error(err))
		return nil, fmt.Errorf("error iterating reservations: %w", err)
	}

	return reservations, nil
}

// CleanExpiredReservations finds and cancels expired reservations
func (r *InventoryRepository) CleanExpiredReservations(ctx context.Context) (int, error) {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()

	// Find expired reservations
	findQuery := `
		SELECT
			id, inventory_item_id, warehouse_id, quantity
		FROM inventory_reservations
		WHERE status = 'PENDING' AND expiration_time < $1
		FOR UPDATE
	`

	rows, err := tx.QueryContext(ctx, findQuery, now)
	if err != nil {
		r.logger.Error("Failed to find expired reservations", zap.Error(err))
		return 0, fmt.Errorf("failed to find expired reservations: %w", err)
	}
	defer rows.Close()

	type expiredReservation struct {
		ID              string
		InventoryItemID string
		WarehouseID     *string
		Quantity        int
	}

	var expiredReservations []expiredReservation
	for rows.Next() {
		var res expiredReservation
		var warehouseID sql.NullString

		if err := rows.Scan(&res.ID, &res.InventoryItemID, &warehouseID, &res.Quantity); err != nil {
			r.logger.Error("Failed to scan expired reservation", zap.Error(err))
			return 0, fmt.Errorf("failed to scan expired reservation: %w", err)
		}

		if warehouseID.Valid {
			res.WarehouseID = &warehouseID.String
		}

		expiredReservations = append(expiredReservations, res)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating expired reservations", zap.Error(err))
		return 0, fmt.Errorf("error iterating expired reservations: %w", err)
	}

	// Update reservation status
	updateQuery := `
		UPDATE inventory_reservations
		SET
			status = 'EXPIRED',
			updated_at = $1
		WHERE id = $2
	`

	// Release inventory and create transaction records
	releaseWarehouseQuery := `
		UPDATE inventory_locations
		SET
			available_quantity = available_quantity + $1,
			reserved_quantity = reserved_quantity - $1,
			updated_at = $2
		WHERE inventory_item_id = $3 AND warehouse_id = $4
	`

	releaseItemQuery := `
		UPDATE inventory_items
		SET
			available_quantity = available_quantity + $1,
			reserved_quantity = reserved_quantity - $1,
			updated_at = $2,
			last_updated = $2
		WHERE id = $3
	`

	transactionQuery := `
		INSERT INTO inventory_transactions (
			id, inventory_item_id, warehouse_id, transaction_type, quantity,
			reference_id, reference_type, notes, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	count := 0
	for _, res := range expiredReservations {
		// Update reservation status
		_, err := tx.ExecContext(ctx, updateQuery, now, res.ID)
		if err != nil {
			r.logger.Error("Failed to update reservation status", zap.Error(err), zap.String("id", res.ID))
			continue
		}

		// Release inventory
		if res.WarehouseID != nil {
			_, err = tx.ExecContext(ctx, releaseWarehouseQuery, res.Quantity, now, res.InventoryItemID, *res.WarehouseID)
		} else {
			_, err = tx.ExecContext(ctx, releaseItemQuery, res.Quantity, now, res.InventoryItemID)
		}

		if err != nil {
			r.logger.Error("Failed to release inventory", zap.Error(err), zap.String("id", res.ID))
			continue
		}

		// Create transaction record
		transactionID := uuid.New().String()
		noteStr := fmt.Sprintf("Reservation expired: %s", res.ID)
		notes := &noteStr
		referenceID := &res.ID
		referenceType := "RESERVATION"

		_, err = tx.ExecContext(
			ctx, transactionQuery,
			transactionID, res.InventoryItemID, res.WarehouseID,
			models.TransactionReservationRelease, res.Quantity, referenceID,
			referenceType, notes, now,
		)

		if err != nil {
			r.logger.Error("Failed to create transaction record", zap.Error(err), zap.String("id", res.ID))
			continue
		}

		count++
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}
