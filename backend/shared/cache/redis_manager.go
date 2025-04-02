package cache

import (
    "context"
    "encoding/json"
    "time"
    "github.com/go-redis/redis/v8"
)

type CacheManager struct {
    client *redis.Client
    defaultTTL time.Duration
}

func NewCacheManager(addr string) (*CacheManager, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: "", // set if required
        DB:       0,
    })

    // Test connection
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }

    return &CacheManager{
        client:     client,
        defaultTTL: 24 * time.Hour,
    }, nil
}

func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    if ttl == 0 {
        ttl = cm.defaultTTL
    }

    return cm.client.Set(ctx, key, data, ttl).Err()
}

func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
    data, err := cm.client.Get(ctx, key).Bytes()
    if err != nil {
        return err
    }

    return json.Unmarshal(data, dest)
}

func (cm *CacheManager) Delete(ctx context.Context, keys ...string) error {
    return cm.client.Del(ctx, keys...).Err()
}

func (cm *CacheManager) Clear(ctx context.Context, pattern string) error {
    iter := cm.client.Scan(ctx, 0, pattern, 0).Iterator()
    for iter.Next(ctx) {
        if err := cm.client.Del(ctx, iter.Val()).Err(); err != nil {
            return err
        }
    }
    return iter.Err()
}