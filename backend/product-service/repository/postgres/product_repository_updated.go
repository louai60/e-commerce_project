// repository/postgres/product_repository.go
// This is an updated version of the product repository that uses the repository base
// with read replica support
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	// "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/lib/pq"
	"github.com/louai60/e-commerce_project/backend/product-service/db"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
)

// ProductRepositoryV2 implements repository.ProductRepository with read replica support
type ProductRepositoryV2 struct {
	*RepositoryBase
}

// NewProductRepositoryV2 creates a new ProductRepositoryV2
func NewProductRepositoryV2(dbConfig *db.DBConfig, logger *zap.Logger) *ProductRepositoryV2 {
	return &ProductRepositoryV2{
		RepositoryBase: NewRepositoryBase(dbConfig, logger),
	}
}

// GetProduct retrieves a product by ID with its core details, brand, categories, images, and variants.
func (r *ProductRepositoryV2) GetProduct(ctx context.Context, id string) (*models.Product, error) {
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

	// Use ExecuteQueryRow for read operations (will use replica if available)
	err := r.ExecuteQueryRow(ctx, query, id).Scan(
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
func (r *ProductRepositoryV2) getProductVariantsAndAttributes(ctx context.Context, product *models.Product) error {
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

	// Use ExecuteQuery for read operations (will use replica if available)
	rows, err := r.ExecuteQuery(ctx, query, product.ID)
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
func (r *ProductRepositoryV2) getProductImages(ctx context.Context, product *models.Product) error {
	const query = `
		SELECT id, product_id, url, alt_text, position, created_at, updated_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY position ASC
	`

	// Use ExecuteQuery for read operations (will use replica if available)
	rows, err := r.ExecuteQuery(ctx, query, product.ID)
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
func (r *ProductRepositoryV2) getProductCategories(ctx context.Context, product *models.Product) error {
	const query = `
		SELECT c.id, c.name, c.slug, c.description, c.parent_id, c.created_at, c.updated_at, c.deleted_at
		FROM categories c
		JOIN product_categories pc ON c.id = pc.category_id
		WHERE pc.product_id = $1 AND c.deleted_at IS NULL
	` // Added c.deleted_at to SELECT and WHERE clause

	// Use ExecuteQuery for read operations (will use replica if available)
	rows, err := r.ExecuteQuery(ctx, query, product.ID)
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

// ListProducts retrieves paginated products with optional filters
func (r *ProductRepositoryV2) ListProducts(ctx context.Context, filters models.ProductFilters) ([]*models.Product, int64, error) {
	// Build base query - Select only core product fields
	baseQuery := `
		SELECT
			p.id, p.title, p.slug, p.description, p.short_description,
			p.weight, p.is_published, p.created_at, p.updated_at, p.deleted_at,
			p.brand_id, p.inventory_status,
			b.id, b.name, b.slug, b.description, b.created_at, b.updated_at, b.deleted_at
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id AND b.deleted_at IS NULL
		WHERE p.deleted_at IS NULL
	`

	// Add filtering conditions
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

	// Combine conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS filtered_products"
	var total int64

	// Use ExecuteQueryRow for read operations (will use replica if available)
	if err := r.ExecuteQueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		r.logger.Error("failed to count products", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Add sorting and pagination
	sortField := "p.created_at"
	sortOrder := "DESC"

	if filters.SortBy != "" {
		switch filters.SortBy {
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

	// Use ExecuteQuery for read operations (will use replica if available)
	rows, err := r.ExecuteQuery(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list products", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		var brandID sql.NullString
		var brand models.Brand
		var brandCreatedAt, brandUpdatedAt sql.NullTime

		// Scan only the selected fields
		var brandIDStr, brandNameStr, brandSlugStr, brandDescStr sql.NullString
		if err := rows.Scan(
			&product.ID, &product.Title, &product.Slug, &product.Description, &product.ShortDescription,
			&product.Weight, &product.IsPublished, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
			&brandID, &product.InventoryStatus,
			&brandIDStr, &brandNameStr, &brandSlugStr, &brandDescStr, &brandCreatedAt, &brandUpdatedAt, &brand.DeletedAt,
		); err != nil {
			r.logger.Error("failed to scan product row in ListProducts", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan product row: %w", err)
		}

		// Assign scanned nullable fields
		if brandID.Valid {
			product.BrandID = &brandID.String

			// Only set brand if we have a valid brand ID
			if brandIDStr.Valid {
				brand.ID = brandIDStr.String
				if brandNameStr.Valid {
					brand.Name = brandNameStr.String
				}
				if brandSlugStr.Valid {
					brand.Slug = brandSlugStr.String
				}
				if brandDescStr.Valid {
					brand.Description = brandDescStr.String
				}
				if brandCreatedAt.Valid {
					brand.CreatedAt = brandCreatedAt.Time
				}
				if brandUpdatedAt.Valid {
					brand.UpdatedAt = brandUpdatedAt.Time
				}
				product.Brand = &brand
			}
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error scanning product rows", zap.Error(err))
		return nil, 0, fmt.Errorf("error scanning product rows: %w", err)
	}

	return products, total, nil
}

// CreateProduct creates a new product with all its associations in a transaction
func (r *ProductRepositoryV2) CreateProduct(ctx context.Context, product *models.Product) error {
	// Use BeginTx from RepositoryBase
	tx, err := r.BeginTx(ctx)
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
func (r *ProductRepositoryV2) UpdateProduct(ctx context.Context, product *models.Product) error {
	// Use BeginTx from RepositoryBase
	tx, err := r.BeginTx(ctx)
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

			_, err = tx.ExecContext(ctx, imageQuery,
				product.ID, img.URL, img.AltText, img.Position, now, now)
			if err != nil {
				r.logger.Error("failed to create product image", zap.Error(err))
				return fmt.Errorf("failed to create product image: %w", err)
			}
		}
	}

	// Update categories (delete all and recreate)
	if _, err = tx.ExecContext(ctx, "DELETE FROM product_categories WHERE product_id = $1", product.ID); err != nil {
		r.logger.Error("failed to delete existing product categories", zap.Error(err))
		return fmt.Errorf("failed to delete existing product categories: %w", err)
	}

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

// DeleteProduct performs a soft delete of a product
func (r *ProductRepositoryV2) DeleteProduct(ctx context.Context, id string) error {
	const query = `
		UPDATE products
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	// Use ExecuteExec for write operations (will use master)
	result, err := r.ExecuteExec(ctx, query, time.Now().UTC(), id)
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
