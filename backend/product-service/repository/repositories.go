package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"go.uber.org/zap"
)

type PostgresProductRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// Ensure PostgresProductRepository implements ProductRepository
var _ ProductRepository = (*PostgresProductRepository)(nil)

type PostgresBrandRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// Ensure PostgresBrandRepository implements BrandRepository
var _ BrandRepository = (*PostgresBrandRepository)(nil)

type PostgresCategoryRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// Ensure PostgresCategoryRepository implements CategoryRepository
var _ CategoryRepository = (*PostgresCategoryRepository)(nil)

// PostgresProductRepository methods
func (r *PostgresProductRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

func (r *PostgresProductRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Insert product
	query := `
        INSERT INTO products (
            id, title, slug, description, short_description, inventory_status,
            weight, is_published, brand_id, created_at, updated_at,
            price, discount_price, inventory_qty, sku
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
        RETURNING id`

	var discountPrice *float64 = nil
	if product.DiscountPrice != nil {
		discountPrice = &product.DiscountPrice.Amount
	}

	err = tx.QueryRowContext(ctx, query,
		product.ID, product.Title, product.Slug, product.Description,
		product.ShortDescription, product.InventoryStatus, product.Weight,
		product.IsPublished, product.BrandID,
		product.CreatedAt, product.UpdatedAt, product.Price.Amount, discountPrice,
		product.InventoryQty, product.SKU,
	).Scan(&product.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				if pqErr.Constraint == "products_slug_key" {
					return models.ErrProductSlugExists
				}
			}
		}
		return fmt.Errorf("failed to create product: %w", err)
	}

	// // Create default variant with price
	// if product.DefaultVariantID != nil && *product.DefaultVariantID != "" {
	// 	variantQuery := `
	//         INSERT INTO product_variants (
	//             id, product_id, sku, price, discount_price,
	//             inventory_qty, created_at, updated_at
	//         ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	// 	var variantDiscountPrice *float64 = nil
	// 	if product.DiscountPrice != nil {
	// 		variantDiscountPrice = &product.DiscountPrice.Amount
	// 	}

	// 	_, err = tx.ExecContext(ctx, variantQuery,
	// 		product.DefaultVariantID, product.ID, product.SKU,
	// 		product.Price.Amount, variantDiscountPrice,
	// 		product.InventoryQty, product.CreatedAt, product.UpdatedAt,
	// 	)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to create default variant: %w", err)
	// 	}
	// }

	// If there are variants, create them
	if len(product.Variants) > 0 {
		// Create each variant
		for i := range product.Variants {
			variant := &product.Variants[i]
			variant.ProductID = product.ID

			variantQuery := `
			INSERT INTO product_variants (
				product_id, sku, title, price, discount_price, inventory_qty, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id`

			err = tx.QueryRowContext(
				ctx, variantQuery,
				variant.ProductID, variant.SKU, variant.Title, variant.Price, variant.DiscountPrice,
				variant.InventoryQty, now, now,
			).Scan(&variant.ID)

			if err != nil {
				r.logger.Error("failed to create product variant", zap.Error(err))
				return fmt.Errorf("failed to create product variant: %w", err)
			}

			// No need to set default variant anymore

			// Create variant attributes if any
			if len(variant.Attributes) > 0 {
				for _, attr := range variant.Attributes {
					// First, get or create the attribute
					var attrID string
					attrQuery := `
					SELECT id FROM attributes WHERE name = $1 AND deleted_at IS NULL`

					err = tx.QueryRowContext(ctx, attrQuery, attr.Name).Scan(&attrID)
					if err == sql.ErrNoRows {
						// Create the attribute
						attrInsertQuery := `
						INSERT INTO attributes (name, created_at, updated_at)
						VALUES ($1, $2, $3)
						RETURNING id`

						err = tx.QueryRowContext(ctx, attrInsertQuery, attr.Name, now, now).Scan(&attrID)
						if err != nil {
							r.logger.Error("failed to create attribute", zap.Error(err))
							return fmt.Errorf("failed to create attribute: %w", err)
						}
					} else if err != nil {
						r.logger.Error("failed to query attribute", zap.Error(err))
						return fmt.Errorf("failed to query attribute: %w", err)
					}

					// Now create the variant attribute
					varAttrQuery := `
					INSERT INTO product_variant_attributes (product_variant_id, attribute_id, value, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5)`

					_, err = tx.ExecContext(ctx, varAttrQuery, variant.ID, attrID, attr.Value, now, now)
					if err != nil {
						r.logger.Error("failed to create variant attribute", zap.Error(err))
						return fmt.Errorf("failed to create variant attribute: %w", err)
					}
				}
			}

			// Handle variant images
			if len(variant.Images) > 0 {
				// First delete existing images
				deleteImagesQuery := `DELETE FROM variant_images WHERE variant_id = $1`
				_, err = tx.ExecContext(ctx, deleteImagesQuery, variant.ID)
				if err != nil {
					r.logger.Error("failed to delete variant images", zap.Error(err))
					return fmt.Errorf("failed to delete variant images: %w", err)
				}

				// Then create new images
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
		}
	}

	// Handle images if any
	if len(product.Images) > 0 {
		imageQuery := `
			INSERT INTO product_images (
				product_id, url, alt_text, position, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id`

		for i := range product.Images {
			img := &product.Images[i]
			img.ProductID = product.ID
			img.CreatedAt = now
			img.UpdatedAt = now

			var imageID string
			err = tx.QueryRowContext(
				ctx, imageQuery,
				img.ProductID, img.URL, img.AltText, img.Position, now, now,
			).Scan(&imageID)

			if err != nil {
				r.logger.Error("failed to create product image", zap.Error(err))
				return fmt.Errorf("failed to create product image: %w", err)
			}

			img.ID = imageID
		}
	}

	return tx.Commit()
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	query := `
		SELECT
			p.id, p.title, p.slug, p.description, p.short_description,
			p.price, p.discount_price, p.sku, p.inventory_qty,
			p.inventory_status, p.weight, p.is_published, p.brand_id,
			p.created_at, p.updated_at,
			b.id, b.name, b.slug, b.description, b.created_at, b.updated_at,
			pv.id, pv.product_id, pv.title, pv.sku, pv.price, pv.discount_price,
			pv.inventory_qty, pv.created_at, pv.updated_at
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		LEFT JOIN product_variants pv ON p.id = pv.product_id
		WHERE p.id = $1
	`

	var product models.Product
	var brand models.Brand
	var variant models.ProductVariant
	var priceAmount float64
	var discountPriceAmount *float64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Title, &product.Slug, &product.Description,
		&product.ShortDescription, &priceAmount, &discountPriceAmount, &product.SKU,
		&product.InventoryQty, &product.InventoryStatus, &product.Weight,
		&product.IsPublished, &product.BrandID, &product.CreatedAt, &product.UpdatedAt,
		&brand.ID, &brand.Name, &brand.Slug,
		&brand.Description, &brand.CreatedAt, &brand.UpdatedAt,
		&variant.ID, &variant.ProductID, &variant.Title, &variant.SKU, &variant.Price,
		&variant.DiscountPrice, &variant.InventoryQty, &variant.CreatedAt, &variant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Set the price fields
	product.Price = models.Price{
		Amount:   priceAmount,
		Currency: "USD", // Default currency
	}
	if discountPriceAmount != nil {
		product.DiscountPrice = &models.Price{
			Amount:   *discountPriceAmount,
			Currency: "USD", // Default currency
		}
	}

	// Set the brand if it exists
	if brand.ID != "" {
		product.Brand = &brand
	}

	// Set the variant if it exists
	if variant.ID != "" {
		product.Variants = []models.ProductVariant{variant}
	}

	return &product, nil
}

func (r *PostgresProductRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	product := &models.Product{}
	query := `
        SELECT id, title, slug, description, short_description, inventory_status,
               weight, is_published, brand_id, created_at, updated_at
        FROM products
        WHERE slug = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&product.ID, &product.Title, &product.Slug, &product.Description,
		&product.ShortDescription, &product.InventoryStatus, &product.Weight,
		&product.IsPublished, &product.BrandID,
		&product.CreatedAt, &product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product not found")
	}
	if err != nil {
		r.logger.Error("failed to get product", zap.Error(err))
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Get images for this product
	imagesQuery := `
		SELECT id, product_id, url, alt_text, position, created_at, updated_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY position ASC`

	imagesRows, err := r.db.QueryContext(ctx, imagesQuery, product.ID)
	if err != nil {
		r.logger.Error("failed to get product images", zap.Error(err))
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}
	defer imagesRows.Close()

	var images []models.ProductImage
	for imagesRows.Next() {
		var img models.ProductImage
		err := imagesRows.Scan(
			&img.ID, &img.ProductID, &img.URL, &img.AltText, &img.Position, &img.CreatedAt, &img.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan product image", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product image: %w", err)
		}
		images = append(images, img)
	}

	product.Images = images

	// Get variants for this product
	variants, err := r.GetProductVariants(ctx, product.ID)
	if err != nil {
		r.logger.Error("failed to get product variants", zap.Error(err))
		return nil, fmt.Errorf("failed to get product variants: %w", err)
	}

	// Convert []*models.ProductVariant to []models.ProductVariant
	productVariants := make([]models.ProductVariant, len(variants))
	for i, v := range variants {
		productVariants[i] = *v
	}
	product.Variants = productVariants

	// Use the first variant's data for backward compatibility
	if len(variants) > 0 {
		defaultVariant := variants[0] // Use first variant

		product.Price = models.Price{
			Amount:   defaultVariant.Price,
			Currency: "USD", // Default currency
		}
		if defaultVariant.DiscountPrice != nil {
			discountPrice := models.Price{
				Amount:   *defaultVariant.DiscountPrice,
				Currency: "USD", // Default currency
			}
			product.DiscountPrice = &discountPrice
		}
		product.SKU = defaultVariant.SKU
		product.InventoryQty = defaultVariant.InventoryQty
	}

	return product, nil
}

func (r *PostgresProductRepository) List(ctx context.Context, offset, limit int) ([]*models.Product, int, error) {
	var total int
	// Remove deleted_at check initially to get total count
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE deleted_at IS NULL").Scan(&total)
	if err != nil {
		r.logger.Error("failed to count products", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	query := `
        SELECT id, title, slug, description, short_description, inventory_status,
               weight, is_published, brand_id, created_at, updated_at
        FROM products
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
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
		err := rows.Scan(
			&product.ID, &product.Title, &product.Slug, &product.Description,
			&product.ShortDescription, &product.InventoryStatus, &product.Weight,
			&product.IsPublished, &product.BrandID,
			&product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan product", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		// Get images for this product
		imagesQuery := `
			SELECT id, product_id, url, alt_text, position, created_at, updated_at
			FROM product_images
			WHERE product_id = $1
			ORDER BY position ASC`

		imagesRows, err := r.db.QueryContext(ctx, imagesQuery, product.ID)
		if err != nil {
			r.logger.Error("failed to get product images", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to get product images: %w", err)
		}

		var images []models.ProductImage
		for imagesRows.Next() {
			var img models.ProductImage
			err := imagesRows.Scan(
				&img.ID, &img.ProductID, &img.URL, &img.AltText, &img.Position, &img.CreatedAt, &img.UpdatedAt,
			)
			if err != nil {
				imagesRows.Close()
				r.logger.Error("failed to scan product image", zap.Error(err))
				return nil, 0, fmt.Errorf("failed to scan product image: %w", err)
			}
			images = append(images, img)
		}
		imagesRows.Close()

		product.Images = images

		// Get variants for this product
		variants, err := r.GetProductVariants(ctx, product.ID)
		if err != nil {
			r.logger.Error("failed to get product variants", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to get product variants: %w", err)
		}

		// Convert []*models.ProductVariant to []models.ProductVariant
		productVariants := make([]models.ProductVariant, len(variants))
		for i, v := range variants {
			productVariants[i] = *v
		}
		product.Variants = productVariants

		// Use the first variant's data for backward compatibility
		if len(variants) > 0 {
			defaultVariant := variants[0] // Use first variant

			product.Price = models.Price{
				Amount:   defaultVariant.Price,
				Currency: "USD", // Default currency
			}
			if defaultVariant.DiscountPrice != nil {
				discountPrice := models.Price{
					Amount:   *defaultVariant.DiscountPrice,
					Currency: "USD", // Default currency
				}
				product.DiscountPrice = &discountPrice
			}
			product.SKU = defaultVariant.SKU
			product.InventoryQty = defaultVariant.InventoryQty
		}

		products = append(products, product)
	}

	return products, total, rows.Err()
}

func (r *PostgresProductRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
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

	// Update the product
	query := `
        UPDATE products
        SET title = $1, slug = $2, description = $3, short_description = $4, weight = $5,
            is_published = $6, brand_id = $7, updated_at = $8
        WHERE id = $9 AND deleted_at IS NULL`

	result, err := tx.ExecContext(
		ctx, query,
		product.Title, product.Slug, product.Description, product.ShortDescription,
		product.Weight, product.IsPublished, product.BrandID, now, product.ID,
	)
	if err != nil {
		r.logger.Error("failed to update product", zap.String("id", product.ID), zap.Error(err))
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected after product update", zap.String("id", product.ID), zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("product not found or not updated") // Or return models.ErrProductNotFound
	}

	// Update variants if any
	if len(product.Variants) > 0 {
		for i := range product.Variants {
			variant := &product.Variants[i]

			// Check if variant exists
			var variantExists bool
			var variantID string

			if variant.ID != "" {
				// Update existing variant
				variantQuery := `
				UPDATE product_variants
				SET sku = $1, title = $2, price = $3, discount_price = $4, inventory_qty = $5, updated_at = $6
				WHERE id = $7 AND product_id = $8 AND deleted_at IS NULL`

				result, err = tx.ExecContext(
					ctx, variantQuery,
					variant.SKU, variant.Title, variant.Price, variant.DiscountPrice,
					variant.InventoryQty, now, variant.ID, product.ID,
				)
				if err != nil {
					r.logger.Error("failed to update product variant", zap.Error(err))
					return fmt.Errorf("failed to update product variant: %w", err)
				}

				rowsAffected, err = result.RowsAffected()
				if err != nil {
					r.logger.Error("failed to get rows affected after variant update", zap.Error(err))
					return fmt.Errorf("failed to get rows affected: %w", err)
				}
				if rowsAffected == 0 {
					// Variant doesn't exist, create it
					variantExists = false
				} else {
					variantExists = true
					variantID = variant.ID
				}
			} else {
				// New variant
				variantExists = false
			}

			if !variantExists {
				// Create new variant
				variant.ProductID = product.ID

				variantQuery := `
				INSERT INTO product_variants (
					product_id, sku, title, price, discount_price, inventory_qty, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id`

				err = tx.QueryRowContext(
					ctx, variantQuery,
					variant.ProductID, variant.SKU, variant.Title, variant.Price, variant.DiscountPrice,
					variant.InventoryQty, now, now,
				).Scan(&variant.ID)

				if err != nil {
					r.logger.Error("failed to create product variant", zap.Error(err))
					return fmt.Errorf("failed to create product variant: %w", err)
				}

				variantID = variant.ID
			}

			// No need to set default variant anymore

			// Update variant attributes if any
			if len(variant.Attributes) > 0 {
				// First delete existing attributes
				deleteAttrQuery := `DELETE FROM product_variant_attributes WHERE product_variant_id = $1`
				_, err = tx.ExecContext(ctx, deleteAttrQuery, variantID)
				if err != nil {
					r.logger.Error("failed to delete variant attributes", zap.Error(err))
					return fmt.Errorf("failed to delete variant attributes: %w", err)
				}

				// Then create new attributes
				for _, attr := range variant.Attributes {
					// First, get or create the attribute
					var attrID string
					attrQuery := `
					SELECT id FROM attributes WHERE name = $1 AND deleted_at IS NULL`

					err = tx.QueryRowContext(ctx, attrQuery, attr.Name).Scan(&attrID)
					if err == sql.ErrNoRows {
						// Create the attribute
						attrInsertQuery := `
						INSERT INTO attributes (name, created_at, updated_at)
						VALUES ($1, $2, $3)
						RETURNING id`

						err = tx.QueryRowContext(ctx, attrInsertQuery, attr.Name, now, now).Scan(&attrID)
						if err != nil {
							r.logger.Error("failed to create attribute", zap.Error(err))
							return fmt.Errorf("failed to create attribute: %w", err)
						}
					} else if err != nil {
						r.logger.Error("failed to query attribute", zap.Error(err))
						return fmt.Errorf("failed to query attribute: %w", err)
					}

					// Now create the variant attribute
					varAttrQuery := `
					INSERT INTO product_variant_attributes (product_variant_id, attribute_id, value, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5)`

					_, err = tx.ExecContext(ctx, varAttrQuery, variantID, attrID, attr.Value, now, now)
					if err != nil {
						r.logger.Error("failed to create variant attribute", zap.Error(err))
						return fmt.Errorf("failed to create variant attribute: %w", err)
					}
				}
			}

			// Handle variant images
			if len(variant.Images) > 0 {
				// First delete existing images
				deleteImagesQuery := `DELETE FROM variant_images WHERE variant_id = $1`
				_, err = tx.ExecContext(ctx, deleteImagesQuery, variantID)
				if err != nil {
					r.logger.Error("failed to delete variant images", zap.Error(err))
					return fmt.Errorf("failed to delete variant images: %w", err)
				}

				// Then create new images
				imageQuery := `
					INSERT INTO variant_images (
						variant_id, url, alt_text, position, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6)
					RETURNING id`

				for i := range variant.Images {
					img := &variant.Images[i]
					img.VariantID = variantID
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
		}
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresProductRepository) DeleteProduct(ctx context.Context, id string) error {
	query := `UPDATE products SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		r.logger.Error("failed to delete product", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected after product delete", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("product not found or already deleted") // Or return models.ErrProductNotFound
	}

	return nil
}

// PostgresBrandRepository methods
func (r *PostgresBrandRepository) CreateBrand(ctx context.Context, brand *models.Brand) error {
	now := time.Now()
	brand.CreatedAt = now
	brand.UpdatedAt = now

	query := `
        INSERT INTO brands (
            name, slug, description, created_at, updated_at, deleted_at
        ) VALUES ($1, $2, $3, $4, $5, NULL)
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

func (r *PostgresBrandRepository) GetBrandByID(ctx context.Context, id string) (*models.Brand, error) {
	brand := &models.Brand{}
	query := `
        SELECT id, name, slug, description, created_at, updated_at, deleted_at
        FROM brands
        WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&brand.ID, &brand.Name, &brand.Slug, &brand.Description,
		&brand.CreatedAt, &brand.UpdatedAt, &brand.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("brand not found")
	}
	if err != nil {
		r.logger.Error("failed to get brand", zap.Error(err))
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}
	return brand, nil
}

func (r *PostgresBrandRepository) GetBrandBySlug(ctx context.Context, slug string) (*models.Brand, error) {
	brand := &models.Brand{}
	query := `
        SELECT id, name, slug, description, created_at, updated_at, deleted_at
        FROM brands
        WHERE slug = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&brand.ID, &brand.Name, &brand.Slug, &brand.Description,
		&brand.CreatedAt, &brand.UpdatedAt, &brand.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("brand not found")
	}
	if err != nil {
		r.logger.Error("failed to get brand", zap.Error(err))
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}
	return brand, nil
}

func (r *PostgresBrandRepository) ListBrands(ctx context.Context, offset, limit int) ([]*models.Brand, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM brands WHERE deleted_at IS NULL").Scan(&total)
	if err != nil {
		r.logger.Error("failed to count brands", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count brands: %w", err)
	}

	query := `
        SELECT id, name, slug, description, created_at, updated_at, deleted_at
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
			&brand.CreatedAt, &brand.UpdatedAt, &brand.DeletedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan brand", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan brand: %w", err)
		}
		brands = append(brands, brand)
	}

	return brands, total, rows.Err()
}

// PostgresCategoryRepository methods
func (r *PostgresCategoryRepository) CreateCategory(ctx context.Context, category *models.Category) error {
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

func (r *PostgresCategoryRepository) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	category := &models.Category{}
	query := `
        SELECT c.id, c.name, c.slug, c.description, c.parent_id, c.created_at, c.updated_at, c.deleted_at,
               p.name as parent_name
        FROM categories c
        LEFT JOIN categories p ON c.parent_id = p.id
        WHERE c.id = $1 AND c.deleted_at IS NULL`

	var parentName sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Slug, &category.Description,
		&category.ParentID, &category.CreatedAt, &category.UpdatedAt, &category.DeletedAt,
		&parentName,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category not found")
	}
	if err != nil {
		r.logger.Error("failed to get category", zap.Error(err))
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// Set parent name if available
	if parentName.Valid {
		category.ParentName = parentName.String
	}

	return category, nil
}

func (r *PostgresCategoryRepository) GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	category := &models.Category{}
	query := `
        SELECT c.id, c.name, c.slug, c.description, c.parent_id, c.created_at, c.updated_at, c.deleted_at,
               p.name as parent_name
        FROM categories c
        LEFT JOIN categories p ON c.parent_id = p.id
        WHERE c.slug = $1 AND c.deleted_at IS NULL`

	var parentName sql.NullString
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&category.ID, &category.Name, &category.Slug, &category.Description,
		&category.ParentID, &category.CreatedAt, &category.UpdatedAt, &category.DeletedAt,
		&parentName,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category not found")
	}
	if err != nil {
		r.logger.Error("failed to get category", zap.Error(err))
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// Set parent name if available
	if parentName.Valid {
		category.ParentName = parentName.String
	}

	return category, nil
}

func (r *PostgresCategoryRepository) ListCategories(ctx context.Context, offset, limit int) ([]*models.Category, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories WHERE deleted_at IS NULL").Scan(&total)
	if err != nil {
		r.logger.Error("failed to count categories", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
	}

	// Modified query to join with parent category to get parent_name
	query := `
        SELECT c.id, c.name, c.slug, c.description, c.parent_id, c.created_at, c.updated_at, c.deleted_at,
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
			&category.ID, &category.Name, &category.Slug, &category.Description,
			&category.ParentID, &category.CreatedAt, &category.UpdatedAt, &category.DeletedAt,
			&parentName,
		)
		if err != nil {
			r.logger.Error("failed to scan category", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan category: %w", err)
		}

		// Set parent name if available
		if parentName.Valid {
			category.ParentName = parentName.String
		}

		categories = append(categories, category)
	}

	return categories, total, rows.Err()
}

func NewProductRepository(db *sql.DB, logger *zap.Logger) ProductRepository {
	if db == nil {
		logger.Fatal("database connection cannot be nil")
		return nil
	}
	return &PostgresProductRepository{
		db:     db,
		logger: logger.Named("ProductRepository"),
	}
}

func NewBrandRepository(db *sql.DB, logger *zap.Logger) BrandRepository {
	if db == nil {
		logger.Fatal("database connection cannot be nil")
		return nil
	}
	return &PostgresBrandRepository{
		db:     db,
		logger: logger.Named("BrandRepository"),
	}
}

func NewCategoryRepository(db *sql.DB, logger *zap.Logger) CategoryRepository {
	if db == nil {
		logger.Fatal("database connection cannot be nil")
		return nil
	}
	return &PostgresCategoryRepository{
		db:     db,
		logger: logger.Named("CategoryRepository"),
	}
}

// GetProductVariants retrieves all variants for a product
func (r *PostgresProductRepository) GetProductVariants(ctx context.Context, productID string) ([]*models.ProductVariant, error) {
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

// GetVariantAttributes fetches attributes for a specific variant by ID
func (r *PostgresProductRepository) GetVariantAttributes(ctx context.Context, variantID string) ([]models.VariantAttributeValue, error) {
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
func (r *PostgresProductRepository) getVariantAttributes(ctx context.Context, variant *models.ProductVariant) error {
	attributes, err := r.GetVariantAttributes(ctx, variant.ID)
	if err != nil {
		return err
	}
	variant.Attributes = attributes
	return nil
}

// CreateVariant creates a new product variant with its attributes
func (r *PostgresProductRepository) CreateVariant(ctx context.Context, tx *sql.Tx, productID string, variant *models.ProductVariant) error {
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
func (r *PostgresProductRepository) createVariantAttributes(ctx context.Context, tx *sql.Tx, variant *models.ProductVariant) error {
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
func (r *PostgresProductRepository) UpdateVariant(ctx context.Context, tx *sql.Tx, variant *models.ProductVariant) error {
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
func (r *PostgresProductRepository) DeleteVariant(ctx context.Context, tx *sql.Tx, variantID string) error {
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
