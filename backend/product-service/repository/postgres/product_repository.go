// repository/postgres/product_repository.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/louai60/e-commerce_project/backend/product-service/models"
)

type ProductRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewProductRepository(db *sql.DB, logger *zap.Logger) *ProductRepository {
	return &ProductRepository{
		db:     db,
		logger: logger,
	}
}

// GetProduct retrieves a product by ID with its core details, brand, categories, images, and variants.
func (r *ProductRepository) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	const query = `
		SELECT
			p.id, p.title, p.slug, p.description, p.short_description,
			p.weight, p.is_published, p.created_at, p.updated_at, p.deleted_at,
			p.brand_id, p.default_variant_id, p.inventory_status,
			b.id, b.name, b.slug, b.description, b.created_at, b.updated_at, b.deleted_at
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id AND b.deleted_at IS NULL
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`

	product := &models.Product{}
	var brandID, defaultVariantID sql.NullString
	var brand models.Brand
	var brandCreatedAt, brandUpdatedAt sql.NullTime

	// Note: Scanning product.DeletedAt which is *time.Time
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Title, &product.Slug, &product.Description, &product.ShortDescription,
		&product.Weight, &product.IsPublished, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
		&brandID, &defaultVariantID, &product.InventoryStatus,
		&brand.ID, &brand.Name, &brand.Slug, &brand.Description, &brandCreatedAt, &brandUpdatedAt, &brand.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrProductNotFound
		}
		r.logger.Error("failed to get product", zap.Error(err), zap.String("product_id", id))
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Assign scanned nullable fields
	if brandID.Valid {
		product.BrandID = &brandID.String
		brand.CreatedAt = brandCreatedAt.Time
		brand.UpdatedAt = brandUpdatedAt.Time
		product.Brand = &brand
	}
	if defaultVariantID.Valid {
		product.DefaultVariantID = &defaultVariantID.String
	}

	// Get associated data
	var errs []error
	errs = append(errs, r.getProductImages(ctx, product))
	errs = append(errs, r.getProductCategories(ctx, product))
	errs = append(errs, r.getProductVariantsAndAttributes(ctx, product)) // Fetch variants

	for _, e := range errs {
		if e != nil {
			// Log the specific error
			r.logger.Error("failed to get product associations", zap.Error(e), zap.String("product_id", id))
			// Return a generic error or the first error encountered
			return nil, fmt.Errorf("failed to get product associations: %w", e)
		}
	}

	return product, nil
}

// getProductVariantsAndAttributes fetches variants and their attributes for a product
func (r *ProductRepository) getProductVariantsAndAttributes(ctx context.Context, product *models.Product) error {
	const query = `
		SELECT
			pv.id, pv.product_id, pv.sku, pv.title, pv.price, pv.discount_price,
			pv.inventory_qty, pv.created_at, pv.updated_at, pv.deleted_at,
			a.id, a.name, pva.value
		FROM product_variants pv
		LEFT JOIN product_variant_attributes pva ON pv.id = pva.product_variant_id
		LEFT JOIN attributes a ON pva.attribute_id = a.id AND a.deleted_at IS NULL
		WHERE pv.product_id = $1 AND pv.deleted_at IS NULL
		ORDER BY pv.created_at, a.name -- Consistent ordering
	`

	rows, err := r.db.QueryContext(ctx, query, product.ID)
	if err != nil {
		r.logger.Error("failed to query product variants and attributes", zap.Error(err), zap.String("product_id", product.ID))
		return fmt.Errorf("failed to query product variants: %w", err)
	}
	defer rows.Close()

	variantsMap := make(map[string]*models.ProductVariant)
	// Need to keep track of the order variants appear in the query result
	variantOrder := []string{}

	for rows.Next() {
		var variant models.ProductVariant
		var attributeID, attributeName, attributeValue sql.NullString
		// Scan variant's DeletedAt as well
		if err := rows.Scan(
			&variant.ID, &variant.ProductID, &variant.SKU, &variant.Title, &variant.Price, &variant.DiscountPrice,
			&variant.InventoryQty, &variant.CreatedAt, &variant.UpdatedAt, &variant.DeletedAt,
			&attributeID, &attributeName, &attributeValue,
		); err != nil {
			r.logger.Error("failed to scan product variant row", zap.Error(err))
			return fmt.Errorf("failed to scan product variant: %w", err)
		}

		// Check if we've seen this variant before
		_, found := variantsMap[variant.ID]
		if !found {
			variantsMap[variant.ID] = &variant              // Store pointer to the variant
			variantOrder = append(variantOrder, variant.ID) // Keep track of order
		}

		// Add attribute if it exists for this row
		if attributeID.Valid && attributeName.Valid && attributeValue.Valid {
			attr := models.VariantAttributeValue{
				Name:  attributeName.String,
				Value: attributeValue.String,
			}
			// Append attribute to the variant stored in the map
			variantsMap[variant.ID].Attributes = append(variantsMap[variant.ID].Attributes, attr)
		}
	}

	// Reconstruct the product.Variants slice in the correct order
	product.Variants = []models.ProductVariant{} // Clear existing (if any)
	for _, variantID := range variantOrder {
		if v, ok := variantsMap[variantID]; ok {
			product.Variants = append(product.Variants, *v) // Append dereferenced variant
		}
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating product variant rows", zap.Error(err))
		return fmt.Errorf("error iterating product variants: %w", err)
	}

	return nil
}

// getProductImages fetches all images associated with a product
func (r *ProductRepository) getProductImages(ctx context.Context, product *models.Product) error {
	const query = `
		SELECT id, product_id, url, alt_text, position, created_at, updated_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, product.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var img models.ProductImage
		if err := rows.Scan(
			&img.ID, &img.ProductID, &img.URL, &img.AltText, &img.Position,
			&img.CreatedAt, &img.UpdatedAt,
		); err != nil {
			return err
		}
		product.Images = append(product.Images, img)
	}

	return rows.Err()
}

// getProductCategories fetches all categories associated with a product
func (r *ProductRepository) getProductCategories(ctx context.Context, product *models.Product) error {
	const query = `
		SELECT c.id, c.name, c.slug, c.description, c.parent_id, c.created_at, c.updated_at, c.deleted_at
		FROM categories c
		JOIN product_categories pc ON c.id = pc.category_id
		WHERE pc.product_id = $1 AND c.deleted_at IS NULL
	` // Added c.deleted_at to SELECT and WHERE clause

	rows, err := r.db.QueryContext(ctx, query, product.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cat models.Category
		var parentID sql.NullString
		// Added &cat.DeletedAt to scan the soft delete timestamp
		if err := rows.Scan(
			&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &parentID,
			&cat.CreatedAt, &cat.UpdatedAt, &cat.DeletedAt,
		); err != nil {
			return err
		}
		if parentID.Valid {
			cat.ParentID = &parentID.String
		}
		product.Categories = append(product.Categories, cat)
	}

	return rows.Err()
}

// ListProducts retrieves paginated products with optional filters (simplified for Phase 1)
// TODO: Enhance filtering/sorting based on variants/attributes in Phase 5
func (r *ProductRepository) ListProducts(ctx context.Context, filters models.ProductFilters) ([]*models.Product, int64, error) {
	// Build base query - Select only core product fields
	baseQuery := `
		SELECT
			p.id, p.title, p.slug, p.description, p.short_description,
			p.weight, p.is_published, p.created_at, p.updated_at, p.deleted_at,
			p.brand_id, p.default_variant_id, p.inventory_status,
			b.id, b.name, b.slug, b.description, b.created_at, b.updated_at, b.deleted_at
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id AND b.deleted_at IS NULL
		WHERE p.deleted_at IS NULL
	`

	// Add filtering conditions (only category filter remains relevant for now)
	var args []interface{}
	var conditions []string

	if filters.Category != "" {
		conditions = append(conditions, `
			EXISTS (
				SELECT 1 FROM product_categories pc
				JOIN categories c ON pc.category_id = c.id
				WHERE pc.product_id = p.id AND c.slug = $1 AND c.deleted_at IS NULL
			)
		`)
		args = append(args, filters.Category)
	}

	// Removed PriceMin, PriceMax, Tags filters as they relate to removed fields

	// Combine conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS filtered_products"
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		r.logger.Error("failed to count products", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Add sorting and pagination (only sort by name or created_at for now)
	sortField := "p.created_at"
	sortOrder := "DESC"

	if filters.SortBy != "" {
		switch filters.SortBy {
		// Removed 'price' sort
		case "name":
			sortField = "p.title"
		}
	}

	if filters.SortOrder != "" {
		sortOrder = strings.ToUpper(filters.SortOrder)
		if sortOrder != "ASC" && sortOrder != "DESC" {
			sortOrder = "DESC"
		}
	}

	query := baseQuery + fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d",
		sortField, sortOrder, len(args)+1, len(args)+2)

	// Ensure PageSize and Page are valid
	if filters.PageSize <= 0 {
		filters.PageSize = 10 // Default page size
	}
	if filters.Page <= 0 {
		filters.Page = 1 // Default page number
	}
	args = append(args, filters.PageSize, (filters.Page-1)*filters.PageSize)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list products", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		var brandID, defaultVariantID sql.NullString
		var brand models.Brand
		var brandCreatedAt, brandUpdatedAt sql.NullTime

		// Scan only the selected fields
		if err := rows.Scan(
			&product.ID, &product.Title, &product.Slug, &product.Description, &product.ShortDescription,
			&product.Weight, &product.IsPublished, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
			&brandID, &defaultVariantID, &product.InventoryStatus,
			&brand.ID, &brand.Name, &brand.Slug, &brand.Description, &brandCreatedAt, &brandUpdatedAt, &brand.DeletedAt,
		); err != nil {
			r.logger.Error("failed to scan product row in ListProducts", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan product row: %w", err)
		}

		// Assign scanned nullable fields
		if brandID.Valid {
			product.BrandID = &brandID.String
			brand.CreatedAt = brandCreatedAt.Time
			brand.UpdatedAt = brandUpdatedAt.Time
			product.Brand = &brand
		}
		if defaultVariantID.Valid {
			product.DefaultVariantID = &defaultVariantID.String
		}

		// Note: Variants, Images, Categories are NOT fetched in ListProducts for performance.
		// They should be fetched individually when viewing a specific product.

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error scanning product rows", zap.Error(err))
		return nil, 0, fmt.Errorf("error scanning product rows: %w", err)
	}

	return products, total, nil
}

// CreateProduct creates a new product with all its associations in a transaction
func (r *ProductRepository) CreateProduct(ctx context.Context, product *models.Product) error {
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

	now := time.Now().UTC()
	product.CreatedAt = now
	product.UpdatedAt = now

	const productQuery = `
		INSERT INTO products (
			title, slug, description, short_description, price, discount_price,
			sku, inventory_qty, inventory_status, weight, is_published, brand_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`

	err = tx.QueryRowContext(ctx, productQuery,
		product.Title, product.Slug, product.Description, product.ShortDescription,
		product.Price, product.DiscountPrice, product.SKU, product.InventoryQty,
		product.InventoryStatus, product.Weight, product.IsPublished, product.BrandID, now, now,
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

	// Handle product images if any
	if len(product.Images) > 0 {
		const imageQuery = `
			INSERT INTO product_images (
				product_id, url, alt_text, position, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6)
		`

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

	// Handle categories if any
	if len(product.Categories) > 0 {
		const categoryQuery = `
			INSERT INTO product_categories (product_id, category_id)
			VALUES ($1, $2)
		`

		for _, category := range product.Categories {
			_, err = tx.ExecContext(ctx, categoryQuery, product.ID, category.ID)
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

// UpdateProduct updates a product and its associations
func (r *ProductRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
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

	now := time.Now().UTC() // Use UTC time
	product.UpdatedAt = now

	const productQuery = `
		UPDATE products SET
			title = $1, slug = $2, description = $3, short_description = $4,
			price = $5, discount_price = $6, sku = $7, inventory_qty = $8,
			inventory_status = $9, weight = $10, is_published = $11, brand_id = $12, updated_at = $13
		WHERE id = $14 AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(ctx, productQuery,
		product.Title, product.Slug, product.Description, product.ShortDescription,
		product.Price, product.DiscountPrice, product.SKU, product.InventoryQty,
		product.InventoryStatus, product.Weight, product.IsPublished, product.BrandID, now, product.ID,
	)
	if err != nil {
		r.logger.Error("failed to update product", zap.Error(err), zap.String("product_id", product.ID))
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrProductNotFound
	}

	// Update images (simplified approach - delete all and recreate)
	if _, err = tx.ExecContext(ctx, "DELETE FROM product_images WHERE product_id = $1", product.ID); err != nil {
		r.logger.Error("failed to delete existing product images", zap.Error(err))
		return fmt.Errorf("failed to delete existing product images: %w", err)
	}

	if len(product.Images) > 0 {
		const imageQuery = `
			INSERT INTO product_images (
				product_id, url, alt_text, position, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6)
		`

		for i := range product.Images {
			img := &product.Images[i]
			img.UpdatedAt = now
			if img.CreatedAt.IsZero() {
				img.CreatedAt = now
			}

			_, err = tx.ExecContext(ctx, imageQuery,
				product.ID, img.URL, img.AltText, img.Position, img.CreatedAt, now)
			if err != nil {
				r.logger.Error("failed to create product image", zap.Error(err))
				return fmt.Errorf("failed to create product image: %w", err)
			}
		}
	}

	// Update categories (simplified approach - delete all and recreate)
	if _, err = tx.ExecContext(ctx, "DELETE FROM product_categories WHERE product_id = $1", product.ID); err != nil {
		r.logger.Error("failed to delete existing product categories", zap.Error(err))
		return fmt.Errorf("failed to delete existing product categories: %w", err)
	}

	if len(product.Categories) > 0 {
		const categoryQuery = `
			INSERT INTO product_categories (product_id, category_id)
			VALUES ($1, $2)
		`

		for _, cat := range product.Categories {
			_, err = tx.ExecContext(ctx, categoryQuery, product.ID, cat.ID)
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

// DeleteProduct performs a soft delete of a product
func (r *ProductRepository) DeleteProduct(ctx context.Context, tx *sql.Tx, id string) error {
	const query = `
		UPDATE products
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now().UTC(), id) // Use UTC time
	if err != nil {
		r.logger.Error("failed to delete product", zap.Error(err), zap.String("product_id", id))
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

// GetProductVariants retrieves all variants for a product
func (r *ProductRepository) GetProductVariants(ctx context.Context, productID string) ([]*models.ProductVariant, error) {
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

		variants = append(variants, &variant)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating product variant rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating product variants: %w", err)
	}

	return variants, nil
}

// getVariantAttributes fetches attributes for a specific variant
func (r *ProductRepository) getVariantAttributes(ctx context.Context, variant *models.ProductVariant) error {
	const query = `
		SELECT a.name, pva.value
		FROM product_variant_attributes pva
		JOIN attributes a ON pva.attribute_id = a.id AND a.deleted_at IS NULL
		WHERE pva.product_variant_id = $1
		ORDER BY a.name
	`

	rows, err := r.db.QueryContext(ctx, query, variant.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var attr models.VariantAttributeValue
		if err := rows.Scan(&attr.Name, &attr.Value); err != nil {
			return err
		}
		variant.Attributes = append(variant.Attributes, attr)
	}

	return rows.Err()
}

// CreateVariant creates a new product variant with its attributes
func (r *ProductRepository) CreateVariant(ctx context.Context, tx *sql.Tx, productID string, variant *models.ProductVariant) error {
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
func (r *ProductRepository) createVariantAttributes(ctx context.Context, tx *sql.Tx, variant *models.ProductVariant) error {
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
func (r *ProductRepository) UpdateVariant(ctx context.Context, tx *sql.Tx, variant *models.ProductVariant) error {
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
func (r *ProductRepository) DeleteVariant(ctx context.Context, tx *sql.Tx, variantID string) error {
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
