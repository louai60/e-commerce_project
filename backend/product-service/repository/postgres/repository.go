package postgres

import (
	"context"
	"database/sql"

	"github.com/louai60/e-commerce_project/backend/product-service/models"
)

// ProductRepository defines the interface for product repository operations
type ProductRepositoryInterface interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	GetProductFixed(ctx context.Context, id string) (*models.Product, error)
	FixProductData(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*models.Product, error)
	GetBySlug(ctx context.Context, slug string) (*models.Product, error)
	ListProducts(ctx context.Context, filters models.ProductFilters) ([]*models.Product, int64, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
	
	// Variant operations
	CreateVariant(ctx context.Context, tx *sql.Tx, productID string, variant *models.ProductVariant) error
	UpdateVariant(ctx context.Context, tx *sql.Tx, variant *models.ProductVariant) error
	DeleteVariant(ctx context.Context, tx *sql.Tx, id string) error
	GetProductVariants(ctx context.Context, productID string) ([]*models.ProductVariant, error)
	GetVariantByID(ctx context.Context, id string) (*models.ProductVariant, error)
	GetVariantImages(ctx context.Context, variantID string) ([]models.VariantImage, error)
	GetVariantAttributes(ctx context.Context, variantID string) ([]models.VariantAttributeValue, error)
	
	// Image operations
	AddProductImage(ctx context.Context, image *models.ProductImage) error
	UpdateProductImage(ctx context.Context, image *models.ProductImage) error
	DeleteProductImage(ctx context.Context, id string) error
	GetProductImages(ctx context.Context, productID string) ([]models.ProductImage, error)
	
	// Category operations
	AddProductCategory(ctx context.Context, productID, categoryID string) error
	RemoveProductCategory(ctx context.Context, productID, categoryID string) error
	GetProductCategories(ctx context.Context, productID string) ([]models.Category, error)
	
	// Tag operations
	AddProductTag(ctx context.Context, tag *models.ProductTag) error
	RemoveProductTag(ctx context.Context, id string) error
	GetProductTags(ctx context.Context, productID string) ([]models.ProductTag, error)
	
	// Specification operations
	AddProductSpecification(ctx context.Context, spec *models.ProductSpecification) error
	UpdateProductSpecification(ctx context.Context, spec *models.ProductSpecification) error
	DeleteProductSpecification(ctx context.Context, id string) error
	GetProductSpecifications(ctx context.Context, productID string) ([]models.ProductSpecification, error)
	
	// SEO operations
	UpsertProductSEO(ctx context.Context, seo *models.ProductSEO) error
	GetProductSEO(ctx context.Context, productID string) (*models.ProductSEO, error)
	
	// Shipping operations
	UpsertProductShipping(ctx context.Context, shipping *models.ProductShipping) error
	GetProductShipping(ctx context.Context, productID string) (*models.ProductShipping, error)
	
	// Discount operations
	AddProductDiscount(ctx context.Context, discount *models.ProductDiscount) error
	UpdateProductDiscount(ctx context.Context, discount *models.ProductDiscount) error
	DeleteProductDiscount(ctx context.Context, id string) error
	GetProductDiscounts(ctx context.Context, productID string) ([]models.ProductDiscount, error)
	
	// Inventory operations
	UpsertInventoryLocation(ctx context.Context, location *models.InventoryLocation) error
	DeleteInventoryLocation(ctx context.Context, id string) error
	GetInventoryLocations(ctx context.Context, productID string) ([]models.InventoryLocation, error)
	
	// Attribute operations
	AddProductAttribute(ctx context.Context, attr *models.ProductAttribute) error
	UpdateProductAttribute(ctx context.Context, attr *models.ProductAttribute) error
	DeleteProductAttribute(ctx context.Context, id string) error
	GetProductAttributes(ctx context.Context, productID string) ([]models.ProductAttribute, error)
}
