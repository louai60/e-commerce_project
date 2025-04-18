package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"go.uber.org/zap"
)

// Tag-related methods
func (r *ProductRepository) GetProductTags(ctx context.Context, productID string) ([]models.ProductTag, error) {
	query := `
		SELECT id, product_id, tag, created_at, updated_at
		FROM product_tags
		WHERE product_id = $1
		ORDER BY tag
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get product tags", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product tags: %w", err)
	}
	defer rows.Close()

	var tags []models.ProductTag
	for rows.Next() {
		var tag models.ProductTag
		if err := rows.Scan(
			&tag.ID, &tag.ProductID, &tag.Tag, &tag.CreatedAt, &tag.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan product tag", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating product tags", zap.Error(err))
		return nil, fmt.Errorf("error iterating product tags: %w", err)
	}

	return tags, nil
}

func (r *ProductRepository) AddProductTag(ctx context.Context, tag *models.ProductTag) error {
	now := time.Now().UTC()
	tag.CreatedAt = now
	tag.UpdatedAt = now

	query := `
		INSERT INTO product_tags (product_id, tag, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, tag.ProductID, tag.Tag, now, now).Scan(&tag.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return fmt.Errorf("tag already exists for this product")
		}
		r.logger.Error("failed to add product tag", zap.Error(err))
		return fmt.Errorf("failed to add product tag: %w", err)
	}

	return nil
}

func (r *ProductRepository) RemoveProductTag(ctx context.Context, productID, tag string) error {
	query := `
		DELETE FROM product_tags
		WHERE product_id = $1 AND tag = $2
	`

	result, err := r.db.ExecContext(ctx, query, productID, tag)
	if err != nil {
		r.logger.Error("failed to remove product tag", zap.Error(err))
		return fmt.Errorf("failed to remove product tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag not found for this product")
	}

	return nil
}

// Attribute-related methods
func (r *ProductRepository) GetProductAttributes(ctx context.Context, productID string) ([]models.ProductAttribute, error) {
	query := `
		SELECT id, product_id, name, value, created_at, updated_at
		FROM product_attributes
		WHERE product_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get product attributes", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product attributes: %w", err)
	}
	defer rows.Close()

	var attributes []models.ProductAttribute
	for rows.Next() {
		var attr models.ProductAttribute
		if err := rows.Scan(
			&attr.ID, &attr.ProductID, &attr.Name, &attr.Value, &attr.CreatedAt, &attr.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan product attribute", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product attribute: %w", err)
		}
		attributes = append(attributes, attr)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating product attributes", zap.Error(err))
		return nil, fmt.Errorf("error iterating product attributes: %w", err)
	}

	return attributes, nil
}

func (r *ProductRepository) AddProductAttribute(ctx context.Context, attribute *models.ProductAttribute) error {
	now := time.Now().UTC()
	attribute.CreatedAt = now
	attribute.UpdatedAt = now

	query := `
		INSERT INTO product_attributes (product_id, name, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, attribute.ProductID, attribute.Name, attribute.Value, now, now).Scan(&attribute.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return fmt.Errorf("attribute already exists for this product")
		}
		r.logger.Error("failed to add product attribute", zap.Error(err))
		return fmt.Errorf("failed to add product attribute: %w", err)
	}

	return nil
}

func (r *ProductRepository) UpdateProductAttribute(ctx context.Context, attribute *models.ProductAttribute) error {
	now := time.Now().UTC()
	attribute.UpdatedAt = now

	query := `
		UPDATE product_attributes
		SET value = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, attribute.Value, now, attribute.ID)
	if err != nil {
		r.logger.Error("failed to update product attribute", zap.Error(err))
		return fmt.Errorf("failed to update product attribute: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("attribute not found")
	}

	return nil
}

func (r *ProductRepository) RemoveProductAttribute(ctx context.Context, attributeID string) error {
	query := `
		DELETE FROM product_attributes
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, attributeID)
	if err != nil {
		r.logger.Error("failed to remove product attribute", zap.Error(err))
		return fmt.Errorf("failed to remove product attribute: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("attribute not found")
	}

	return nil
}

// Specification-related methods
func (r *ProductRepository) GetProductSpecifications(ctx context.Context, productID string) ([]models.ProductSpecification, error) {
	query := `
		SELECT id, product_id, name, value, unit, created_at, updated_at
		FROM product_specifications
		WHERE product_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get product specifications", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product specifications: %w", err)
	}
	defer rows.Close()

	var specs []models.ProductSpecification
	for rows.Next() {
		var spec models.ProductSpecification
		if err := rows.Scan(
			&spec.ID, &spec.ProductID, &spec.Name, &spec.Value, &spec.Unit, &spec.CreatedAt, &spec.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan product specification", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product specification: %w", err)
		}
		specs = append(specs, spec)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating product specifications", zap.Error(err))
		return nil, fmt.Errorf("error iterating product specifications: %w", err)
	}

	return specs, nil
}

func (r *ProductRepository) AddProductSpecification(ctx context.Context, spec *models.ProductSpecification) error {
	now := time.Now().UTC()
	spec.CreatedAt = now
	spec.UpdatedAt = now

	query := `
		INSERT INTO product_specifications (product_id, name, value, unit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, spec.ProductID, spec.Name, spec.Value, spec.Unit, now, now).Scan(&spec.ID)
	if err != nil {
		r.logger.Error("failed to add product specification", zap.Error(err))
		return fmt.Errorf("failed to add product specification: %w", err)
	}

	return nil
}

func (r *ProductRepository) UpdateProductSpecification(ctx context.Context, spec *models.ProductSpecification) error {
	now := time.Now().UTC()
	spec.UpdatedAt = now

	query := `
		UPDATE product_specifications
		SET value = $1, unit = $2, updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, spec.Value, spec.Unit, now, spec.ID)
	if err != nil {
		r.logger.Error("failed to update product specification", zap.Error(err))
		return fmt.Errorf("failed to update product specification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("specification not found")
	}

	return nil
}

func (r *ProductRepository) RemoveProductSpecification(ctx context.Context, specID string) error {
	query := `
		DELETE FROM product_specifications
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, specID)
	if err != nil {
		r.logger.Error("failed to remove product specification", zap.Error(err))
		return fmt.Errorf("failed to remove product specification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("specification not found")
	}

	return nil
}

// SEO-related methods
func (r *ProductRepository) GetProductSEO(ctx context.Context, productID string) (*models.ProductSEO, error) {
	query := `
		SELECT id, product_id, meta_title, meta_description, keywords, tags, created_at, updated_at
		FROM product_seo
		WHERE product_id = $1
	`

	var seo models.ProductSEO
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&seo.ID, &seo.ProductID, &seo.MetaTitle, &seo.MetaDescription, 
		pq.Array(&seo.Keywords), pq.Array(&seo.Tags), &seo.CreatedAt, &seo.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No SEO data found
		}
		r.logger.Error("failed to get product SEO", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product SEO: %w", err)
	}

	return &seo, nil
}

func (r *ProductRepository) UpsertProductSEO(ctx context.Context, seo *models.ProductSEO) error {
	now := time.Now().UTC()
	seo.UpdatedAt = now

	// Check if SEO data already exists for this product
	existingSEO, err := r.GetProductSEO(ctx, seo.ProductID)
	if err != nil {
		return err
	}

	if existingSEO == nil {
		// Insert new SEO data
		seo.CreatedAt = now
		query := `
			INSERT INTO product_seo (product_id, meta_title, meta_description, keywords, tags, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`

		err := r.db.QueryRowContext(ctx, query, 
			seo.ProductID, seo.MetaTitle, seo.MetaDescription, 
			pq.Array(seo.Keywords), pq.Array(seo.Tags), now, now,
		).Scan(&seo.ID)
		if err != nil {
			r.logger.Error("failed to insert product SEO", zap.Error(err))
			return fmt.Errorf("failed to insert product SEO: %w", err)
		}
	} else {
		// Update existing SEO data
		seo.ID = existingSEO.ID
		seo.CreatedAt = existingSEO.CreatedAt
		query := `
			UPDATE product_seo
			SET meta_title = $1, meta_description = $2, keywords = $3, tags = $4, updated_at = $5
			WHERE id = $6
		`

		_, err := r.db.ExecContext(ctx, query, 
			seo.MetaTitle, seo.MetaDescription, 
			pq.Array(seo.Keywords), pq.Array(seo.Tags), now, seo.ID,
		)
		if err != nil {
			r.logger.Error("failed to update product SEO", zap.Error(err))
			return fmt.Errorf("failed to update product SEO: %w", err)
		}
	}

	return nil
}

// Shipping-related methods
func (r *ProductRepository) GetProductShipping(ctx context.Context, productID string) (*models.ProductShipping, error) {
	query := `
		SELECT id, product_id, free_shipping, estimated_days, express_available, created_at, updated_at
		FROM product_shipping
		WHERE product_id = $1
	`

	var shipping models.ProductShipping
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&shipping.ID, &shipping.ProductID, &shipping.FreeShipping, &shipping.EstimatedDays, 
		&shipping.ExpressAvailable, &shipping.CreatedAt, &shipping.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No shipping data found
		}
		r.logger.Error("failed to get product shipping", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product shipping: %w", err)
	}

	return &shipping, nil
}

func (r *ProductRepository) UpsertProductShipping(ctx context.Context, shipping *models.ProductShipping) error {
	now := time.Now().UTC()
	shipping.UpdatedAt = now

	// Check if shipping data already exists for this product
	existingShipping, err := r.GetProductShipping(ctx, shipping.ProductID)
	if err != nil {
		return err
	}

	if existingShipping == nil {
		// Insert new shipping data
		shipping.CreatedAt = now
		query := `
			INSERT INTO product_shipping (product_id, free_shipping, estimated_days, express_available, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		err := r.db.QueryRowContext(ctx, query, 
			shipping.ProductID, shipping.FreeShipping, shipping.EstimatedDays, 
			shipping.ExpressAvailable, now, now,
		).Scan(&shipping.ID)
		if err != nil {
			r.logger.Error("failed to insert product shipping", zap.Error(err))
			return fmt.Errorf("failed to insert product shipping: %w", err)
		}
	} else {
		// Update existing shipping data
		shipping.ID = existingShipping.ID
		shipping.CreatedAt = existingShipping.CreatedAt
		query := `
			UPDATE product_shipping
			SET free_shipping = $1, estimated_days = $2, express_available = $3, updated_at = $4
			WHERE id = $5
		`

		_, err := r.db.ExecContext(ctx, query, 
			shipping.FreeShipping, shipping.EstimatedDays, 
			shipping.ExpressAvailable, now, shipping.ID,
		)
		if err != nil {
			r.logger.Error("failed to update product shipping", zap.Error(err))
			return fmt.Errorf("failed to update product shipping: %w", err)
		}
	}

	return nil
}

// Discount-related methods
func (r *ProductRepository) GetProductDiscounts(ctx context.Context, productID string) ([]models.ProductDiscount, error) {
	query := `
		SELECT id, product_id, discount_type, value, expires_at, created_at, updated_at
		FROM product_discounts
		WHERE product_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get product discounts", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product discounts: %w", err)
	}
	defer rows.Close()

	var discounts []models.ProductDiscount
	for rows.Next() {
		var discount models.ProductDiscount
		if err := rows.Scan(
			&discount.ID, &discount.ProductID, &discount.Type, &discount.Value, 
			&discount.ExpiresAt, &discount.CreatedAt, &discount.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan product discount", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product discount: %w", err)
		}
		discounts = append(discounts, discount)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating product discounts", zap.Error(err))
		return nil, fmt.Errorf("error iterating product discounts: %w", err)
	}

	return discounts, nil
}

func (r *ProductRepository) AddProductDiscount(ctx context.Context, discount *models.ProductDiscount) error {
	now := time.Now().UTC()
	discount.CreatedAt = now
	discount.UpdatedAt = now

	query := `
		INSERT INTO product_discounts (product_id, discount_type, value, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, 
		discount.ProductID, discount.Type, discount.Value, 
		discount.ExpiresAt, now, now,
	).Scan(&discount.ID)
	if err != nil {
		r.logger.Error("failed to add product discount", zap.Error(err))
		return fmt.Errorf("failed to add product discount: %w", err)
	}

	return nil
}

func (r *ProductRepository) UpdateProductDiscount(ctx context.Context, discount *models.ProductDiscount) error {
	now := time.Now().UTC()
	discount.UpdatedAt = now

	query := `
		UPDATE product_discounts
		SET discount_type = $1, value = $2, expires_at = $3, updated_at = $4
		WHERE id = $5
	`

	result, err := r.db.ExecContext(ctx, query, 
		discount.Type, discount.Value, discount.ExpiresAt, now, discount.ID,
	)
	if err != nil {
		r.logger.Error("failed to update product discount", zap.Error(err))
		return fmt.Errorf("failed to update product discount: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("discount not found")
	}

	return nil
}

func (r *ProductRepository) RemoveProductDiscount(ctx context.Context, discountID string) error {
	query := `
		DELETE FROM product_discounts
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, discountID)
	if err != nil {
		r.logger.Error("failed to remove product discount", zap.Error(err))
		return fmt.Errorf("failed to remove product discount: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("discount not found")
	}

	return nil
}

// Inventory-related methods
func (r *ProductRepository) GetInventoryLocations(ctx context.Context, productID string) ([]models.InventoryLocation, error) {
	query := `
		SELECT id, product_id, warehouse_id, available_qty, created_at, updated_at
		FROM product_inventory_locations
		WHERE product_id = $1
		ORDER BY warehouse_id
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get inventory locations", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get inventory locations: %w", err)
	}
	defer rows.Close()

	var locations []models.InventoryLocation
	for rows.Next() {
		var location models.InventoryLocation
		if err := rows.Scan(
			&location.ID, &location.ProductID, &location.WarehouseID, 
			&location.AvailableQty, &location.CreatedAt, &location.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan inventory location", zap.Error(err))
			return nil, fmt.Errorf("failed to scan inventory location: %w", err)
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating inventory locations", zap.Error(err))
		return nil, fmt.Errorf("error iterating inventory locations: %w", err)
	}

	return locations, nil
}

func (r *ProductRepository) UpsertInventoryLocation(ctx context.Context, location *models.InventoryLocation) error {
	now := time.Now().UTC()
	location.UpdatedAt = now

	// Check if location already exists
	query := `
		SELECT id, created_at FROM product_inventory_locations
		WHERE product_id = $1 AND warehouse_id = $2
	`
	var existingID string
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, location.ProductID, location.WarehouseID).Scan(&existingID, &createdAt)
	
	if err != nil && err != sql.ErrNoRows {
		r.logger.Error("failed to check existing inventory location", zap.Error(err))
		return fmt.Errorf("failed to check existing inventory location: %w", err)
	}

	if err == sql.ErrNoRows {
		// Insert new location
		location.CreatedAt = now
		query := `
			INSERT INTO product_inventory_locations (product_id, warehouse_id, available_qty, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`

		err := r.db.QueryRowContext(ctx, query, 
			location.ProductID, location.WarehouseID, location.AvailableQty, now, now,
		).Scan(&location.ID)
		if err != nil {
			r.logger.Error("failed to insert inventory location", zap.Error(err))
			return fmt.Errorf("failed to insert inventory location: %w", err)
		}
	} else {
		// Update existing location
		location.ID = existingID
		location.CreatedAt = createdAt
		query := `
			UPDATE product_inventory_locations
			SET available_qty = $1, updated_at = $2
			WHERE id = $3
		`

		_, err := r.db.ExecContext(ctx, query, location.AvailableQty, now, location.ID)
		if err != nil {
			r.logger.Error("failed to update inventory location", zap.Error(err))
			return fmt.Errorf("failed to update inventory location: %w", err)
		}
	}

	return nil
}

func (r *ProductRepository) RemoveInventoryLocation(ctx context.Context, productID, warehouseID string) error {
	query := `
		DELETE FROM product_inventory_locations
		WHERE product_id = $1 AND warehouse_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, productID, warehouseID)
	if err != nil {
		r.logger.Error("failed to remove inventory location", zap.Error(err))
		return fmt.Errorf("failed to remove inventory location: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("inventory location not found")
	}

	return nil
}
