package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// TTLProvider defines an interface for getting TTL based on key type
type TTLProvider interface {
	GetTTL(keyType string) time.Duration
}

// DefaultTTLProvider provides default TTL values for different key types
type DefaultTTLProvider struct {
	defaultTTL time.Duration
	ttlMap     map[string]time.Duration
}

// NewDefaultTTLProvider creates a new TTL provider with default values
func NewDefaultTTLProvider(defaultTTL time.Duration) *DefaultTTLProvider {
	return &DefaultTTLProvider{
		defaultTTL: defaultTTL,
		ttlMap: map[string]time.Duration{
			"product":        15 * time.Minute,
			"product_list":   5 * time.Minute,
			"category":       1 * time.Hour,
			"category_list":  30 * time.Minute,
			"brand":          1 * time.Hour,
			"brand_list":     30 * time.Minute,
			"user":           30 * time.Minute,
			"token":          24 * time.Hour,
			"session":        7 * 24 * time.Hour,
			"high_frequency": 1 * time.Minute,
			"low_frequency":  2 * time.Hour,
			"static_data":    24 * time.Hour,
		},
	}
}

// GetTTL returns the TTL for a given key type
func (p *DefaultTTLProvider) GetTTL(keyType string) time.Duration {
	if ttl, ok := p.ttlMap[keyType]; ok {
		return ttl
	}
	return p.defaultTTL
}

// SetTTL sets a custom TTL for a key type
func (p *DefaultTTLProvider) SetTTL(keyType string, ttl time.Duration) {
	p.ttlMap[keyType] = ttl
}

// TieredCache implements a two-level cache with memory and Redis
type TieredCache struct {
	memoryCache    *MemoryCache
	redisClient    *redis.Client
	ttlProvider    TTLProvider
	keyMutexes     *sync.Map // For cache stampede protection
	metrics        *CacheMetrics
	circuitBreaker *CircuitBreaker
}

// TieredCacheOptions defines options for creating a tiered cache
type TieredCacheOptions struct {
	RedisOptions *redis.Options
	DefaultTTL   time.Duration
	// Circuit breaker options
	FailureThreshold         int64
	ResetTimeout             time.Duration
	HalfOpenSuccessThreshold int64
}

// NewTieredCache creates a new tiered cache with memory and Redis layers
func NewTieredCache(opts TieredCacheOptions) (*TieredCache, error) {
	// Create Redis client
	redisClient := redis.NewClient(opts.RedisOptions)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Create memory cache
	memoryCache := NewMemoryCache()

	// Create TTL provider
	ttlProvider := NewDefaultTTLProvider(opts.DefaultTTL)

	// Create metrics collector
	metrics := NewCacheMetrics()

	// Create circuit breaker
	circuitBreaker := NewCircuitBreaker(CircuitBreakerOptions{
		FailureThreshold:         opts.FailureThreshold,
		ResetTimeout:             opts.ResetTimeout,
		HalfOpenSuccessThreshold: opts.HalfOpenSuccessThreshold,
	})

	return &TieredCache{
		memoryCache:    memoryCache,
		redisClient:    redisClient,
		ttlProvider:    ttlProvider,
		keyMutexes:     &sync.Map{},
		metrics:        metrics,
		circuitBreaker: circuitBreaker,
	}, nil
}

// Get retrieves an item from the cache, trying memory first, then Redis
func (c *TieredCache) Get(ctx context.Context, key string, keyType string) ([]byte, error) {
	startTime := time.Now()
	defer func() {
		c.metrics.RecordLatency(time.Since(startTime).Nanoseconds())
	}()

	// Try L1 (memory) cache first
	if value, found := c.memoryCache.Get(key); found {
		if data, ok := value.([]byte); ok {
			c.metrics.RecordHit()
			return data, nil
		}
	}

	// Try L2 (Redis) cache with circuit breaker
	var data []byte
	err := c.circuitBreaker.Execute(func() error {
		var redisErr error
		data, redisErr = c.redisClient.Get(ctx, key).Bytes()
		return redisErr
	})

	// Handle circuit breaker open state
	if err == ErrCircuitOpen {
		c.metrics.RecordError()
		return nil, fmt.Errorf("redis circuit breaker open: %w", err)
	}

	if err == nil {
		// Store in memory cache with a shorter TTL
		c.memoryCache.Set(key, data, 30*time.Second)
		c.metrics.RecordHit()
		return data, nil
	}

	if err != redis.Nil {
		c.metrics.RecordError()
		return nil, fmt.Errorf("redis error: %w", err)
	}

	c.metrics.RecordMiss()
	return nil, fmt.Errorf("key not found in cache")
}

// GetObject retrieves and unmarshals an object from the cache
func (c *TieredCache) GetObject(ctx context.Context, key string, keyType string, dest interface{}) error {
	data, err := c.Get(ctx, key, keyType)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Set stores an item in both memory and Redis caches
func (c *TieredCache) Set(ctx context.Context, key string, value []byte, keyType string) error {
	startTime := time.Now()
	defer func() {
		c.metrics.RecordLatency(time.Since(startTime).Nanoseconds())
	}()

	// Get TTL based on key type
	ttl := c.ttlProvider.GetTTL(keyType)

	// Store in Redis (L2) with circuit breaker
	err := c.circuitBreaker.Execute(func() error {
		return c.redisClient.Set(ctx, key, value, ttl).Err()
	})

	// Handle circuit breaker open state
	if err == ErrCircuitOpen {
		c.metrics.RecordError()
		return fmt.Errorf("redis circuit breaker open: %w", err)
	}

	if err != nil {
		c.metrics.RecordError()
		return fmt.Errorf("redis set error: %w", err)
	}

	// Store in memory (L1) with a shorter TTL
	memoryTTL := 30 * time.Second
	if memoryTTL > ttl {
		memoryTTL = ttl
	}
	c.memoryCache.Set(key, value, memoryTTL)

	return nil
}

// SetObject marshals and stores an object in the cache
func (c *TieredCache) SetObject(ctx context.Context, key string, value interface{}, keyType string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	return c.Set(ctx, key, data, keyType)
}

// Delete removes an item from both memory and Redis caches
func (c *TieredCache) Delete(ctx context.Context, key string) error {
	startTime := time.Now()
	defer func() {
		c.metrics.RecordLatency(time.Since(startTime).Nanoseconds())
	}()

	// Delete from memory cache
	c.memoryCache.Delete(key)

	// Delete from Redis with circuit breaker
	err := c.circuitBreaker.Execute(func() error {
		return c.redisClient.Del(ctx, key).Err()
	})

	// Handle circuit breaker open state
	if err == ErrCircuitOpen {
		c.metrics.RecordError()
		return fmt.Errorf("redis circuit breaker open: %w", err)
	}

	if err != nil {
		c.metrics.RecordError()
		return fmt.Errorf("redis delete error: %w", err)
	}

	return nil
}

// DeleteByPattern removes items matching a pattern from both caches
func (c *TieredCache) DeleteByPattern(ctx context.Context, pattern string) error {
	startTime := time.Now()
	defer func() {
		c.metrics.RecordLatency(time.Since(startTime).Nanoseconds())
	}()

	// Use circuit breaker for the scan operation
	var keys []string
	err := c.circuitBreaker.Execute(func() error {
		// Scan for keys matching the pattern
		iter := c.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		return iter.Err()
	})

	// Handle circuit breaker open state
	if err == ErrCircuitOpen {
		c.metrics.RecordError()
		return fmt.Errorf("redis circuit breaker open: %w", err)
	}

	if err != nil {
		c.metrics.RecordError()
		return fmt.Errorf("redis scan error: %w", err)
	}

	// Delete keys in batches to avoid overwhelming Redis
	for i := 0; i < len(keys); i += 100 {
		end := i + 100
		if end > len(keys) {
			end = len(keys)
		}
		batch := keys[i:end]

		if len(batch) > 0 {
			// Delete from Redis with circuit breaker
			err := c.circuitBreaker.Execute(func() error {
				return c.redisClient.Del(ctx, batch...).Err()
			})

			if err != nil {
				c.metrics.RecordError()
				return fmt.Errorf("redis batch delete error: %w", err)
			}

			// Also delete from memory cache
			for _, key := range batch {
				c.memoryCache.Delete(key)
			}
		}
	}

	return nil
}

// GetOrSet implements cache stampede protection using a mutex
func (c *TieredCache) GetOrSet(ctx context.Context, key string, keyType string, getter func() (interface{}, error)) ([]byte, error) {
	// Try to get from cache first
	data, err := c.Get(ctx, key, keyType)
	if err == nil {
		return data, nil
	}

	// Cache miss, use mutex to prevent stampede
	mutexInterface, _ := c.keyMutexes.LoadOrStore(key, &sync.Mutex{})
	mutex := mutexInterface.(*sync.Mutex)

	mutex.Lock()
	defer func() {
		mutex.Unlock()
		c.keyMutexes.Delete(key) // Clean up mutex after use
	}()

	// Try again after acquiring lock (another goroutine might have populated the cache)
	data, err = c.Get(ctx, key, keyType)
	if err == nil {
		return data, nil
	}

	// Still not in cache, call getter function
	value, err := getter()
	if err != nil {
		return nil, err
	}

	// Marshal the value
	data, err = json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	// Store in cache
	if err := c.Set(ctx, key, data, keyType); err != nil {
		return nil, err
	}

	return data, nil
}

// GetObjectOrSet retrieves an object with stampede protection
func (c *TieredCache) GetObjectOrSet(ctx context.Context, key string, keyType string, dest interface{}, getter func() (interface{}, error)) error {
	data, err := c.GetOrSet(ctx, key, keyType, getter)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Close closes the Redis client
func (c *TieredCache) Close() error {
	return c.redisClient.Close()
}

// HealthCheck checks if the Redis connection is healthy
func (c *TieredCache) HealthCheck(ctx context.Context) error {
	return c.redisClient.Ping(ctx).Err()
}

// ClearMemoryCache clears only the memory cache
func (c *TieredCache) ClearMemoryCache() {
	c.memoryCache.Clear()
}

// GetMemoryCacheStats returns statistics about the memory cache
func (c *TieredCache) GetMemoryCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"count": c.memoryCache.Count(),
	}
}

// GetMetrics returns all cache metrics
func (c *TieredCache) GetMetrics() map[string]interface{} {
	metrics := c.metrics.GetMetrics()

	// Add circuit breaker metrics
	metrics["circuit_breaker"] = c.circuitBreaker.GetMetrics()

	// Add memory cache stats
	metrics["memory_cache"] = c.GetMemoryCacheStats()

	return metrics
}

// ResetMetrics resets all cache metrics
func (c *TieredCache) ResetMetrics() {
	c.metrics.Reset()
}

// GetCircuitBreakerState returns the current state of the circuit breaker
func (c *TieredCache) GetCircuitBreakerState() CircuitState {
	return c.circuitBreaker.GetState()
}

// ResetCircuitBreaker resets the circuit breaker to closed state
func (c *TieredCache) ResetCircuitBreaker() {
	c.circuitBreaker.Reset()
}
