package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WarmupConfig defines configuration for cache warm-up
type WarmupConfig struct {
	// Enabled indicates whether warm-up is enabled
	Enabled bool
	// Keys is a map of key types to slices of keys to warm up
	Keys map[string][]string
	// Getters is a map of key types to getter functions
	Getters map[string]func(ctx context.Context, key string) (interface{}, error)
	// Concurrency is the number of concurrent warm-up operations
	Concurrency int
	// Logger is the logger to use
	Logger *zap.Logger
}

// WarmupResult contains the results of a warm-up operation
type WarmupResult struct {
	// SuccessCount is the number of successfully warmed up keys
	SuccessCount int
	// ErrorCount is the number of keys that failed to warm up
	ErrorCount int
	// Duration is the total duration of the warm-up operation
	Duration time.Duration
	// Errors is a map of keys to errors
	Errors map[string]error
}

// WarmupCache warms up the cache with the given configuration
func WarmupCache(ctx context.Context, cache *TieredCache, config WarmupConfig) (*WarmupResult, error) {
	if !config.Enabled {
		return &WarmupResult{}, nil
	}

	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	if config.Concurrency <= 0 {
		config.Concurrency = 5
	}

	startTime := time.Now()
	result := &WarmupResult{
		Errors: make(map[string]error),
	}

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, config.Concurrency)
	var wg sync.WaitGroup

	// Process each key type
	for keyType, keys := range config.Keys {
		getter, ok := config.Getters[keyType]
		if !ok {
			config.Logger.Warn("No getter function for key type", zap.String("keyType", keyType))
			continue
		}

		// Process each key
		for _, key := range keys {
			wg.Add(1)
			sem <- struct{}{} // Acquire semaphore

			go func(keyType, key string) {
				defer func() {
					<-sem // Release semaphore
					wg.Done()
				}()

				// Check if key already exists in cache
				_, err := cache.Get(ctx, key, keyType)
				if err == nil {
					config.Logger.Debug("Key already in cache", zap.String("key", key), zap.String("keyType", keyType))
					return
				}

				// Get data from source
				config.Logger.Debug("Warming up key", zap.String("key", key), zap.String("keyType", keyType))
				data, err := getter(ctx, key)
				if err != nil {
					config.Logger.Error("Failed to get data for warm-up",
						zap.String("key", key),
						zap.String("keyType", keyType),
						zap.Error(err))
					
					result.ErrorCount++
					result.Errors[key] = err
					return
				}

				// Store in cache
				err = cache.SetObject(ctx, key, data, keyType)
				if err != nil {
					config.Logger.Error("Failed to store data in cache",
						zap.String("key", key),
						zap.String("keyType", keyType),
						zap.Error(err))
					
					result.ErrorCount++
					result.Errors[key] = err
					return
				}

				result.SuccessCount++
				config.Logger.Debug("Successfully warmed up key", zap.String("key", key), zap.String("keyType", keyType))
			}(keyType, key)
		}
	}

	// Wait for all warm-up operations to complete
	wg.Wait()
	result.Duration = time.Since(startTime)

	config.Logger.Info("Cache warm-up completed",
		zap.Int("successCount", result.SuccessCount),
		zap.Int("errorCount", result.ErrorCount),
		zap.Duration("duration", result.Duration))

	return result, nil
}

// WarmupCacheWithRetry warms up the cache with retries
func WarmupCacheWithRetry(ctx context.Context, cache *TieredCache, config WarmupConfig, retries int, retryDelay time.Duration) (*WarmupResult, error) {
	var lastErr error
	var result *WarmupResult

	for i := 0; i <= retries; i++ {
		if i > 0 {
			config.Logger.Info("Retrying cache warm-up", zap.Int("attempt", i+1), zap.Int("maxRetries", retries))
			time.Sleep(retryDelay)
		}

		// Filter out keys that have already been successfully warmed up
		if result != nil {
			for keyType, keys := range config.Keys {
				var remainingKeys []string
				for _, key := range keys {
					if _, hasError := result.Errors[key]; hasError {
						remainingKeys = append(remainingKeys, key)
					}
				}
				config.Keys[keyType] = remainingKeys
			}
		}

		var err error
		result, err = WarmupCache(ctx, cache, config)
		if err != nil {
			lastErr = err
			continue
		}

		if result.ErrorCount == 0 {
			return result, nil
		}

		lastErr = fmt.Errorf("failed to warm up %d keys", result.ErrorCount)
	}

	return result, lastErr
}
