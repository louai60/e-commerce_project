package repository

import (
	"context"
	"database/sql"

	"github.com/louai60/e-commerce_project/backend/product-service/models"
)

type ProductRepository interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id string) (*models.Product, error)
	GetBySlug(ctx context.Context, slug string) (*models.Product, error)
	List(ctx context.Context, offset, limit int) ([]*models.Product, int, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error

	// Variant-specific methods
	GetProductVariants(ctx context.Context, productID string) ([]*models.ProductVariant, error)
	CreateVariant(ctx context.Context, tx *sql.Tx, productID string, variant *models.ProductVariant) error
	UpdateVariant(ctx context.Context, tx *sql.Tx, variant *models.ProductVariant) error
	DeleteVariant(ctx context.Context, tx *sql.Tx, variantID string) error

	// Variant attribute methods
	GetVariantAttributes(ctx context.Context, variantID string) ([]models.VariantAttributeValue, error)

	// Variant image methods
	AddVariantImage(ctx context.Context, image *models.VariantImage) error
	GetVariantImages(ctx context.Context, variantID string) ([]models.VariantImage, error)
	UpdateVariantImage(ctx context.Context, image *models.VariantImage) error
	DeleteVariantImage(ctx context.Context, id string) error

	// Product image methods
	GetProductImages(ctx context.Context, productID string) ([]models.ProductImage, error)

	// Tag-related methods
	GetProductTags(ctx context.Context, productID string) ([]models.ProductTag, error)
	AddProductTag(ctx context.Context, tag *models.ProductTag) error
	RemoveProductTag(ctx context.Context, productID, tag string) error

	// Attribute-related methods
	GetProductAttributes(ctx context.Context, productID string) ([]models.ProductAttribute, error)
	AddProductAttribute(ctx context.Context, attribute *models.ProductAttribute) error
	UpdateProductAttribute(ctx context.Context, attribute *models.ProductAttribute) error
	RemoveProductAttribute(ctx context.Context, attributeID string) error

	// Specification-related methods
	GetProductSpecifications(ctx context.Context, productID string) ([]models.ProductSpecification, error)
	AddProductSpecification(ctx context.Context, spec *models.ProductSpecification) error
	UpdateProductSpecification(ctx context.Context, spec *models.ProductSpecification) error
	RemoveProductSpecification(ctx context.Context, specID string) error

	// SEO-related methods
	GetProductSEO(ctx context.Context, productID string) (*models.ProductSEO, error)
	UpsertProductSEO(ctx context.Context, seo *models.ProductSEO) error

	// Shipping-related methods
	GetProductShipping(ctx context.Context, productID string) (*models.ProductShipping, error)
	UpsertProductShipping(ctx context.Context, shipping *models.ProductShipping) error

	// Discount-related methods
	GetProductDiscounts(ctx context.Context, productID string) ([]models.ProductDiscount, error)
	AddProductDiscount(ctx context.Context, discount *models.ProductDiscount) error
	UpdateProductDiscount(ctx context.Context, discount *models.ProductDiscount) error
	RemoveProductDiscount(ctx context.Context, discountID string) error

	// Inventory-related methods
	GetInventoryLocations(ctx context.Context, productID string) ([]models.InventoryLocation, error)
	UpsertInventoryLocation(ctx context.Context, location *models.InventoryLocation) error
	RemoveInventoryLocation(ctx context.Context, productID, warehouseID string) error

	// SKU-related methods
	IsSKUExists(ctx context.Context, sku string) (bool, error)
}

type BrandRepository interface {
	CreateBrand(ctx context.Context, brand *models.Brand) error
	GetBrandByID(ctx context.Context, id string) (*models.Brand, error)
	GetBrandBySlug(ctx context.Context, slug string) (*models.Brand, error)
	ListBrands(ctx context.Context, offset, limit int) ([]*models.Brand, int, error)
}

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *models.Category) error
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
	ListCategories(ctx context.Context, offset, limit int) ([]*models.Category, int, error)
}
