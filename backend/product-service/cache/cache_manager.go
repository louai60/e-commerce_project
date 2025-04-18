package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"go.uber.org/zap"
)

// Error constants
var (
	ErrCacheKeyNotFound = fmt.Errorf("cache key not found")
	ErrCacheMarshaling  = fmt.Errorf("failed to marshal data")
	ErrCacheConnection  = fmt.Errorf("redis connection error")
)

const (
	// Key prefixes
	ProductKeyPrefix      = "product:"
	ProductListKeyPrefix  = "product:list:"
	CategoryKeyPrefix     = "category:"
	CategoryListKeyPrefix = "category:list:"
	BrandKeyPrefix        = "brand:"
	BrandListKeyPrefix    = "brand:list:"

	// TTL constants
	DefaultTTL  = 15 * time.Minute
	ExtendedTTL = 1 * time.Hour
	ShortTTL    = 5 * time.Minute

	// Operation timeouts
	DefaultTimeout  = 5 * time.Second
	ExtendedTimeout = 10 * time.Second

	// Batch operation sizes
	MaxBatchSize = 100
)

type CacheManager struct {
	redis      *redis.Client
	logger     *zap.Logger
	defaultTTL time.Duration
}

type CacheOptions struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
	TTL      time.Duration
}

func NewCacheManager(opts CacheOptions) (*CacheManager, error) {
	if opts.TTL == 0 {
		opts.TTL = DefaultTTL
	}

	client := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
		PoolSize: opts.PoolSize,

		// Add connection pool settings
		MinIdleConns: 10,
		MaxConnAge:   time.Hour,
		IdleTimeout:  10 * time.Minute,
	})

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	return &CacheManager{
		redis:      client,
		defaultTTL: opts.TTL,
	}, nil
}

// Add context timeout wrapper
func (cm *CacheManager) withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

// Enhanced Product methods
func (cm *CacheManager) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	ctx, cancel := cm.withTimeout(ctx, DefaultTimeout)
	defer cancel()

	key := fmt.Sprintf("%s%s", ProductKeyPrefix, id)
	data, err := cm.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheKeyNotFound
	} else if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var product models.Product
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCacheMarshaling, err)
	}

	// Get variants from cache
	variants, err := cm.GetProductVariants(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert []*models.ProductVariant to []models.ProductVariant
	if variants != nil {
		productVariants := make([]models.ProductVariant, len(variants))
		for i, v := range variants {
			productVariants[i] = *v
		}
		product.Variants = productVariants
	}

	return &product, nil
}

// Batch operations
func (cm *CacheManager) GetProductsBatch(ctx context.Context, ids []string) (map[string]*models.Product, error) {
	ctx, cancel := cm.withTimeout(ctx, ExtendedTimeout)
	defer cancel()

	pipe := cm.redis.Pipeline()
	cmds := make(map[string]*redis.StringCmd, len(ids))

	for _, id := range ids {
		key := fmt.Sprintf("%s%s", ProductKeyPrefix, id)
		cmds[id] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("pipeline exec error: %w", err)
	}

	results := make(map[string]*models.Product)
	for id, cmd := range cmds {
		data, err := cmd.Bytes()
		if err == redis.Nil {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("redis get error for id %s: %w", id, err)
		}

		var product models.Product
		if err := json.Unmarshal(data, &product); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrCacheMarshaling, err)
		}
		results[id] = &product
	}

	return results, nil
}

// Cache maintenance methods
func (cm *CacheManager) ClearExpiredKeys(ctx context.Context) error {
	return cm.redis.Do(ctx, "MEMORY", "PURGE").Err()
}

func (cm *CacheManager) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	_, err := cm.redis.Info(ctx, "memory", "stats").Result()
	if err != nil {
		return nil, err
	}

	// Parse INFO command output into map
	stats := make(map[string]interface{})
	// ... parse logic here
	return stats, nil
}

// Close connection
func (cm *CacheManager) Close() error {
	return cm.redis.Close()
}

// Health check
func (cm *CacheManager) HealthCheck(ctx context.Context) error {
	ctx, cancel := cm.withTimeout(ctx, DefaultTimeout)
	defer cancel()

	return cm.redis.Ping(ctx).Err()
}

func (cm *CacheManager) SetProduct(ctx context.Context, product *models.Product) error {
	key := fmt.Sprintf("%s%s", ProductKeyPrefix, product.ID)

	// Cache product without variants
	productCopy := *product
	productCopy.Variants = nil
	data, err := json.Marshal(productCopy)
	if err != nil {
		return err
	}

	pipe := cm.redis.Pipeline()
	pipe.Set(ctx, key, data, DefaultTTL)

	// Cache variants separately
	if len(product.Variants) > 0 {
		variantsKey := fmt.Sprintf("product:variants:%s", product.ID)
		// Convert []models.ProductVariant to []*models.ProductVariant
		variantPtrs := make([]*models.ProductVariant, len(product.Variants))
		for i := range product.Variants {
			variantPtrs[i] = &product.Variants[i]
		}
		variantsData, err := json.Marshal(variantPtrs)
		if err != nil {
			return err
		}
		pipe.Set(ctx, variantsKey, variantsData, DefaultTTL)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (cm *CacheManager) GetProductList(ctx context.Context, filterKey string) ([]*models.Product, error) {
	key := fmt.Sprintf("%s%s", ProductListKeyPrefix, filterKey)
	data, err := cm.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var products []*models.Product
	if err := json.Unmarshal(data, &products); err != nil {
		return nil, err
	}

	return products, nil
}

func (cm *CacheManager) SetProductList(ctx context.Context, filterKey string, products []*models.Product) error {
	key := fmt.Sprintf("%s%s", ProductListKeyPrefix, filterKey)
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}

	return cm.redis.Set(ctx, key, data, DefaultTTL).Err()
}

func (cm *CacheManager) InvalidateProduct(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", ProductKeyPrefix, id)
	return cm.redis.Del(ctx, key).Err()
}

func (cm *CacheManager) InvalidateProductLists(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", ProductListKeyPrefix)
	iter := cm.redis.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := cm.redis.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// Category-related methods
func (cm *CacheManager) GetCategory(ctx context.Context, id string) (*models.Category, error) {
	key := fmt.Sprintf("%s%s", CategoryKeyPrefix, id)
	data, err := cm.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var category models.Category
	if err := json.Unmarshal(data, &category); err != nil {
		return nil, err
	}

	return &category, nil
}

func (cm *CacheManager) SetCategory(ctx context.Context, category *models.Category) error {
	key := fmt.Sprintf("%s%s", CategoryKeyPrefix, category.ID)
	data, err := json.Marshal(category)
	if err != nil {
		return err
	}

	return cm.redis.Set(ctx, key, data, DefaultTTL).Err()
}

func (cm *CacheManager) GetCategoryList(ctx context.Context, filterKey string) ([]*models.Category, error) {
	key := fmt.Sprintf("%s%s", CategoryListKeyPrefix, filterKey)
	data, err := cm.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var categories []*models.Category
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func (cm *CacheManager) SetCategoryList(ctx context.Context, filterKey string, categories []*models.Category) error {
	key := fmt.Sprintf("%s%s", CategoryListKeyPrefix, filterKey)
	data, err := json.Marshal(categories)
	if err != nil {
		return err
	}

	return cm.redis.Set(ctx, key, data, DefaultTTL).Err()
}

func (cm *CacheManager) InvalidateCategory(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", CategoryKeyPrefix, id)
	return cm.redis.Del(ctx, key).Err()
}

func (cm *CacheManager) InvalidateCategoryLists(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", CategoryListKeyPrefix)
	iter := cm.redis.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := cm.redis.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// --- Brand Cache Methods ---

const (
	BrandCacheKeyByID   = BrandKeyPrefix + "%s"
	BrandCacheKeyBySlug = BrandKeyPrefix + "slug:%s"
)

func (cm *CacheManager) GetBrand(ctx context.Context, key string) (*models.Brand, error) {
	data, err := cm.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err // Includes redis.Nil for cache miss
	}

	var brand models.Brand
	if err := json.Unmarshal(data, &brand); err != nil {
		return nil, fmt.Errorf("failed to unmarshal brand data from cache: %w", err)
	}

	return &brand, nil
}

func (cm *CacheManager) SetBrand(ctx context.Context, key string, brand *models.Brand) error {
	data, err := json.Marshal(brand)
	if err != nil {
		return fmt.Errorf("failed to marshal brand data for cache: %w", err)
	}

	return cm.redis.Set(ctx, key, data, DefaultTTL).Err()
}

func (cm *CacheManager) GetBrandList(ctx context.Context, filterKey string) ([]*models.Brand, error) {
	key := fmt.Sprintf("%s%s", BrandListKeyPrefix, filterKey)
	data, err := cm.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var brands []*models.Brand
	if err := json.Unmarshal(data, &brands); err != nil {
		return nil, fmt.Errorf("failed to unmarshal brand list data from cache: %w", err)
	}

	return brands, nil
}

func (cm *CacheManager) SetBrandList(ctx context.Context, filterKey string, brands []*models.Brand) error {
	key := fmt.Sprintf("%s%s", BrandListKeyPrefix, filterKey)
	data, err := json.Marshal(brands)
	if err != nil {
		return fmt.Errorf("failed to marshal brand list data for cache: %w", err)
	}

	return cm.redis.Set(ctx, key, data, DefaultTTL).Err()
}

func (cm *CacheManager) InvalidateBrand(ctx context.Context, id string, slug string) error {
	// Invalidate both ID and Slug keys
	keysToDelete := []string{
		fmt.Sprintf(BrandCacheKeyByID, id),
		fmt.Sprintf(BrandCacheKeyBySlug, slug),
	}

	// Remove empty slug key if slug is empty
	if slug == "" && len(keysToDelete) > 1 { // Ensure there's more than one key before slicing
		keysToDelete = keysToDelete[:1]
	} else if slug == "" { // Handle case where only ID key exists initially
		keysToDelete = []string{fmt.Sprintf(BrandCacheKeyByID, id)}
	}

	if len(keysToDelete) > 0 {
		if err := cm.redis.Del(ctx, keysToDelete...).Err(); err != nil {
			return fmt.Errorf("failed to invalidate brand cache for ID %s: %w", id, err)
		}
	}
	return nil
}

func (cm *CacheManager) InvalidateBrandLists(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", BrandListKeyPrefix)
	iter := cm.redis.Scan(ctx, 0, pattern, 0).Iterator()

	keysToDelete := []string{}
	for iter.Next(ctx) {
		keysToDelete = append(keysToDelete, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed scanning brand list keys for invalidation: %w", err)
	}

	if len(keysToDelete) > 0 {
		if err := cm.redis.Del(ctx, keysToDelete...).Err(); err != nil {
			// Log or handle partial failure? For now, return the error.
			return fmt.Errorf("failed to delete brand list keys: %w", err)
		}
	}
	return nil
}

// BatchInvalidate invalidates multiple cache entries atomically
func (cm *CacheManager) BatchInvalidate(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	return cm.redis.Del(ctx, keys...).Err()
}

// InvalidateProductAndRelated invalidates a product and all related caches
func (cm *CacheManager) InvalidateProductAndRelated(ctx context.Context, productID string) error {
	ctx, cancel := cm.withTimeout(ctx, ExtendedTimeout)
	defer cancel()

	// Collect all keys to invalidate
	keysToDelete := []string{
		fmt.Sprintf("%s%s", ProductKeyPrefix, productID),
	}

	// Invalidate product lists
	if err := cm.InvalidateProductLists(ctx); err != nil {
		return fmt.Errorf("failed to invalidate product lists: %w", err)
	}

	// Invalidate related category lists (since product count might change)
	if err := cm.InvalidateCategoryLists(ctx); err != nil {
		return fmt.Errorf("failed to invalidate category lists: %w", err)
	}

	// Batch delete collected keys
	return cm.BatchInvalidate(ctx, keysToDelete)
}

// InvalidateByPattern invalidates all keys matching a pattern
func (cm *CacheManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	ctx, cancel := cm.withTimeout(ctx, ExtendedTimeout)
	defer cancel()

	iter := cm.redis.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		// Batch delete in chunks to avoid memory issues with large datasets
		if len(keys) >= MaxBatchSize {
			if err := cm.BatchInvalidate(ctx, keys); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("error scanning keys: %w", err)
	}

	if len(keys) > 0 {
		return cm.BatchInvalidate(ctx, keys)
	}

	return nil
}

// InvalidateProductsByCategory invalidates all product caches related to a category
func (cm *CacheManager) InvalidateProductsByCategory(ctx context.Context, categoryID string) error {
	// Invalidate category-specific product lists
	pattern := fmt.Sprintf("%s*category:%s*", ProductListKeyPrefix, categoryID)
	if err := cm.InvalidateByPattern(ctx, pattern); err != nil {
		return fmt.Errorf("failed to invalidate category products: %w", err)
	}

	// Also invalidate general product lists as they might include products from this category
	return cm.InvalidateProductLists(ctx)
}

func (cm *CacheManager) GetProductVariants(ctx context.Context, productID string) ([]*models.ProductVariant, error) {
	key := fmt.Sprintf("product:variants:%s", productID)
	data, err := cm.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var variants []*models.ProductVariant
	if err := json.Unmarshal(data, &variants); err != nil {
		return nil, err
	}

	return variants, nil
}
