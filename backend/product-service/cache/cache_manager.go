package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/louai60/e-commerce_project/backend/product-service/models"
)

const (
    ProductKeyPrefix     = "product:"
    ProductListKeyPrefix = "product:list:"
    DefaultTTL          = 15 * time.Minute
)

type CacheManager struct {
    client *redis.Client
}

func NewCacheManager(redisAddr string) (*CacheManager, error) {
    client := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    // Test connection
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    return &CacheManager{client: client}, nil
}

// Product-specific cache methods
func (cm *CacheManager) GetProduct(ctx context.Context, id string) (*models.Product, error) {
    key := fmt.Sprintf("%s%s", ProductKeyPrefix, id)
    data, err := cm.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, err
    }

    var product models.Product
    if err := json.Unmarshal(data, &product); err != nil {
        return nil, err
    }

    return &product, nil
}

func (cm *CacheManager) SetProduct(ctx context.Context, product *models.Product) error {
    key := fmt.Sprintf("%s%s", ProductKeyPrefix, product.ID)
    data, err := json.Marshal(product)
    if err != nil {
        return err
    }

    return cm.client.Set(ctx, key, data, DefaultTTL).Err()
}

func (cm *CacheManager) GetProductList(ctx context.Context, filters string) ([]*models.Product, error) {
    key := fmt.Sprintf("%s%s", ProductListKeyPrefix, filters)
    data, err := cm.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, err
    }

    var products []*models.Product
    if err := json.Unmarshal(data, &products); err != nil {
        return nil, err
    }

    return products, nil
}

func (cm *CacheManager) SetProductList(ctx context.Context, filters string, products []*models.Product) error {
    key := fmt.Sprintf("%s%s", ProductListKeyPrefix, filters)
    data, err := json.Marshal(products)
    if err != nil {
        return err
    }

    return cm.client.Set(ctx, key, data, DefaultTTL).Err()
}

func (cm *CacheManager) InvalidateProduct(ctx context.Context, id string) error {
    key := fmt.Sprintf("%s%s", ProductKeyPrefix, id)
    return cm.client.Del(ctx, key).Err()
}

func (cm *CacheManager) InvalidateProductLists(ctx context.Context) error {
    pattern := fmt.Sprintf("%s*", ProductListKeyPrefix)
    iter := cm.client.Scan(ctx, 0, pattern, 0).Iterator()
    
    for iter.Next(ctx) {
        if err := cm.client.Del(ctx, iter.Val()).Err(); err != nil {
            return err
        }
    }
    
    return iter.Err()
}