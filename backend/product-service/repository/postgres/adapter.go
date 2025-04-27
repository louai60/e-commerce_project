package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/louai60/e-commerce_project/backend/product-service/db"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
	"go.uber.org/zap"
)

// ProductRepositoryAdapter adapts the ProductRepository to the repository.ProductRepository interface
type ProductRepositoryAdapter struct {
	repo   *ProductRepository // Using the original ProductRepository for now
	logger *zap.Logger
}

// NewProductRepositoryAdapter creates a new adapter for the ProductRepository
func NewProductRepositoryAdapter(dbConfig *db.DBConfig, logger *zap.Logger) repository.ProductRepository {
	return &ProductRepositoryAdapter{
		repo:   &ProductRepository{db: dbConfig.Master, logger: logger},
		logger: logger,
	}
}

// Ensure ProductRepositoryAdapter implements repository.ProductRepository at compile time
// This line will cause a compilation error if the adapter doesn't implement all methods
var _ repository.ProductRepository = (*ProductRepositoryAdapter)(nil)

// BeginTx starts a new transaction
func (a *ProductRepositoryAdapter) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return a.repo.db.BeginTx(ctx, nil)
}

// CreateProduct creates a new product
func (a *ProductRepositoryAdapter) CreateProduct(ctx context.Context, product *models.Product) error {
	return a.repo.CreateProduct(ctx, product)
}

// GetByID retrieves a product by ID
func (a *ProductRepositoryAdapter) GetByID(ctx context.Context, id string) (*models.Product, error) {
	return a.repo.GetProduct(ctx, id)
}

// GetBySlug retrieves a product by slug
func (a *ProductRepositoryAdapter) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	// Implement using the new repository
	// This is a simplified implementation - you may need to adjust based on your actual repository methods
	const query = `
		SELECT
			p.id, p.title, p.slug, p.description, p.short_description,
			p.weight, p.is_published, p.created_at, p.updated_at, p.deleted_at,
			p.brand_id, p.inventory_status
		FROM products p
		WHERE p.slug = $1 AND p.deleted_at IS NULL
	`

	product := &models.Product{}
	var brandID sql.NullString

	err := a.repo.db.QueryRowContext(ctx, query, slug).Scan(
		&product.ID, &product.Title, &product.Slug, &product.Description, &product.ShortDescription,
		&product.Weight, &product.IsPublished, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
		&brandID, &product.InventoryStatus,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrProductNotFound
		}
		a.logger.Error("failed to get product by slug", zap.Error(err), zap.String("slug", slug))
		return nil, err
	}

	// Assign scanned nullable fields
	if brandID.Valid {
		product.BrandID = &brandID.String
	}

	// Get associated data
	if err := a.repo.getProductImages(ctx, product); err != nil {
		a.logger.Error("failed to get product images", zap.Error(err), zap.String("product_id", product.ID))
	}

	if err := a.repo.getProductCategories(ctx, product); err != nil {
		a.logger.Error("failed to get product categories", zap.Error(err), zap.String("product_id", product.ID))
	}

	if err := a.repo.getProductVariantsAndAttributes(ctx, product); err != nil {
		a.logger.Error("failed to get product variants", zap.Error(err), zap.String("product_id", product.ID))
	}

	return product, nil
}

// List retrieves a paginated list of products
func (a *ProductRepositoryAdapter) List(ctx context.Context, offset, limit int) ([]*models.Product, int, error) {
	// Convert to the new filters format
	filters := models.ProductFilters{
		Page:     offset/limit + 1,
		PageSize: limit,
	}

	products, total, err := a.repo.ListProducts(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	return products, int(total), nil
}

// UpdateProduct updates an existing product
func (a *ProductRepositoryAdapter) UpdateProduct(ctx context.Context, product *models.Product) error {
	return a.repo.UpdateProduct(ctx, product)
}

// DeleteProduct soft-deletes a product
func (a *ProductRepositoryAdapter) DeleteProduct(ctx context.Context, id string) error {
	// The repository method requires a transaction parameter, but the interface doesn't
	// So we'll handle the transaction here
	tx, err := a.repo.db.BeginTx(ctx, nil)
	if err != nil {
		a.logger.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				a.logger.Error("failed to rollback transaction", zap.Error(rbErr))
			}
		}
	}()

	// Call the repository method with the transaction
	err = a.repo.DeleteProduct(ctx, tx, id)
	if err != nil {
		return err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		a.logger.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Implement the remaining methods from the ProductRepository interface
// These are stubs that you'll need to implement based on your actual repository methods

// AddImage adds an image to a product
func (a *ProductRepositoryAdapter) AddImage(ctx context.Context, image *models.ProductImage) error {
	// Implement this method
	return nil
}

// RemoveImage removes an image from a product
func (a *ProductRepositoryAdapter) RemoveImage(ctx context.Context, imageID string) error {
	// Implement this method
	return nil
}

// GetProductImages gets all images for a product
func (a *ProductRepositoryAdapter) GetProductImages(ctx context.Context, productID string) ([]models.ProductImage, error) {
	product := &models.Product{ID: productID}
	if err := a.repo.getProductImages(ctx, product); err != nil {
		return nil, err
	}
	return product.Images, nil
}

// CreateVariant creates a new product variant
func (a *ProductRepositoryAdapter) CreateVariant(ctx context.Context, tx *sql.Tx, productID string, variant *models.ProductVariant) error {
	// Implement this method
	return nil
}

// UpdateVariant updates an existing product variant
func (a *ProductRepositoryAdapter) UpdateVariant(ctx context.Context, tx *sql.Tx, variant *models.ProductVariant) error {
	// Implement this method
	return nil
}

// DeleteVariant deletes a product variant
func (a *ProductRepositoryAdapter) DeleteVariant(ctx context.Context, tx *sql.Tx, variantID string) error {
	// Implement this method
	return nil
}

// GetVariantByID gets a product variant by ID
func (a *ProductRepositoryAdapter) GetVariantByID(ctx context.Context, variantID string) (*models.ProductVariant, error) {
	// Implement this method
	return nil, nil
}

// GetVariantsBySKU gets product variants by SKU
func (a *ProductRepositoryAdapter) GetVariantsBySKU(ctx context.Context, sku string) ([]models.ProductVariant, error) {
	// Implement this method
	return nil, nil
}

// GetVariantsByProductID gets all variants for a product
func (a *ProductRepositoryAdapter) GetVariantsByProductID(ctx context.Context, productID string) ([]models.ProductVariant, error) {
	const query = `
		SELECT
			pv.id, pv.product_id, pv.sku, pv.title, pv.price, pv.discount_price,
			pv.inventory_qty, pv.created_at, pv.updated_at, pv.deleted_at
		FROM product_variants pv
		WHERE pv.product_id = $1 AND pv.deleted_at IS NULL
		ORDER BY pv.created_at
	`

	rows, err := a.repo.db.QueryContext(ctx, query, productID)
	if err != nil {
		a.logger.Error("failed to query product variants", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to query product variants: %w", err)
	}
	defer rows.Close()

	var variants []models.ProductVariant
	for rows.Next() {
		var variant models.ProductVariant
		if err := rows.Scan(
			&variant.ID, &variant.ProductID, &variant.SKU, &variant.Title, &variant.Price, &variant.DiscountPrice,
			&variant.InventoryQty, &variant.CreatedAt, &variant.UpdatedAt, &variant.DeletedAt,
		); err != nil {
			a.logger.Error("failed to scan product variant", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product variant: %w", err)
		}

		// Get attributes for this variant
		attributes, err := a.GetVariantAttributes(ctx, variant.ID)
		if err != nil {
			a.logger.Error("failed to get variant attributes", zap.Error(err), zap.String("variant_id", variant.ID))
			return nil, fmt.Errorf("failed to get variant attributes: %w", err)
		}
		variant.Attributes = attributes

		variants = append(variants, variant)
	}

	if err := rows.Err(); err != nil {
		a.logger.Error("error iterating product variant rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating product variants: %w", err)
	}

	return variants, nil
}

// GetProductVariants gets all variants for a product (alias for GetVariantsByProductID)
func (a *ProductRepositoryAdapter) GetProductVariants(ctx context.Context, productID string) ([]*models.ProductVariant, error) {
	// This is an alias for GetVariantsByProductID but with pointer slice return type
	variants, err := a.GetVariantsByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Convert []models.ProductVariant to []*models.ProductVariant
	result := make([]*models.ProductVariant, len(variants))
	for i := range variants {
		result[i] = &variants[i]
	}

	return result, nil
}

// AddVariantAttribute adds an attribute to a variant
func (a *ProductRepositoryAdapter) AddVariantAttribute(ctx context.Context, tx *sql.Tx, variantID, attributeID, value string) error {
	// Implement this method
	return nil
}

// RemoveVariantAttribute removes an attribute from a variant
func (a *ProductRepositoryAdapter) RemoveVariantAttribute(ctx context.Context, tx *sql.Tx, variantID, attributeID string) error {
	// Implement this method
	return nil
}

// GetVariantAttributes gets all attributes for a variant
func (a *ProductRepositoryAdapter) GetVariantAttributes(ctx context.Context, variantID string) ([]models.VariantAttributeValue, error) {
	const query = `
		SELECT a.name, pva.value
		FROM product_variant_attributes pva
		JOIN attributes a ON pva.attribute_id = a.id AND a.deleted_at IS NULL
		WHERE pva.product_variant_id = $1
		ORDER BY a.name
	`

	rows, err := a.repo.db.QueryContext(ctx, query, variantID)
	if err != nil {
		a.logger.Error("failed to get variant attributes", zap.Error(err), zap.String("variant_id", variantID))
		return nil, fmt.Errorf("failed to get variant attributes: %w", err)
	}
	defer rows.Close()

	var attributes []models.VariantAttributeValue
	for rows.Next() {
		var attr models.VariantAttributeValue
		if err := rows.Scan(&attr.Name, &attr.Value); err != nil {
			a.logger.Error("failed to scan variant attribute", zap.Error(err))
			return nil, fmt.Errorf("failed to scan variant attribute: %w", err)
		}
		attributes = append(attributes, attr)
	}

	if err := rows.Err(); err != nil {
		a.logger.Error("error iterating variant attributes", zap.Error(err))
		return nil, fmt.Errorf("error iterating variant attributes: %w", err)
	}

	return attributes, nil
}

// AddVariantImage adds an image to a variant
func (a *ProductRepositoryAdapter) AddVariantImage(ctx context.Context, image *models.VariantImage) error {
	// Implement this method
	return nil
}

// RemoveVariantImage removes an image from a variant
func (a *ProductRepositoryAdapter) RemoveVariantImage(ctx context.Context, variantID, imageID string) error {
	// Implement this method
	return nil
}

// DeleteVariantImage deletes a variant image
func (a *ProductRepositoryAdapter) DeleteVariantImage(ctx context.Context, id string) error {
	// Implement this method
	return nil
}

// UpdateVariantImage updates a variant image
func (a *ProductRepositoryAdapter) UpdateVariantImage(ctx context.Context, image *models.VariantImage) error {
	// Implement this method
	return nil
}

// GetVariantImages gets all images for a variant
func (a *ProductRepositoryAdapter) GetVariantImages(ctx context.Context, variantID string) ([]models.VariantImage, error) {
	// Implement this method
	return nil, nil
}

// GetProductAttributes gets all attributes for a product
func (a *ProductRepositoryAdapter) GetProductAttributes(ctx context.Context, productID string) ([]models.ProductAttribute, error) {
	query := `
		SELECT id, product_id, name, value, created_at, updated_at
		FROM product_attributes
		WHERE product_id = $1
		ORDER BY name
	`

	rows, err := a.repo.db.QueryContext(ctx, query, productID)
	if err != nil {
		a.logger.Error("failed to get product attributes", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product attributes: %w", err)
	}
	defer rows.Close()

	var attributes []models.ProductAttribute
	for rows.Next() {
		var attr models.ProductAttribute
		if err := rows.Scan(
			&attr.ID, &attr.ProductID, &attr.Name, &attr.Value,
			&attr.CreatedAt, &attr.UpdatedAt,
		); err != nil {
			a.logger.Error("failed to scan product attribute", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product attribute: %w", err)
		}
		attributes = append(attributes, attr)
	}

	if err := rows.Err(); err != nil {
		a.logger.Error("error iterating product attributes", zap.Error(err))
		return nil, fmt.Errorf("error iterating product attributes: %w", err)
	}

	return attributes, nil
}

// AddProductAttribute adds an attribute to a product
func (a *ProductRepositoryAdapter) AddProductAttribute(ctx context.Context, attribute *models.ProductAttribute) error {
	now := time.Now().UTC()
	attribute.CreatedAt = now
	attribute.UpdatedAt = now

	query := `
		INSERT INTO product_attributes (product_id, name, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := a.repo.db.QueryRowContext(ctx, query,
		attribute.ProductID, attribute.Name, attribute.Value, now, now,
	).Scan(&attribute.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return fmt.Errorf("attribute already exists for this product")
		}
		a.logger.Error("failed to add product attribute", zap.Error(err))
		return fmt.Errorf("failed to add product attribute: %w", err)
	}

	return nil
}

// UpdateProductAttribute updates an existing product attribute
func (a *ProductRepositoryAdapter) UpdateProductAttribute(ctx context.Context, attribute *models.ProductAttribute) error {
	// Implement this method
	return nil
}

// RemoveProductAttribute removes an attribute from a product
func (a *ProductRepositoryAdapter) RemoveProductAttribute(ctx context.Context, attributeID string) error {
	// Implement this method
	return nil
}

// GetProductSpecifications gets all specifications for a product
func (a *ProductRepositoryAdapter) GetProductSpecifications(ctx context.Context, productID string) ([]models.ProductSpecification, error) {
	query := `
		SELECT id, product_id, name, value, unit, created_at, updated_at
		FROM product_specifications
		WHERE product_id = $1
		ORDER BY name
	`

	rows, err := a.repo.db.QueryContext(ctx, query, productID)
	if err != nil {
		a.logger.Error("failed to get product specifications", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product specifications: %w", err)
	}
	defer rows.Close()

	var specs []models.ProductSpecification
	for rows.Next() {
		var spec models.ProductSpecification
		if err := rows.Scan(
			&spec.ID, &spec.ProductID, &spec.Name, &spec.Value, &spec.Unit,
			&spec.CreatedAt, &spec.UpdatedAt,
		); err != nil {
			a.logger.Error("failed to scan product specification", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product specification: %w", err)
		}
		specs = append(specs, spec)
	}

	if err := rows.Err(); err != nil {
		a.logger.Error("error iterating product specifications", zap.Error(err))
		return nil, fmt.Errorf("error iterating product specifications: %w", err)
	}

	return specs, nil
}

// AddProductSpecification adds a specification to a product
func (a *ProductRepositoryAdapter) AddProductSpecification(ctx context.Context, spec *models.ProductSpecification) error {
	now := time.Now().UTC()
	spec.CreatedAt = now
	spec.UpdatedAt = now

	query := `
		INSERT INTO product_specifications (product_id, name, value, unit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := a.repo.db.QueryRowContext(ctx, query, spec.ProductID, spec.Name, spec.Value, spec.Unit, now, now).Scan(&spec.ID)
	if err != nil {
		a.logger.Error("failed to add product specification", zap.Error(err))
		return fmt.Errorf("failed to add product specification: %w", err)
	}

	return nil
}

// UpdateProductSpecification updates an existing product specification
func (a *ProductRepositoryAdapter) UpdateProductSpecification(ctx context.Context, spec *models.ProductSpecification) error {
	// Implement this method
	return nil
}

// RemoveProductSpecification removes a specification from a product
func (a *ProductRepositoryAdapter) RemoveProductSpecification(ctx context.Context, specID string) error {
	// Implement this method
	return nil
}

// GetProductSEO gets the SEO data for a product
func (a *ProductRepositoryAdapter) GetProductSEO(ctx context.Context, productID string) (*models.ProductSEO, error) {
	query := `
		SELECT id, product_id, meta_title, meta_description, keywords, tags, created_at, updated_at
		FROM product_seo
		WHERE product_id = $1
	`

	var seo models.ProductSEO
	err := a.repo.db.QueryRowContext(ctx, query, productID).Scan(
		&seo.ID, &seo.ProductID, &seo.MetaTitle, &seo.MetaDescription,
		pq.Array(&seo.Keywords), pq.Array(&seo.Tags), &seo.CreatedAt, &seo.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No SEO data found
		}
		a.logger.Error("failed to get product SEO", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product SEO: %w", err)
	}

	return &seo, nil
}

// UpsertProductSEO creates or updates the SEO data for a product
func (a *ProductRepositoryAdapter) UpsertProductSEO(ctx context.Context, seo *models.ProductSEO) error {
	now := time.Now().UTC()
	seo.UpdatedAt = now

	// Check if SEO data already exists for this product
	existingSEO, err := a.GetProductSEO(ctx, seo.ProductID)
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

		err := a.repo.db.QueryRowContext(ctx, query,
			seo.ProductID, seo.MetaTitle, seo.MetaDescription,
			pq.Array(seo.Keywords), pq.Array(seo.Tags), now, now,
		).Scan(&seo.ID)
		if err != nil {
			a.logger.Error("failed to insert product SEO", zap.Error(err))
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

		_, err := a.repo.db.ExecContext(ctx, query,
			seo.MetaTitle, seo.MetaDescription,
			pq.Array(seo.Keywords), pq.Array(seo.Tags), now, seo.ID,
		)
		if err != nil {
			a.logger.Error("failed to update product SEO", zap.Error(err))
			return fmt.Errorf("failed to update product SEO: %w", err)
		}
	}

	return nil
}

// GetProductShipping gets the shipping data for a product
func (a *ProductRepositoryAdapter) GetProductShipping(ctx context.Context, productID string) (*models.ProductShipping, error) {
	query := `
		SELECT id, product_id, free_shipping, estimated_days, express_available, created_at, updated_at
		FROM product_shipping
		WHERE product_id = $1
	`

	var shipping models.ProductShipping
	err := a.repo.db.QueryRowContext(ctx, query, productID).Scan(
		&shipping.ID, &shipping.ProductID, &shipping.FreeShipping,
		&shipping.EstimatedDays, &shipping.ExpressAvailable,
		&shipping.CreatedAt, &shipping.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No shipping data found
		}
		a.logger.Error("failed to get product shipping", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product shipping: %w", err)
	}

	return &shipping, nil
}

// UpsertProductShipping creates or updates the shipping data for a product
func (a *ProductRepositoryAdapter) UpsertProductShipping(ctx context.Context, shipping *models.ProductShipping) error {
	now := time.Now().UTC()
	shipping.UpdatedAt = now

	// Check if shipping data already exists for this product
	existingShipping, err := a.GetProductShipping(ctx, shipping.ProductID)
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

		err := a.repo.db.QueryRowContext(ctx, query,
			shipping.ProductID, shipping.FreeShipping, shipping.EstimatedDays,
			shipping.ExpressAvailable, now, now,
		).Scan(&shipping.ID)
		if err != nil {
			a.logger.Error("failed to insert product shipping", zap.Error(err))
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

		_, err := a.repo.db.ExecContext(ctx, query,
			shipping.FreeShipping, shipping.EstimatedDays,
			shipping.ExpressAvailable, now, shipping.ID,
		)
		if err != nil {
			a.logger.Error("failed to update product shipping", zap.Error(err))
			return fmt.Errorf("failed to update product shipping: %w", err)
		}
	}

	return nil
}

// GetProductDiscounts gets all discounts for a product
func (a *ProductRepositoryAdapter) GetProductDiscounts(ctx context.Context, productID string) ([]models.ProductDiscount, error) {
	query := `
		SELECT id, product_id, discount_type, value, expires_at, created_at, updated_at
		FROM product_discounts
		WHERE product_id = $1
		ORDER BY created_at DESC
	`

	rows, err := a.repo.db.QueryContext(ctx, query, productID)
	if err != nil {
		a.logger.Error("failed to get product discounts", zap.Error(err), zap.String("product_id", productID))
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
			a.logger.Error("failed to scan product discount", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product discount: %w", err)
		}
		discounts = append(discounts, discount)
	}

	if err := rows.Err(); err != nil {
		a.logger.Error("error iterating product discounts", zap.Error(err))
		return nil, fmt.Errorf("error iterating product discounts: %w", err)
	}

	return discounts, nil
}

// AddProductDiscount adds a discount to a product
func (a *ProductRepositoryAdapter) AddProductDiscount(ctx context.Context, discount *models.ProductDiscount) error {
	now := time.Now().UTC()
	discount.CreatedAt = now
	discount.UpdatedAt = now

	query := `
		INSERT INTO product_discounts (product_id, discount_type, value, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := a.repo.db.QueryRowContext(ctx, query,
		discount.ProductID, discount.Type, discount.Value,
		discount.ExpiresAt, now, now,
	).Scan(&discount.ID)
	if err != nil {
		a.logger.Error("failed to add product discount", zap.Error(err))
		return fmt.Errorf("failed to add product discount: %w", err)
	}

	return nil
}

// UpdateProductDiscount updates an existing product discount
func (a *ProductRepositoryAdapter) UpdateProductDiscount(ctx context.Context, discount *models.ProductDiscount) error {
	// Implement this method
	return nil
}

// RemoveProductDiscount removes a discount from a product
func (a *ProductRepositoryAdapter) RemoveProductDiscount(ctx context.Context, discountID string) error {
	// Implement this method
	return nil
}

// GetInventoryLocations gets all inventory locations for a product
func (a *ProductRepositoryAdapter) GetInventoryLocations(ctx context.Context, productID string) ([]models.InventoryLocation, error) {
	query := `
		SELECT id, product_id, warehouse_id, available_qty, created_at, updated_at
		FROM product_inventory_locations
		WHERE product_id = $1
		ORDER BY warehouse_id
	`

	rows, err := a.repo.db.QueryContext(ctx, query, productID)
	if err != nil {
		a.logger.Error("failed to get inventory locations", zap.Error(err), zap.String("product_id", productID))
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
			a.logger.Error("failed to scan inventory location", zap.Error(err))
			return nil, fmt.Errorf("failed to scan inventory location: %w", err)
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		a.logger.Error("error iterating inventory locations", zap.Error(err))
		return nil, fmt.Errorf("error iterating inventory locations: %w", err)
	}

	return locations, nil
}

// UpsertInventoryLocation creates or updates an inventory location for a product
func (a *ProductRepositoryAdapter) UpsertInventoryLocation(ctx context.Context, location *models.InventoryLocation) error {
	now := time.Now().UTC()
	location.UpdatedAt = now

	// Check if location already exists
	query := `
		SELECT id, created_at FROM product_inventory_locations
		WHERE product_id = $1 AND warehouse_id = $2
	`
	var existingID string
	var createdAt time.Time
	err := a.repo.db.QueryRowContext(ctx, query, location.ProductID, location.WarehouseID).Scan(&existingID, &createdAt)

	if err != nil && err != sql.ErrNoRows {
		a.logger.Error("failed to check existing inventory location", zap.Error(err))
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

		err := a.repo.db.QueryRowContext(ctx, query,
			location.ProductID, location.WarehouseID, location.AvailableQty, now, now,
		).Scan(&location.ID)
		if err != nil {
			a.logger.Error("failed to insert inventory location", zap.Error(err))
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

		_, err := a.repo.db.ExecContext(ctx, query,
			location.AvailableQty, now, location.ID,
		)
		if err != nil {
			a.logger.Error("failed to update inventory location", zap.Error(err))
			return fmt.Errorf("failed to update inventory location: %w", err)
		}
	}

	return nil
}

// RemoveInventoryLocation removes an inventory location from a product
func (a *ProductRepositoryAdapter) RemoveInventoryLocation(ctx context.Context, productID, warehouseID string) error {
	// Implement this method
	return nil
}

// GetProductTags gets all tags for a product
func (a *ProductRepositoryAdapter) GetProductTags(ctx context.Context, productID string) ([]models.ProductTag, error) {
	query := `
		SELECT id, product_id, tag, created_at, updated_at
		FROM product_tags
		WHERE product_id = $1
		ORDER BY tag
	`

	rows, err := a.repo.db.QueryContext(ctx, query, productID)
	if err != nil {
		a.logger.Error("failed to get product tags", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product tags: %w", err)
	}
	defer rows.Close()

	var tags []models.ProductTag
	for rows.Next() {
		var tag models.ProductTag
		if err := rows.Scan(
			&tag.ID, &tag.ProductID, &tag.Tag, &tag.CreatedAt, &tag.UpdatedAt,
		); err != nil {
			a.logger.Error("failed to scan product tag", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		a.logger.Error("error iterating product tags", zap.Error(err))
		return nil, fmt.Errorf("error iterating product tags: %w", err)
	}

	return tags, nil
}

// AddProductTag adds a tag to a product
func (a *ProductRepositoryAdapter) AddProductTag(ctx context.Context, tag *models.ProductTag) error {
	// Implement this method using the underlying repository
	now := time.Now().UTC()
	tag.CreatedAt = now
	tag.UpdatedAt = now

	query := `
		INSERT INTO product_tags (product_id, tag, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := a.repo.db.QueryRowContext(ctx, query, tag.ProductID, tag.Tag, now, now).Scan(&tag.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return fmt.Errorf("tag already exists for this product")
		}
		a.logger.Error("failed to add product tag", zap.Error(err))
		return fmt.Errorf("failed to add product tag: %w", err)
	}

	return nil
}

// RemoveProductTag removes a tag from a product
func (a *ProductRepositoryAdapter) RemoveProductTag(ctx context.Context, productID, tag string) error {
	// Implement this method
	query := `
		DELETE FROM product_tags
		WHERE product_id = $1 AND tag = $2
	`

	result, err := a.repo.db.ExecContext(ctx, query, productID, tag)
	if err != nil {
		a.logger.Error("failed to remove product tag", zap.Error(err))
		return fmt.Errorf("failed to remove product tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		a.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag not found for this product")
	}

	return nil
}
