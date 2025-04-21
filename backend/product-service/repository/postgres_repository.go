package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// Ensure PostgresRepository implements all required interfaces
var _ ProductRepository = (*PostgresRepository)(nil)
var _ BrandRepository = (*PostgresRepository)(nil)
var _ CategoryRepository = (*PostgresRepository)(nil)

func NewPostgresRepository(db *sql.DB, logger *zap.Logger) (*PostgresRepository, error) {
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (r *PostgresRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("failed to rollback transaction", zap.Error(rbErr))
			}
		}
	}()

	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Insert product
	query := `
		INSERT INTO products (
			title, slug, description, short_description, price,
			discount_price, sku, inventory_qty, weight, is_published, brand_id,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	err = tx.QueryRowContext(
		ctx, query,
		product.Title, product.Slug, product.Description, product.ShortDescription,
		product.Price, product.DiscountPrice, product.SKU, product.InventoryQty,
		product.Weight, product.IsPublished, product.BrandID, now, now,
	).Scan(&product.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return models.ErrProductAlreadyExists
			}
		}
		r.logger.Error("failed to create product", zap.Error(err))
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Insert images if any
	if len(product.Images) > 0 {
		imageQuery := `
			INSERT INTO product_images (
				product_id, url, alt_text, position, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6)`

		for i := range product.Images {
			img := &product.Images[i]
			img.ProductID = product.ID
			img.CreatedAt = now
			img.UpdatedAt = now

			_, err = tx.ExecContext(ctx, imageQuery,
				img.ProductID, img.URL, img.AltText, img.Position, now, now)
			if err != nil {
				r.logger.Error("failed to create product image", zap.Error(err))
				return fmt.Errorf("failed to create product image: %w", err)
			}
		}
	}

	// Insert categories if any
	if len(product.Categories) > 0 {
		categoryQuery := `
			INSERT INTO product_categories (
				product_id, category_id, created_at
			) VALUES ($1, $2, $3)`

		for _, cat := range product.Categories {
			_, err = tx.ExecContext(ctx, categoryQuery, product.ID, cat.ID, now)
			if err != nil {
				r.logger.Error("failed to create product category association", zap.Error(err))
				return fmt.Errorf("failed to create product category association: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	product := &models.Product{}

	query := `
		SELECT p.*, b.id as brand_id, b.name as brand_name, b.slug as brand_slug,
			   b.description as brand_description, b.created_at as brand_created_at,
			   b.updated_at as brand_updated_at
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		WHERE p.id = $1 AND p.deleted_at IS NULL`

	var brand models.Brand
	var brandCreatedAt, brandUpdatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Title, &product.Slug, &product.Description,
		&product.ShortDescription, &product.Price, &product.DiscountPrice,
		&product.SKU, &product.InventoryQty, &product.Weight,
		&product.IsPublished, &product.CreatedAt, &product.UpdatedAt,
		&product.BrandID,
		&brand.ID, &brand.Name, &brand.Slug, &brand.Description,
		&brandCreatedAt, &brandUpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrProductNotFound
		}
		r.logger.Error("failed to get product", zap.Error(err), zap.String("product_id", id))
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if product.BrandID != nil {
		brand.CreatedAt = brandCreatedAt.Time
		brand.UpdatedAt = brandUpdatedAt.Time
		product.Brand = &brand
	}

	// Get images
	if err := r.getProductImages(ctx, product); err != nil {
		r.logger.Error("failed to get product images", zap.Error(err))
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}

	// Get categories
	if err := r.getProductCategories(ctx, product); err != nil {
		r.logger.Error("failed to get product categories", zap.Error(err))
		return nil, fmt.Errorf("failed to get product categories: %w", err)
	}

	return product, nil
}

func (r *PostgresRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	product := &models.Product{}

	query := `
		SELECT p.*, b.id as brand_id, b.name as brand_name, b.slug as brand_slug,
			   b.description as brand_description, b.created_at as brand_created_at,
			   b.updated_at as brand_updated_at
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		WHERE p.slug = $1 AND p.deleted_at IS NULL`

	var brand models.Brand
	var brandCreatedAt, brandUpdatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&product.ID, &product.Title, &product.Slug, &product.Description,
		&product.ShortDescription, &product.Price, &product.DiscountPrice,
		&product.SKU, &product.InventoryQty, &product.Weight,
		&product.IsPublished, &product.CreatedAt, &product.UpdatedAt,
		&product.BrandID,
		&brand.ID, &brand.Name, &brand.Slug, &brand.Description,
		&brandCreatedAt, &brandUpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrProductNotFound
		}
		r.logger.Error("failed to get product by slug", zap.Error(err), zap.String("slug", slug))
		return nil, fmt.Errorf("failed to get product by slug: %w", err)
	}

	if product.BrandID != nil {
		brand.CreatedAt = brandCreatedAt.Time
		brand.UpdatedAt = brandUpdatedAt.Time
		product.Brand = &brand
	}

	// Get images
	if err := r.getProductImages(ctx, product); err != nil {
		return nil, err
	}

	// Get categories
	if err := r.getProductCategories(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// GetProduct is redundant with GetByID, removing it.

// List implements the ProductRepository interface method.
// Note: The interface defines offset and limit as int, not int32.
func (r *PostgresRepository) List(ctx context.Context, offset, limit int) ([]*models.Product, int, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		r.logger.Error("failed to count products", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	query := `
		SELECT p.*, b.id as brand_id, b.name as brand_name, b.slug as brand_slug,
			   b.description as brand_description, b.created_at as brand_created_at,
			   b.updated_at as brand_updated_at
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		WHERE p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("failed to list products", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		var brand models.Brand
		var brandCreatedAt, brandUpdatedAt sql.NullTime

		err := rows.Scan(
			&product.ID, &product.Title, &product.Slug, &product.Description,
			&product.ShortDescription, &product.Price, &product.DiscountPrice,
			&product.SKU, &product.InventoryQty, &product.Weight,
			&product.IsPublished, &product.CreatedAt, &product.UpdatedAt,
			&product.BrandID,
			&brand.ID, &brand.Name, &brand.Slug, &brand.Description,
			&brandCreatedAt, &brandUpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan product row", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan product row: %w", err)
		}

		if product.BrandID != nil {
			brand.CreatedAt = brandCreatedAt.Time
			brand.UpdatedAt = brandUpdatedAt.Time
			product.Brand = &brand
		}

		if err := r.getProductImages(ctx, product); err != nil {
			r.logger.Error("failed to get product images", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to get product images: %w", err)
		}

		if err := r.getProductCategories(ctx, product); err != nil {
			r.logger.Error("failed to get product categories", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to get product categories: %w", err)
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating product rows", zap.Error(err))
		return nil, 0, fmt.Errorf("error iterating product rows: %w", err)
	}

	return products, total, nil
}

// UpdateProduct implements the ProductRepository interface method.
func (r *PostgresRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("failed to rollback transaction", zap.Error(rbErr))
			}
		}
	}()

	now := time.Now()
	query := `
		UPDATE products SET
			title = $1, slug = $2, description = $3, short_description = $4,
			price = $5, discount_price = $6, sku = $7, inventory_qty = $8,
			weight = $9, is_published = $10, brand_id = $11, updated_at = $12
		WHERE id = $13 AND deleted_at IS NULL`

	result, err := tx.ExecContext(ctx, query,
		product.Title, product.Slug, product.Description, product.ShortDescription,
		product.Price, product.DiscountPrice, product.SKU, product.InventoryQty,
		product.Weight, product.IsPublished, product.BrandID, now, product.ID,
	)
	if err != nil {
		r.logger.Error("failed to update product", zap.Error(err))
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrProductNotFound
	}

	// Update images
	if err := r.updateProductImages(ctx, tx, product, now); err != nil {
		return err
	}

	// Update categories
	if err := r.updateProductCategories(ctx, tx, product, now); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	product.UpdatedAt = now
	return nil
}

// DeleteProduct implements the ProductRepository interface method.
func (r *PostgresRepository) DeleteProduct(ctx context.Context, id string) error {
	query := `UPDATE products SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		r.logger.Error("failed to delete product", zap.Error(err))
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrProductNotFound
	}

	return nil
}

// BeginTx starts a new transaction
func (r *PostgresRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// --- Stubs for ProductRepository methods defined in interface but not yet implemented ---

func (r *PostgresRepository) AddImage(ctx context.Context, image *models.ProductImage) error {
	// TODO: Implement AddImage
	return errors.New("AddImage not implemented")
}

func (r *PostgresRepository) RemoveImage(ctx context.Context, imageID string) error {
	// TODO: Implement RemoveImage
	return errors.New("RemoveImage not implemented")
}

func (r *PostgresRepository) UpdateImagePositions(ctx context.Context, productID string, imageIDs []string) error {
	// TODO: Implement UpdateImagePositions
	return errors.New("UpdateImagePositions not implemented")
}

func (r *PostgresRepository) AddCategories(ctx context.Context, productID string, categoryIDs []string) error {
	// TODO: Implement AddCategories
	return errors.New("AddCategories not implemented")
}

func (r *PostgresRepository) RemoveCategories(ctx context.Context, productID string, categoryIDs []string) error {
	// TODO: Implement RemoveCategories
	return errors.New("RemoveCategories not implemented")
}

// --- Variant-specific methods ---

// GetProductVariants retrieves all variants for a product
func (r *PostgresRepository) GetProductVariants(ctx context.Context, productID string) ([]*models.ProductVariant, error) {
	const query = `
		SELECT
			pv.id, pv.product_id, pv.sku, pv.title, pv.price, pv.discount_price,
			pv.inventory_qty, pv.created_at, pv.updated_at, pv.deleted_at
		FROM product_variants pv
		WHERE pv.product_id = $1 AND pv.deleted_at IS NULL
		ORDER BY pv.created_at
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to query product variants", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to query product variants: %w", err)
	}
	defer rows.Close()

	var variants []*models.ProductVariant
	for rows.Next() {
		var variant models.ProductVariant
		if err := rows.Scan(
			&variant.ID, &variant.ProductID, &variant.SKU, &variant.Title, &variant.Price, &variant.DiscountPrice,
			&variant.InventoryQty, &variant.CreatedAt, &variant.UpdatedAt, &variant.DeletedAt,
		); err != nil {
			r.logger.Error("failed to scan product variant", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product variant: %w", err)
		}

		// Get attributes for this variant
		if err := r.getVariantAttributes(ctx, &variant); err != nil {
			r.logger.Error("failed to get variant attributes", zap.Error(err), zap.String("variant_id", variant.ID))
			return nil, fmt.Errorf("failed to get variant attributes: %w", err)
		}

		// Get images for this variant
		variantImages, err := r.GetVariantImages(ctx, variant.ID)
		if err != nil {
			r.logger.Error("failed to get variant images", zap.Error(err), zap.String("variant_id", variant.ID))
			// Continue even if images fail to load - don't fail the whole query
		} else {
			variant.Images = variantImages
		}

		variants = append(variants, &variant)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating product variant rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating product variants: %w", err)
	}

	return variants, nil
}

// GetVariantAttributes fetches attributes for a specific variant by ID
func (r *PostgresRepository) GetVariantAttributes(ctx context.Context, variantID string) ([]models.VariantAttributeValue, error) {
	const query = `
		SELECT a.name, pva.value
		FROM product_variant_attributes pva
		JOIN attributes a ON pva.attribute_id = a.id AND a.deleted_at IS NULL
		WHERE pva.product_variant_id = $1
		ORDER BY a.name
	`

	rows, err := r.db.QueryContext(ctx, query, variantID)
	if err != nil {
		r.logger.Error("failed to get variant attributes", zap.Error(err), zap.String("variant_id", variantID))
		return nil, fmt.Errorf("failed to get variant attributes: %w", err)
	}
	defer rows.Close()

	var attributes []models.VariantAttributeValue
	for rows.Next() {
		var attr models.VariantAttributeValue
		if err := rows.Scan(&attr.Name, &attr.Value); err != nil {
			r.logger.Error("failed to scan variant attribute", zap.Error(err))
			return nil, fmt.Errorf("failed to scan variant attribute: %w", err)
		}
		attributes = append(attributes, attr)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating variant attributes rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating variant attributes rows: %w", err)
	}

	return attributes, nil
}

// getVariantAttributes fetches attributes for a specific variant
func (r *PostgresRepository) getVariantAttributes(ctx context.Context, variant *models.ProductVariant) error {
	attributes, err := r.GetVariantAttributes(ctx, variant.ID)
	if err != nil {
		return err
	}
	variant.Attributes = attributes
	return nil
}

// CreateVariant creates a new product variant with its attributes
func (r *PostgresRepository) CreateVariant(ctx context.Context, tx *sql.Tx, productID string, variant *models.ProductVariant) error {
	// Check if we need to manage the transaction ourselves
	var manageTx bool
	if tx == nil {
		manageTx = true
		var err error
		tx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			r.logger.Error("failed to begin transaction", zap.Error(err))
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					r.logger.Error("failed to rollback transaction", zap.Error(rbErr))
				}
			}
		}()
	}

	now := time.Now().UTC()
	variant.ProductID = productID
	variant.CreatedAt = now
	variant.UpdatedAt = now

	// Insert the variant
	const variantQuery = `
		INSERT INTO product_variants (
			product_id, sku, title, price, discount_price, inventory_qty, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := tx.QueryRowContext(ctx, variantQuery,
		variant.ProductID, variant.SKU, variant.Title, variant.Price, variant.DiscountPrice,
		variant.InventoryQty, now, now,
	).Scan(&variant.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return models.ErrVariantAlreadyExists
		}
		r.logger.Error("failed to create variant", zap.Error(err))
		return fmt.Errorf("failed to create variant: %w", err)
	}

	// Insert variant attributes if any
	if len(variant.Attributes) > 0 {
		if err := r.createVariantAttributes(ctx, tx, variant); err != nil {
			return err
		}
	}

	// Insert variant images if any
	if len(variant.Images) > 0 {
		imageQuery := `
			INSERT INTO variant_images (
				variant_id, url, alt_text, position, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id`

		for i := range variant.Images {
			img := &variant.Images[i]
			img.VariantID = variant.ID
			img.CreatedAt = now
			img.UpdatedAt = now

			err = tx.QueryRowContext(
				ctx, imageQuery,
				img.VariantID, img.URL, img.AltText, img.Position, now, now,
			).Scan(&img.ID)

			if err != nil {
				r.logger.Error("failed to create variant image", zap.Error(err))
				return fmt.Errorf("failed to create variant image: %w", err)
			}
		}
	}

	// Commit if we're managing the transaction
	if manageTx {
		if err := tx.Commit(); err != nil {
			r.logger.Error("failed to commit transaction", zap.Error(err))
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// createVariantAttributes creates attribute associations for a variant
func (r *PostgresRepository) createVariantAttributes(ctx context.Context, tx *sql.Tx, variant *models.ProductVariant) error {
	// First, ensure all attributes exist in the attributes table
	for _, attr := range variant.Attributes {
		// Check if attribute exists
		var attrID string
		err := tx.QueryRowContext(ctx,
			"SELECT id FROM attributes WHERE name = $1 AND deleted_at IS NULL",
			attr.Name,
		).Scan(&attrID)

		if err != nil {
			if err == sql.ErrNoRows {
				// Create the attribute
				now := time.Now().UTC()
				err = tx.QueryRowContext(ctx,
					"INSERT INTO attributes (name, created_at, updated_at) VALUES ($1, $2, $3) RETURNING id",
					attr.Name, now, now,
				).Scan(&attrID)

				if err != nil {
					r.logger.Error("failed to create attribute", zap.Error(err), zap.String("name", attr.Name))
					return fmt.Errorf("failed to create attribute: %w", err)
				}
			} else {
				r.logger.Error("failed to check attribute existence", zap.Error(err), zap.String("name", attr.Name))
				return fmt.Errorf("failed to check attribute existence: %w", err)
			}
		}

		// Create the variant-attribute association
		_, err = tx.ExecContext(ctx,
			"INSERT INTO product_variant_attributes (product_variant_id, attribute_id, value, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
			variant.ID, attrID, attr.Value, time.Now().UTC(), time.Now().UTC(),
		)

		if err != nil {
			r.logger.Error("failed to create variant attribute", zap.Error(err))
			return fmt.Errorf("failed to create variant attribute: %w", err)
		}
	}

	return nil
}

// UpdateVariant updates an existing product variant and its attributes
func (r *PostgresRepository) UpdateVariant(ctx context.Context, tx *sql.Tx, variant *models.ProductVariant) error {
	// Check if we need to manage the transaction ourselves
	var manageTx bool
	if tx == nil {
		manageTx = true
		var err error
		tx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			r.logger.Error("failed to begin transaction", zap.Error(err))
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					r.logger.Error("failed to rollback transaction", zap.Error(rbErr))
				}
			}
		}()
	}

	now := time.Now().UTC()
	variant.UpdatedAt = now

	// Update the variant
	const variantQuery = `
		UPDATE product_variants SET
			sku = $1, title = $2, price = $3, discount_price = $4,
			inventory_qty = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(ctx, variantQuery,
		variant.SKU, variant.Title, variant.Price, variant.DiscountPrice,
		variant.InventoryQty, now, variant.ID,
	)

	if err != nil {
		r.logger.Error("failed to update variant", zap.Error(err), zap.String("variant_id", variant.ID))
		return fmt.Errorf("failed to update variant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrVariantNotFound
	}

	// Delete existing attributes and recreate them
	_, err = tx.ExecContext(ctx, "DELETE FROM product_variant_attributes WHERE product_variant_id = $1", variant.ID)
	if err != nil {
		r.logger.Error("failed to delete variant attributes", zap.Error(err), zap.String("variant_id", variant.ID))
		return fmt.Errorf("failed to delete variant attributes: %w", err)
	}

	// Insert new attributes if any
	if len(variant.Attributes) > 0 {
		if err := r.createVariantAttributes(ctx, tx, variant); err != nil {
			return err
		}
	}

	// Commit if we're managing the transaction
	if manageTx {
		if err := tx.Commit(); err != nil {
			r.logger.Error("failed to commit transaction", zap.Error(err))
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// DeleteVariant performs a soft delete of a product variant
func (r *PostgresRepository) DeleteVariant(ctx context.Context, tx *sql.Tx, variantID string) error {
	// Check if we need to manage the transaction ourselves
	var manageTx bool
	if tx == nil {
		manageTx = true
		var err error
		tx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			r.logger.Error("failed to begin transaction", zap.Error(err))
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					r.logger.Error("failed to rollback transaction", zap.Error(rbErr))
				}
			}
		}()
	}

	// Soft delete the variant
	const query = `
		UPDATE product_variants
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(ctx, query, time.Now().UTC(), variantID)
	if err != nil {
		r.logger.Error("failed to delete variant", zap.Error(err), zap.String("variant_id", variantID))
		return fmt.Errorf("failed to delete variant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrVariantNotFound
	}

	// Commit if we're managing the transaction
	if manageTx {
		if err := tx.Commit(); err != nil {
			r.logger.Error("failed to commit transaction", zap.Error(err))
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// --- Stubs for BrandRepository ---

func (r *PostgresRepository) CreateBrand(ctx context.Context, brand *models.Brand) error {
	now := time.Now()
	brand.CreatedAt = now
	brand.UpdatedAt = now

	query := `
		INSERT INTO brands (
			name, slug, description, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.db.QueryRowContext(
		ctx, query,
		brand.Name, brand.Slug, brand.Description, now, now,
	).Scan(&brand.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return fmt.Errorf("brand already exists")
			}
		}
		r.logger.Error("failed to create brand", zap.Error(err))
		return fmt.Errorf("failed to create brand: %w", err)
	}

	return nil
}

func (r *PostgresRepository) UpdateBrand(ctx context.Context, brand *models.Brand) error {
	// TODO: Implement UpdateBrand
	return errors.New("UpdateBrand not implemented")
}

func (r *PostgresRepository) DeleteBrand(ctx context.Context, id string) error {
	// TODO: Implement DeleteBrand
	return errors.New("DeleteBrand not implemented")
}

// GetBrandByID implements the BrandRepository interface method.
func (r *PostgresRepository) GetBrandByID(ctx context.Context, id string) (*models.Brand, error) {
	brand := &models.Brand{}
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM brands
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&brand.ID, &brand.Name, &brand.Slug, &brand.Description,
		&brand.CreatedAt, &brand.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("brand not found")
		}
		r.logger.Error("failed to get brand by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get brand by ID: %w", err)
	}

	return brand, nil
}

// GetBrandBySlug implements the BrandRepository interface method.
func (r *PostgresRepository) GetBrandBySlug(ctx context.Context, slug string) (*models.Brand, error) {
	brand := &models.Brand{}
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM brands
		WHERE slug = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&brand.ID, &brand.Name, &brand.Slug, &brand.Description,
		&brand.CreatedAt, &brand.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("brand not found")
		}
		r.logger.Error("failed to get brand by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get brand by slug: %w", err)
	}

	return brand, nil
}

// ListBrands implements the BrandRepository interface method.
func (r *PostgresRepository) ListBrands(ctx context.Context, offset, limit int) ([]*models.Brand, int, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM brands WHERE deleted_at IS NULL`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		r.logger.Error("failed to count brands", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count brands: %w", err)
	}

	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM brands
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("failed to list brands", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list brands: %w", err)
	}
	defer rows.Close()

	var brands []*models.Brand
	for rows.Next() {
		brand := &models.Brand{}
		err := rows.Scan(
			&brand.ID, &brand.Name, &brand.Slug, &brand.Description,
			&brand.CreatedAt, &brand.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan brand row", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan brand row: %w", err)
		}
		brands = append(brands, brand)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating brand rows", zap.Error(err))
		return nil, 0, fmt.Errorf("error iterating brand rows: %w", err)
	}

	return brands, total, nil
}

// --- Stubs for CategoryRepository ---

func (r *PostgresRepository) CreateCategory(ctx context.Context, category *models.Category) error {
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	query := `
		INSERT INTO categories (
			name, slug, description, parent_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRowContext(
		ctx, query,
		category.Name, category.Slug, category.Description,
		category.ParentID, now, now,
	).Scan(&category.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return fmt.Errorf("category already exists")
			}
		}
		r.logger.Error("failed to create category", zap.Error(err))
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (r *PostgresRepository) UpdateCategory(ctx context.Context, category *models.Category) error {
	category.UpdatedAt = time.Now()

	query := `
        UPDATE categories
        SET name = $1,
            slug = $2,
            description = $3,
            parent_id = $4,
            updated_at = $5
        WHERE id = $6 AND deleted_at IS NULL
        RETURNING id`

	result, err := r.db.ExecContext(ctx, query,
		category.Name,
		category.Slug,
		category.Description,
		category.ParentID,
		category.UpdatedAt,
		category.ID,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return fmt.Errorf("category with this slug already exists")
			}
		}
		r.logger.Error("failed to update category", zap.Error(err))
		return fmt.Errorf("failed to update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *PostgresRepository) DeleteCategory(ctx context.Context, id string) error {
	// Start a transaction to handle related records
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, update any child categories to remove their parent_id
	updateChildrenQuery := `
        UPDATE categories
        SET parent_id = NULL,
            updated_at = $1
        WHERE parent_id = $2`

	_, err = tx.ExecContext(ctx, updateChildrenQuery, time.Now(), id)
	if err != nil {
		r.logger.Error("failed to update child categories", zap.Error(err))
		return fmt.Errorf("failed to update child categories: %w", err)
	}

	// Remove category associations from products
	deleteAssociationsQuery := `
        DELETE FROM product_categories
        WHERE category_id = $1`

	_, err = tx.ExecContext(ctx, deleteAssociationsQuery, id)
	if err != nil {
		r.logger.Error("failed to delete category associations", zap.Error(err))
		return fmt.Errorf("failed to delete category associations: %w", err)
	}

	// Soft delete the category
	deleteQuery := `
        UPDATE categories
        SET deleted_at = $1
        WHERE id = $2 AND deleted_at IS NULL`

	result, err := tx.ExecContext(ctx, deleteQuery, time.Now(), id)
	if err != nil {
		r.logger.Error("failed to delete category", zap.Error(err))
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	category := &models.Category{}
	query := `
        SELECT
            c.id,
            c.name,
            c.slug,
            c.description,
            c.parent_id,
            c.created_at,
            c.updated_at,
            p.name as parent_name
        FROM categories c
        LEFT JOIN categories p ON c.parent_id = p.id
        WHERE c.id = $1 AND c.deleted_at IS NULL`

	var parentName sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.ParentID,
		&category.CreatedAt,
		&category.UpdatedAt,
		&parentName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		r.logger.Error("failed to get category by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get category by ID: %w", err)
	}

	if parentName.Valid {
		category.ParentName = parentName.String
	}

	return category, nil
}

func (r *PostgresRepository) GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	category := &models.Category{}
	query := `
        SELECT
            c.id,
            c.name,
            c.slug,
            c.description,
            c.parent_id,
            c.created_at,
            c.updated_at,
            p.name as parent_name
        FROM categories c
        LEFT JOIN categories p ON c.parent_id = p.id
        WHERE c.slug = $1 AND c.deleted_at IS NULL`

	var parentName sql.NullString
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.ParentID,
		&category.CreatedAt,
		&category.UpdatedAt,
		&parentName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		r.logger.Error("failed to get category by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	if parentName.Valid {
		category.ParentName = parentName.String
	}

	return category, nil
}

func (r *PostgresRepository) ListCategories(ctx context.Context, offset, limit int) ([]*models.Category, int, error) {
	// First, get total count
	var total int
	countQuery := `
        SELECT COUNT(*)
        FROM categories
        WHERE deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		r.logger.Error("failed to get total category count", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get total category count: %w", err)
	}

	// Then get paginated results
	query := `
        SELECT
            c.id,
            c.name,
            c.slug,
            c.description,
            c.parent_id,
            c.created_at,
            c.updated_at,
            p.name as parent_name
        FROM categories c
        LEFT JOIN categories p ON c.parent_id = p.id
        WHERE c.deleted_at IS NULL
        ORDER BY c.created_at DESC
        LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("failed to list categories", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		category := &models.Category{}
		var parentName sql.NullString
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.ParentID,
			&category.CreatedAt,
			&category.UpdatedAt,
			&parentName,
		)
		if err != nil {
			r.logger.Error("failed to scan category row", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan category row: %w", err)
		}
		if parentName.Valid {
			category.ParentName = parentName.String
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating category rows", zap.Error(err))
		return nil, 0, fmt.Errorf("error iterating category rows: %w", err)
	}

	return categories, total, nil
}

func (r *PostgresRepository) GetChildren(ctx context.Context, parentID string) ([]*models.Category, error) {
	query := `
        SELECT
            id,
            name,
            slug,
            description,
            parent_id,
            created_at,
            updated_at
        FROM categories
        WHERE parent_id = $1 AND deleted_at IS NULL
        ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		r.logger.Error("failed to get child categories", zap.Error(err))
		return nil, fmt.Errorf("failed to get child categories: %w", err)
	}
	defer rows.Close()

	var children []*models.Category
	for rows.Next() {
		child := &models.Category{}
		err := rows.Scan(
			&child.ID,
			&child.Name,
			&child.Slug,
			&child.Description,
			&child.ParentID,
			&child.CreatedAt,
			&child.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan child category", zap.Error(err))
			return nil, fmt.Errorf("failed to scan child category: %w", err)
		}
		children = append(children, child)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating child categories", zap.Error(err))
		return nil, fmt.Errorf("error iterating child categories: %w", err)
	}

	return children, nil
}

// Ping checks database connectivity (moved from repository.go)
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Helper methods

func (r *PostgresRepository) getProductImages(ctx context.Context, product *models.Product) error {
	query := `
		SELECT id, url, alt_text, position, created_at, updated_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY position`

	rows, err := r.db.QueryContext(ctx, query, product.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var img models.ProductImage
		err := rows.Scan(
			&img.ID, &img.URL, &img.AltText, &img.Position,
			&img.CreatedAt, &img.UpdatedAt,
		)
		if err != nil {
			return err
		}
		img.ProductID = product.ID
		product.Images = append(product.Images, img)
	}
	return rows.Err()
}

func (r *PostgresRepository) getProductCategories(ctx context.Context, product *models.Product) error {
	query := `
		SELECT c.id, c.name, c.slug, c.description, c.parent_id,
			   c.created_at, c.updated_at
		FROM categories c
		JOIN product_categories pc ON c.id = pc.category_id
		WHERE pc.product_id = $1`

	rows, err := r.db.QueryContext(ctx, query, product.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cat models.Category
		err := rows.Scan(
			&cat.ID, &cat.Name, &cat.Slug, &cat.Description,
			&cat.ParentID, &cat.CreatedAt, &cat.UpdatedAt,
		)
		if err != nil {
			return err
		}
		product.Categories = append(product.Categories, cat)
	}
	return rows.Err()
}

func (r *PostgresRepository) updateProductImages(ctx context.Context, tx *sql.Tx, product *models.Product, now time.Time) error {
	// Delete existing images
	_, err := tx.ExecContext(ctx, "DELETE FROM product_images WHERE product_id = $1", product.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing images: %w", err)
	}

	// Insert new images
	if len(product.Images) > 0 {
		imageQuery := `
			INSERT INTO product_images (
				product_id, url, alt_text, position, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6)`

		for i := range product.Images {
			img := &product.Images[i]
			img.ProductID = product.ID
			img.UpdatedAt = now
			if img.CreatedAt.IsZero() {
				img.CreatedAt = now
			}

			_, err = tx.ExecContext(ctx, imageQuery,
				img.ProductID, img.URL, img.AltText, img.Position,
				img.CreatedAt, img.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to create product image: %w", err)
			}
		}
	}
	return nil
}

func (r *PostgresRepository) updateProductCategories(ctx context.Context, tx *sql.Tx, product *models.Product, now time.Time) error {
	// Delete existing category associations
	_, err := tx.ExecContext(ctx, "DELETE FROM product_categories WHERE product_id = $1", product.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing category associations: %w", err)
	}

	// Insert new category associations
	if len(product.Categories) > 0 {
		categoryQuery := `
			INSERT INTO product_categories (product_id, category_id, created_at)
			VALUES ($1, $2, $3)`

		for _, cat := range product.Categories {
			_, err = tx.ExecContext(ctx, categoryQuery, product.ID, cat.ID, now)
			if err != nil {
				return fmt.Errorf("failed to create product category association: %w", err)
			}
		}
	}
	return nil
}

// AddProductAttribute adds a new attribute to a product
func (r *PostgresRepository) AddProductAttribute(ctx context.Context, attribute *models.ProductAttribute) error {
	now := time.Now()
	attribute.CreatedAt = now
	attribute.UpdatedAt = now

	query := `
		INSERT INTO product_attributes (
			product_id, name, value, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.db.QueryRowContext(
		ctx, query,
		attribute.ProductID, attribute.Name, attribute.Value, now, now,
	).Scan(&attribute.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return fmt.Errorf("attribute already exists for this product")
			}
		}
		r.logger.Error("failed to add product attribute", zap.Error(err))
		return fmt.Errorf("failed to add product attribute: %w", err)
	}

	return nil
}

// AddProductDiscount adds a new discount to a product
func (r *PostgresRepository) AddProductDiscount(ctx context.Context, discount *models.ProductDiscount) error {
	now := time.Now()
	discount.CreatedAt = now
	discount.UpdatedAt = now

	query := `
		INSERT INTO product_discounts (
			product_id, discount_type, value, expires_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRowContext(
		ctx, query,
		discount.ProductID, discount.Type, discount.Value, discount.ExpiresAt, now, now,
	).Scan(&discount.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return fmt.Errorf("discount already exists for this product")
			}
		}
		r.logger.Error("failed to add product discount", zap.Error(err))
		return fmt.Errorf("failed to add product discount: %w", err)
	}

	return nil
}

// AddProductSpecification adds a new specification to a product
func (r *PostgresRepository) AddProductSpecification(ctx context.Context, spec *models.ProductSpecification) error {
	now := time.Now()
	spec.CreatedAt = now
	spec.UpdatedAt = now

	query := `
		INSERT INTO product_specifications (
			product_id, name, value, unit, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRowContext(
		ctx, query,
		spec.ProductID, spec.Name, spec.Value, spec.Unit, now, now,
	).Scan(&spec.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return fmt.Errorf("specification already exists for this product")
			}
		}
		r.logger.Error("failed to add product specification", zap.Error(err))
		return fmt.Errorf("failed to add product specification: %w", err)
	}

	return nil
}

// AddProductTag adds a new tag to a product
func (r *PostgresRepository) AddProductTag(ctx context.Context, tag *models.ProductTag) error {
	now := time.Now()
	tag.CreatedAt = now
	tag.UpdatedAt = now

	query := `
		INSERT INTO product_tags (
			product_id, tag, created_at, updated_at
		) VALUES ($1, $2, $3, $4)
		RETURNING id`

	err := r.db.QueryRowContext(
		ctx, query,
		tag.ProductID, tag.Tag, now, now,
	).Scan(&tag.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return fmt.Errorf("tag already exists for this product")
			}
		}
		r.logger.Error("failed to add product tag", zap.Error(err))
		return fmt.Errorf("failed to add product tag: %w", err)
	}

	return nil
}

// GetInventoryLocations retrieves all inventory locations for a product
func (r *PostgresRepository) GetInventoryLocations(ctx context.Context, productID string) ([]models.InventoryLocation, error) {
	query := `
		SELECT id, product_id, warehouse_id, available_qty, created_at, updated_at
		FROM product_inventory_locations
		WHERE product_id = $1
		ORDER BY warehouse_id`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get inventory locations", zap.Error(err))
		return nil, fmt.Errorf("failed to get inventory locations: %w", err)
	}
	defer rows.Close()

	var locations []models.InventoryLocation
	for rows.Next() {
		var location models.InventoryLocation
		err := rows.Scan(
			&location.ID, &location.ProductID, &location.WarehouseID, &location.AvailableQty,
			&location.CreatedAt, &location.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan inventory location", zap.Error(err))
			return nil, fmt.Errorf("failed to scan inventory location: %w", err)
		}
		locations = append(locations, location)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating inventory location rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating inventory location rows: %w", err)
	}

	return locations, nil
}

// GetProductAttributes retrieves all attributes for a product
func (r *PostgresRepository) GetProductAttributes(ctx context.Context, productID string) ([]models.ProductAttribute, error) {
	query := `
		SELECT id, product_id, name, value, created_at, updated_at
		FROM product_attributes
		WHERE product_id = $1
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get product attributes", zap.Error(err))
		return nil, fmt.Errorf("failed to get product attributes: %w", err)
	}
	defer rows.Close()

	var attributes []models.ProductAttribute
	for rows.Next() {
		var attr models.ProductAttribute
		err := rows.Scan(
			&attr.ID, &attr.ProductID, &attr.Name, &attr.Value,
			&attr.CreatedAt, &attr.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan product attribute", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product attribute: %w", err)
		}
		attributes = append(attributes, attr)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating product attribute rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating product attribute rows: %w", err)
	}

	return attributes, nil
}

// GetProductDiscounts retrieves all discounts for a product
func (r *PostgresRepository) GetProductDiscounts(ctx context.Context, productID string) ([]models.ProductDiscount, error) {
	query := `
		SELECT id, product_id, discount_type, value, expires_at, created_at, updated_at
		FROM product_discounts
		WHERE product_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get product discounts", zap.Error(err))
		return nil, fmt.Errorf("failed to get product discounts: %w", err)
	}
	defer rows.Close()

	var discounts []models.ProductDiscount
	for rows.Next() {
		var discount models.ProductDiscount
		err := rows.Scan(
			&discount.ID, &discount.ProductID, &discount.Type, &discount.Value,
			&discount.ExpiresAt, &discount.CreatedAt, &discount.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan product discount", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product discount: %w", err)
		}
		discounts = append(discounts, discount)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating product discount rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating product discount rows: %w", err)
	}

	return discounts, nil
}

// GetProductSEO retrieves SEO information for a product
func (r *PostgresRepository) GetProductSEO(ctx context.Context, productID string) (*models.ProductSEO, error) {
	query := `
		SELECT id, product_id, meta_title, meta_description, keywords, tags, created_at, updated_at
		FROM product_seo
		WHERE product_id = $1`

	var seo models.ProductSEO
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&seo.ID, &seo.ProductID, &seo.MetaTitle, &seo.MetaDescription,
		&seo.Keywords, &seo.Tags, &seo.CreatedAt, &seo.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("failed to get product SEO", zap.Error(err))
		return nil, fmt.Errorf("failed to get product SEO: %w", err)
	}

	return &seo, nil
}

// GetProductShipping retrieves shipping information for a product
func (r *PostgresRepository) GetProductShipping(ctx context.Context, productID string) (*models.ProductShipping, error) {
	query := `
		SELECT id, product_id, free_shipping, estimated_days, express_available, created_at, updated_at
		FROM product_shipping
		WHERE product_id = $1`

	var shipping models.ProductShipping
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&shipping.ID, &shipping.ProductID, &shipping.FreeShipping, &shipping.EstimatedDays,
		&shipping.ExpressAvailable, &shipping.CreatedAt, &shipping.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No shipping info found
		}
		r.logger.Error("failed to get product shipping", zap.Error(err))
		return nil, fmt.Errorf("failed to get product shipping: %w", err)
	}

	return &shipping, nil
}

// GetProductSpecifications retrieves all specifications for a product
func (r *PostgresRepository) GetProductSpecifications(ctx context.Context, productID string) ([]models.ProductSpecification, error) {
	query := `
		SELECT id, product_id, name, value, unit, created_at, updated_at
		FROM product_specifications
		WHERE product_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get product specifications", zap.Error(err))
		return nil, fmt.Errorf("failed to get product specifications: %w", err)
	}
	defer rows.Close()

	var specs []models.ProductSpecification
	for rows.Next() {
		var spec models.ProductSpecification
		err := rows.Scan(
			&spec.ID, &spec.ProductID, &spec.Name, &spec.Value,
			&spec.Unit, &spec.CreatedAt, &spec.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan product specification", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product specification: %w", err)
		}
		specs = append(specs, spec)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating product specification rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating product specification rows: %w", err)
	}

	return specs, nil
}

// GetProductTags retrieves all tags for a product
func (r *PostgresRepository) GetProductTags(ctx context.Context, productID string) ([]models.ProductTag, error) {
	query := `
		SELECT id, product_id, tag, created_at, updated_at
		FROM product_tags
		WHERE product_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get product tags", zap.Error(err))
		return nil, fmt.Errorf("failed to get product tags: %w", err)
	}
	defer rows.Close()

	var tags []models.ProductTag
	for rows.Next() {
		var tag models.ProductTag
		err := rows.Scan(
			&tag.ID, &tag.ProductID, &tag.Tag,
			&tag.CreatedAt, &tag.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan product tag", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating product tag rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating product tag rows: %w", err)
	}

	return tags, nil
}

// RemoveInventoryLocation removes a product's inventory location
func (r *PostgresRepository) RemoveInventoryLocation(ctx context.Context, productID, warehouseID string) error {
	query := `
		DELETE FROM inventory_locations
		WHERE product_id = $1 AND warehouse_id = $2`

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

// RemoveProductAttribute removes a product attribute
func (r *PostgresRepository) RemoveProductAttribute(ctx context.Context, attributeID string) error {
	query := `DELETE FROM product_attributes WHERE id = $1`

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
		return fmt.Errorf("product attribute not found")
	}

	return nil
}

// RemoveProductDiscount removes a product discount
func (r *PostgresRepository) RemoveProductDiscount(ctx context.Context, discountID string) error {
	query := `DELETE FROM product_discounts WHERE id = $1`

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
		return fmt.Errorf("product discount not found")
	}

	return nil
}

// RemoveProductSpecification removes a product specification
func (r *PostgresRepository) RemoveProductSpecification(ctx context.Context, specID string) error {
	query := `DELETE FROM product_specifications WHERE id = $1`

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
		return fmt.Errorf("product specification not found")
	}

	return nil
}

// RemoveProductTag removes a product tag
func (r *PostgresRepository) RemoveProductTag(ctx context.Context, productID, tag string) error {
	query := `DELETE FROM product_tags WHERE product_id = $1 AND tag = $2`

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
		return fmt.Errorf("product tag not found")
	}

	return nil
}

// UpdateProductAttribute updates an existing product attribute
func (r *PostgresRepository) UpdateProductAttribute(ctx context.Context, attribute *models.ProductAttribute) error {
	now := time.Now()
	attribute.UpdatedAt = now

	query := `
		UPDATE product_attributes
		SET name = $1, value = $2, updated_at = $3
		WHERE id = $4`

	result, err := r.db.ExecContext(ctx, query,
		attribute.Name, attribute.Value, now, attribute.ID)
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
		return fmt.Errorf("product attribute not found")
	}

	return nil
}

// UpdateProductDiscount updates an existing product discount
func (r *PostgresRepository) UpdateProductDiscount(ctx context.Context, discount *models.ProductDiscount) error {
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

func (r *PostgresRepository) UpdateProductSpecification(ctx context.Context, spec *models.ProductSpecification) error {
	now := time.Now().UTC()
	spec.UpdatedAt = now

	query := `
		UPDATE product_specifications
		SET name = $1, value = $2, unit = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, spec.Name, spec.Value, spec.Unit, now, spec.ID)
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

func (r *PostgresRepository) UpsertInventoryLocation(ctx context.Context, location *models.InventoryLocation) error {
	now := time.Now().UTC()
	location.UpdatedAt = now

	query := `
		INSERT INTO product_inventory_locations (
			product_id, warehouse_id, available_qty, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (product_id, warehouse_id)
		DO UPDATE SET
			available_qty = EXCLUDED.available_qty,
			updated_at = EXCLUDED.updated_at
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		location.ProductID,
		location.WarehouseID,
		location.AvailableQty,
		now,
		now,
	).Scan(&location.ID)

	if err != nil {
		r.logger.Error("failed to upsert inventory location", zap.Error(err))
		return fmt.Errorf("failed to upsert inventory location: %w", err)
	}

	location.CreatedAt = now
	return nil
}

func (r *PostgresRepository) UpsertProductSEO(ctx context.Context, seo *models.ProductSEO) error {
	now := time.Now().UTC()
	seo.UpdatedAt = now

	query := `
		INSERT INTO product_seo (
			product_id, meta_title, meta_description, keywords, tags, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (product_id)
		DO UPDATE SET
			meta_title = EXCLUDED.meta_title,
			meta_description = EXCLUDED.meta_description,
			keywords = EXCLUDED.keywords,
			tags = EXCLUDED.tags,
			updated_at = EXCLUDED.updated_at
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		seo.ProductID,
		seo.MetaTitle,
		seo.MetaDescription,
		seo.Keywords,
		seo.Tags,
		now,
		now,
	).Scan(&seo.ID)

	if err != nil {
		r.logger.Error("failed to upsert product SEO", zap.Error(err))
		return fmt.Errorf("failed to upsert product SEO: %w", err)
	}

	seo.CreatedAt = now
	return nil
}

func (r *PostgresRepository) UpsertProductShipping(ctx context.Context, shipping *models.ProductShipping) error {
	now := time.Now().UTC()
	shipping.UpdatedAt = now

	query := `
		INSERT INTO product_shipping (
			product_id, free_shipping, estimated_days, express_available, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (product_id)
		DO UPDATE SET
			free_shipping = EXCLUDED.free_shipping,
			estimated_days = EXCLUDED.estimated_days,
			express_available = EXCLUDED.express_available,
			updated_at = EXCLUDED.updated_at
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		shipping.ProductID,
		shipping.FreeShipping,
		shipping.EstimatedDays,
		shipping.ExpressAvailable,
		now,
		now,
	).Scan(&shipping.ID)

	if err != nil {
		r.logger.Error("failed to upsert product shipping", zap.Error(err))
		return fmt.Errorf("failed to upsert product shipping: %w", err)
	}

	shipping.CreatedAt = now
	return nil
}

// AddVariantImage adds a new image to a variant
func (r *PostgresRepository) AddVariantImage(ctx context.Context, image *models.VariantImage) error {
	query := `
		INSERT INTO variant_images (id, variant_id, url, alt_text, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	image.CreatedAt = now
	image.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		image.ID,
		image.VariantID,
		image.URL,
		image.AltText,
		image.Position,
		image.CreatedAt,
		image.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to add variant image", zap.Error(err))
		return fmt.Errorf("failed to add variant image: %w", err)
	}

	return nil
}

// GetVariantImages gets all images for a variant
func (r *PostgresRepository) GetVariantImages(ctx context.Context, variantID string) ([]models.VariantImage, error) {
	query := `
		SELECT id, variant_id, url, alt_text, position, created_at, updated_at
		FROM variant_images
		WHERE variant_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, variantID)
	if err != nil {
		r.logger.Error("failed to get variant images", zap.Error(err))
		return nil, fmt.Errorf("failed to get variant images: %w", err)
	}
	defer rows.Close()

	var images []models.VariantImage
	for rows.Next() {
		var image models.VariantImage
		err := rows.Scan(
			&image.ID,
			&image.VariantID,
			&image.URL,
			&image.AltText,
			&image.Position,
			&image.CreatedAt,
			&image.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan variant image", zap.Error(err))
			return nil, fmt.Errorf("failed to scan variant image: %w", err)
		}
		images = append(images, image)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating variant images", zap.Error(err))
		return nil, fmt.Errorf("error iterating variant images: %w", err)
	}

	return images, nil
}

// UpdateVariantImage updates an existing variant image
func (r *PostgresRepository) UpdateVariantImage(ctx context.Context, image *models.VariantImage) error {
	query := `
		UPDATE variant_images
		SET url = $1, alt_text = $2, position = $3, updated_at = $4
		WHERE id = $5
	`

	image.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		image.URL,
		image.AltText,
		image.Position,
		image.UpdatedAt,
		image.ID,
	)

	if err != nil {
		r.logger.Error("failed to update variant image", zap.Error(err))
		return fmt.Errorf("failed to update variant image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("variant image not found: %s", image.ID)
	}

	return nil
}

// DeleteVariantImage deletes a variant image
func (r *PostgresRepository) DeleteVariantImage(ctx context.Context, id string) error {
	query := `
		DELETE FROM variant_images
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete variant image", zap.Error(err))
		return fmt.Errorf("failed to delete variant image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("variant image not found: %s", id)
	}

	return nil
}
