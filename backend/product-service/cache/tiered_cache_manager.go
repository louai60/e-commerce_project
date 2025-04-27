package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"github.com/louai60/e-commerce_project/backend/shared/cache"
	"go.uber.org/zap"
)

// TieredCacheManager implements a two-level cache with memory and Redis
type TieredCacheManager struct {
	tieredCache *cache.TieredCache
	logger      *zap.Logger
	keyMutexes  *sync.Map // For cache stampede protection
}

// TieredCacheOptions defines options for creating a tiered cache manager
type TieredCacheOptions struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisPoolSize int
	DefaultTTL    time.Duration
	Logger        *zap.Logger
	// Circuit breaker options
	FailureThreshold         int64
	ResetTimeout             time.Duration
	HalfOpenSuccessThreshold int64
}

// NewTieredCacheManager creates a new tiered cache manager
func NewTieredCacheManager(opts TieredCacheOptions) (*TieredCacheManager, error) {
	// Create Redis options
	redisOpts := &redis.Options{
		Addr:         opts.RedisAddr,
		Password:     opts.RedisPassword,
		DB:           opts.RedisDB,
		PoolSize:     opts.RedisPoolSize,
		MinIdleConns: 10,
		MaxConnAge:   time.Hour,
		IdleTimeout:  10 * time.Minute,
	}

	// Create tiered cache
	tieredCache, err := cache.NewTieredCache(cache.TieredCacheOptions{
		RedisOptions: redisOpts,
		DefaultTTL:   opts.DefaultTTL,
		// Pass circuit breaker options
		FailureThreshold:         opts.FailureThreshold,
		ResetTimeout:             opts.ResetTimeout,
		HalfOpenSuccessThreshold: opts.HalfOpenSuccessThreshold,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tiered cache: %w", err)
	}

	return &TieredCacheManager{
		tieredCache: tieredCache,
		logger:      opts.Logger,
		keyMutexes:  &sync.Map{},
	}, nil
}

// withTimeout adds a timeout to a context if one doesn't already exist
func (cm *TieredCacheManager) withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

// GetProduct retrieves a product from the cache
func (cm *TieredCacheManager) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	ctx, cancel := cm.withTimeout(ctx, DefaultTimeout)
	defer cancel()

	key := fmt.Sprintf("%s%s", ProductKeyPrefix, id)

	// Use mutex for cache stampede protection
	mutexInterface, _ := cm.keyMutexes.LoadOrStore(key, &sync.Mutex{})
	mutex := mutexInterface.(*sync.Mutex)

	// Try to get from cache first without locking
	var product models.Product
	err := cm.tieredCache.GetObject(ctx, key, "product", &product)
	if err == nil {
		// Get variants from cache
		variants, err := cm.GetProductVariants(ctx, id)
		if err == nil && variants != nil {
			// Convert []*models.ProductVariant to []models.ProductVariant
			productVariants := make([]models.ProductVariant, len(variants))
			for i, v := range variants {
				productVariants[i] = *v
			}
			product.Variants = productVariants
		}
		return &product, nil
	}

	// Cache miss, lock to prevent stampede
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		// Clean up mutex after use
		cm.keyMutexes.Delete(key)
	}()

	// Try again after acquiring lock (another goroutine might have populated the cache)
	err = cm.tieredCache.GetObject(ctx, key, "product", &product)
	if err == nil {
		// Get variants from cache
		variants, err := cm.GetProductVariants(ctx, id)
		if err == nil && variants != nil {
			// Convert []*models.ProductVariant to []models.ProductVariant
			productVariants := make([]models.ProductVariant, len(variants))
			for i, v := range variants {
				productVariants[i] = *v
			}
			product.Variants = productVariants
		}
		return &product, nil
	}

	return nil, ErrCacheKeyNotFound
}

// SetProduct stores a product in the cache
func (cm *TieredCacheManager) SetProduct(ctx context.Context, product *models.Product) error {
	key := fmt.Sprintf("%s%s", ProductKeyPrefix, product.ID)

	// Cache product without variants
	productCopy := *product
	productCopy.Variants = nil

	// Store in tiered cache
	if err := cm.tieredCache.SetObject(ctx, key, productCopy, "product"); err != nil {
		return err
	}

	// Cache variants separately
	if len(product.Variants) > 0 {
		variantsKey := fmt.Sprintf("product:variants:%s", product.ID)
		// Convert []models.ProductVariant to []*models.ProductVariant
		variantPtrs := make([]*models.ProductVariant, len(product.Variants))
		for i := range product.Variants {
			variantPtrs[i] = &product.Variants[i]
		}

		if err := cm.tieredCache.SetObject(ctx, variantsKey, variantPtrs, "product"); err != nil {
			return err
		}
	}

	return nil
}

// GetProductVariants retrieves product variants from the cache
func (cm *TieredCacheManager) GetProductVariants(ctx context.Context, productID string) ([]*models.ProductVariant, error) {
	key := fmt.Sprintf("product:variants:%s", productID)

	var variants []*models.ProductVariant
	err := cm.tieredCache.GetObject(ctx, key, "product", &variants)
	if err != nil {
		return nil, nil // Return nil instead of error for variants
	}

	return variants, nil
}

// GetProductList retrieves a list of products from the cache
func (cm *TieredCacheManager) GetProductList(ctx context.Context, filterKey string) ([]*models.Product, error) {
	key := fmt.Sprintf("%s%s", ProductListKeyPrefix, filterKey)

	var products []*models.Product
	err := cm.tieredCache.GetObject(ctx, key, "product_list", &products)
	if err != nil {
		return nil, err
	}

	return products, nil
}

// SetProductList stores a list of products in the cache
func (cm *TieredCacheManager) SetProductList(ctx context.Context, filterKey string, products []*models.Product) error {
	key := fmt.Sprintf("%s%s", ProductListKeyPrefix, filterKey)
	return cm.tieredCache.SetObject(ctx, key, products, "product_list")
}

// InvalidateProduct removes a product from the cache
func (cm *TieredCacheManager) InvalidateProduct(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", ProductKeyPrefix, id)
	variantsKey := fmt.Sprintf("product:variants:%s", id)

	// Delete both product and variants
	if err := cm.tieredCache.Delete(ctx, key); err != nil {
		return err
	}

	// Ignore errors for variants deletion
	_ = cm.tieredCache.Delete(ctx, variantsKey)

	return nil
}

// InvalidateProductLists removes all product lists from the cache
func (cm *TieredCacheManager) InvalidateProductLists(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", ProductListKeyPrefix)
	return cm.tieredCache.DeleteByPattern(ctx, pattern)
}

// Category-related methods
func (cm *TieredCacheManager) GetCategory(ctx context.Context, id string) (*models.Category, error) {
	key := fmt.Sprintf("%s%s", CategoryKeyPrefix, id)

	var category models.Category
	err := cm.tieredCache.GetObject(ctx, key, "category", &category)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (cm *TieredCacheManager) SetCategory(ctx context.Context, category *models.Category) error {
	key := fmt.Sprintf("%s%s", CategoryKeyPrefix, category.ID)
	return cm.tieredCache.SetObject(ctx, key, category, "category")
}

func (cm *TieredCacheManager) GetCategoryList(ctx context.Context, filterKey string) ([]*models.Category, error) {
	key := fmt.Sprintf("%s%s", CategoryListKeyPrefix, filterKey)

	var categories []*models.Category
	err := cm.tieredCache.GetObject(ctx, key, "category_list", &categories)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (cm *TieredCacheManager) SetCategoryList(ctx context.Context, filterKey string, categories []*models.Category) error {
	key := fmt.Sprintf("%s%s", CategoryListKeyPrefix, filterKey)
	return cm.tieredCache.SetObject(ctx, key, categories, "category_list")
}

func (cm *TieredCacheManager) InvalidateCategory(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", CategoryKeyPrefix, id)
	return cm.tieredCache.Delete(ctx, key)
}

func (cm *TieredCacheManager) InvalidateCategoryLists(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", CategoryListKeyPrefix)
	return cm.tieredCache.DeleteByPattern(ctx, pattern)
}

// Brand-related methods
func (cm *TieredCacheManager) GetBrand(ctx context.Context, key string) (*models.Brand, error) {
	var brand models.Brand
	err := cm.tieredCache.GetObject(ctx, key, "brand", &brand)
	if err != nil {
		return nil, err
	}

	return &brand, nil
}

func (cm *TieredCacheManager) SetBrand(ctx context.Context, key string, brand *models.Brand) error {
	return cm.tieredCache.SetObject(ctx, key, brand, "brand")
}

func (cm *TieredCacheManager) GetBrandList(ctx context.Context, filterKey string) ([]*models.Brand, error) {
	key := fmt.Sprintf("%s%s", BrandListKeyPrefix, filterKey)

	var brands []*models.Brand
	err := cm.tieredCache.GetObject(ctx, key, "brand_list", &brands)
	if err != nil {
		return nil, err
	}

	return brands, nil
}

func (cm *TieredCacheManager) SetBrandList(ctx context.Context, filterKey string, brands []*models.Brand) error {
	key := fmt.Sprintf("%s%s", BrandListKeyPrefix, filterKey)
	return cm.tieredCache.SetObject(ctx, key, brands, "brand_list")
}

func (cm *TieredCacheManager) InvalidateBrand(ctx context.Context, id string, slug string) error {
	// Invalidate both ID and Slug keys
	keysToDelete := []string{
		fmt.Sprintf(BrandCacheKeyByID, id),
	}

	if slug != "" {
		keysToDelete = append(keysToDelete, fmt.Sprintf(BrandCacheKeyBySlug, slug))
	}

	for _, key := range keysToDelete {
		if err := cm.tieredCache.Delete(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func (cm *TieredCacheManager) InvalidateBrandLists(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", BrandListKeyPrefix)
	return cm.tieredCache.DeleteByPattern(ctx, pattern)
}

// InvalidateProductAndRelated invalidates a product and all related caches
func (cm *TieredCacheManager) InvalidateProductAndRelated(ctx context.Context, productID string) error {
	// Invalidate product
	if err := cm.InvalidateProduct(ctx, productID); err != nil {
		return fmt.Errorf("failed to invalidate product: %w", err)
	}

	// Invalidate product lists
	if err := cm.InvalidateProductLists(ctx); err != nil {
		return fmt.Errorf("failed to invalidate product lists: %w", err)
	}

	// Invalidate related category lists (since product count might change)
	if err := cm.InvalidateCategoryLists(ctx); err != nil {
		return fmt.Errorf("failed to invalidate category lists: %w", err)
	}

	return nil
}

// InvalidateByPattern invalidates all keys matching a pattern
func (cm *TieredCacheManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	return cm.tieredCache.DeleteByPattern(ctx, pattern)
}

// InvalidateProductsByCategory invalidates all product caches related to a category
func (cm *TieredCacheManager) InvalidateProductsByCategory(ctx context.Context, categoryID string) error {
	// Invalidate category-specific product lists
	pattern := fmt.Sprintf("%s*category:%s*", ProductListKeyPrefix, categoryID)
	if err := cm.InvalidateByPattern(ctx, pattern); err != nil {
		return fmt.Errorf("failed to invalidate category products: %w", err)
	}

	// Also invalidate general product lists as they might include products from this category
	return cm.InvalidateProductLists(ctx)
}

// Close closes the cache manager
func (cm *TieredCacheManager) Close() error {
	return cm.tieredCache.Close()
}

// HealthCheck checks if the cache is healthy
func (cm *TieredCacheManager) HealthCheck(ctx context.Context) error {
	return cm.tieredCache.HealthCheck(ctx)
}

// GetCacheStats returns statistics about the cache
func (cm *TieredCacheManager) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	return cm.tieredCache.GetMemoryCacheStats(), nil
}

// ClearExpiredKeys clears expired keys from the cache
func (cm *TieredCacheManager) ClearExpiredKeys(ctx context.Context) error {
	// Clear memory cache
	cm.tieredCache.ClearMemoryCache()
	return nil
}

// WarmupResult contains the results of a cache warm-up operation
type WarmupResult struct {
	// SuccessCount is the number of successfully warmed up keys
	SuccessCount int
	// ErrorCount is the number of keys that failed to warm up
	ErrorCount int
	// Duration is the total duration of the warm-up operation
	Duration time.Duration
}

// WarmupCache warms up the cache with critical data
func (cm *TieredCacheManager) WarmupCache(ctx context.Context) (*WarmupResult, error) {
	cm.logger.Info("Starting cache warm-up for critical data")
	startTime := time.Now()
	result := &WarmupResult{}

	// Get top categories for warm-up
	categories, err := cm.warmupCategories(ctx)
	if err != nil {
		cm.logger.Error("Failed to warm up categories", zap.Error(err))
		result.ErrorCount++
	} else {
		result.SuccessCount += len(categories)
	}

	// Get top brands for warm-up
	brands, err := cm.warmupBrands(ctx)
	if err != nil {
		cm.logger.Error("Failed to warm up brands", zap.Error(err))
		result.ErrorCount++
	} else {
		result.SuccessCount += len(brands)
	}

	// Get featured products for warm-up
	products, err := cm.warmupProducts(ctx)
	if err != nil {
		cm.logger.Error("Failed to warm up products", zap.Error(err))
		result.ErrorCount++
	} else {
		result.SuccessCount += len(products)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// warmupCategories warms up category cache
func (cm *TieredCacheManager) warmupCategories(ctx context.Context) ([]*models.Category, error) {
	// This would typically call the repository to get categories
	// For now, we'll just return an empty slice as a placeholder
	return []*models.Category{}, nil
}

// warmupBrands warms up brand cache
func (cm *TieredCacheManager) warmupBrands(ctx context.Context) ([]*models.Brand, error) {
	// This would typically call the repository to get brands
	// For now, we'll just return an empty slice as a placeholder
	return []*models.Brand{}, nil
}

// warmupProducts warms up product cache
func (cm *TieredCacheManager) warmupProducts(ctx context.Context) ([]*models.Product, error) {
	// This would typically call the repository to get products
	// For now, we'll just return an empty slice as a placeholder
	return []*models.Product{}, nil
}
