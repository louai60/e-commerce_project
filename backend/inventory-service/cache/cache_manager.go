package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	// Default TTL for cache entries
	DefaultTTL = 15 * time.Minute
	// Default timeout for cache operations
	DefaultTimeout = 5 * time.Second
	// Extended timeout for batch operations
	ExtendedTimeout = 10 * time.Second
	// Key prefixes
	InventoryKeyPrefix = "inventory:"
	WarehouseKeyPrefix = "warehouse:"
)

// CacheManager handles caching operations for inventory service
type CacheManager struct {
	redis      *redis.Client
	defaultTTL time.Duration
	logger     *zap.Logger
}

// CacheOptions defines options for creating a cache manager
type CacheOptions struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
	TTL      time.Duration
	Logger   *zap.Logger
}

// NewCacheManager creates a new cache manager
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
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &CacheManager{
		redis:      client,
		defaultTTL: opts.TTL,
		logger:     opts.Logger,
	}, nil
}

// Helper method to add timeout to context if not already present
func (cm *CacheManager) withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

// Close closes the Redis connection
func (cm *CacheManager) Close() error {
	return cm.redis.Close()
}

// Generic Get method
func (cm *CacheManager) Get(ctx context.Context, key string, value interface{}) error {
	ctx, cancel := cm.withTimeout(ctx, DefaultTimeout)
	defer cancel()

	data, err := cm.redis.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
}

// Generic Set method
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}) error {
	ctx, cancel := cm.withTimeout(ctx, DefaultTimeout)
	defer cancel()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cm.redis.Set(ctx, key, data, cm.defaultTTL).Err()
}

// Invalidate removes a key from the cache
func (cm *CacheManager) Invalidate(ctx context.Context, key string) error {
	ctx, cancel := cm.withTimeout(ctx, DefaultTimeout)
	defer cancel()

	return cm.redis.Del(ctx, key).Err()
}
