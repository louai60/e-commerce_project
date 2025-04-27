package cache

import (
	"context"

	"github.com/louai60/e-commerce_project/backend/product-service/models"
)

// CacheInterface defines the interface for cache managers
type CacheInterface interface {
	// Product methods
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	SetProduct(ctx context.Context, product *models.Product) error
	GetProductList(ctx context.Context, filterKey string) ([]*models.Product, error)
	SetProductList(ctx context.Context, filterKey string, products []*models.Product) error
	InvalidateProduct(ctx context.Context, id string) error
	InvalidateProductLists(ctx context.Context) error
	GetProductVariants(ctx context.Context, productID string) ([]*models.ProductVariant, error)
	
	// Category methods
	GetCategory(ctx context.Context, id string) (*models.Category, error)
	SetCategory(ctx context.Context, category *models.Category) error
	GetCategoryList(ctx context.Context, filterKey string) ([]*models.Category, error)
	SetCategoryList(ctx context.Context, filterKey string, categories []*models.Category) error
	InvalidateCategory(ctx context.Context, id string) error
	InvalidateCategoryLists(ctx context.Context) error
	
	// Brand methods
	GetBrand(ctx context.Context, key string) (*models.Brand, error)
	SetBrand(ctx context.Context, key string, brand *models.Brand) error
	GetBrandList(ctx context.Context, filterKey string) ([]*models.Brand, error)
	SetBrandList(ctx context.Context, filterKey string, brands []*models.Brand) error
	InvalidateBrand(ctx context.Context, id string, slug string) error
	InvalidateBrandLists(ctx context.Context) error
	
	// General methods
	InvalidateProductAndRelated(ctx context.Context, productID string) error
	InvalidateByPattern(ctx context.Context, pattern string) error
	InvalidateProductsByCategory(ctx context.Context, categoryID string) error
	Close() error
	HealthCheck(ctx context.Context) error
	GetCacheStats(ctx context.Context) (map[string]interface{}, error)
	ClearExpiredKeys(ctx context.Context) error
}
