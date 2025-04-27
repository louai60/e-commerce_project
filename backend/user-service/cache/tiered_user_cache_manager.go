package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	sharedCache "github.com/louai60/e-commerce_project/backend/shared/cache"
	"github.com/louai60/e-commerce_project/backend/user-service/models"
	"go.uber.org/zap"
)

// TieredUserCacheManager implements a two-level cache with memory and Redis for user data
type TieredUserCacheManager struct {
	tieredCache *sharedCache.TieredCache
	logger      *zap.Logger
}

// TieredUserCacheOptions defines options for creating a tiered user cache manager
type TieredUserCacheOptions struct {
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

// NewTieredUserCacheManager creates a new tiered user cache manager
func NewTieredUserCacheManager(opts TieredUserCacheOptions) (*TieredUserCacheManager, error) {
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
	tieredCache, err := sharedCache.NewTieredCache(sharedCache.TieredCacheOptions{
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

	return &TieredUserCacheManager{
		tieredCache: tieredCache,
		logger:      opts.Logger,
	}, nil
}

// GetUser retrieves a user from the cache
func (cm *TieredUserCacheManager) GetUser(ctx context.Context, userID string) (*models.User, error) {
	key := fmt.Sprintf("%s%s", UserKeyPrefix, userID)

	var user models.User
	err := cm.tieredCache.GetObject(ctx, key, "user", &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// SetUser stores a user in the cache
func (cm *TieredUserCacheManager) SetUser(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("%s%s", UserKeyPrefix, user.UserID)
	return cm.tieredCache.SetObject(ctx, key, user, "user")
}

// StoreToken stores a token in the cache
func (cm *TieredUserCacheManager) StoreToken(ctx context.Context, userID, tokenType, token string) error {
	key := fmt.Sprintf("%s%s:%s", TokenKeyPrefix, userID, tokenType)
	return cm.tieredCache.Set(ctx, key, []byte(token), "token")
}

// GetToken retrieves a token from the cache
func (cm *TieredUserCacheManager) GetToken(ctx context.Context, userID, tokenType string) (string, error) {
	key := fmt.Sprintf("%s%s:%s", TokenKeyPrefix, userID, tokenType)

	data, err := cm.tieredCache.Get(ctx, key, "token")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// InvalidateToken removes a token from the cache
func (cm *TieredUserCacheManager) InvalidateToken(ctx context.Context, userID, tokenType string) error {
	key := fmt.Sprintf("%s%s:%s", TokenKeyPrefix, userID, tokenType)
	return cm.tieredCache.Delete(ctx, key)
}

// StoreSession stores a session in the cache
func (cm *TieredUserCacheManager) StoreSession(ctx context.Context, sessionID string, userData map[string]interface{}) error {
	key := fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)
	return cm.tieredCache.SetObject(ctx, key, userData, "session")
}

// GetSession retrieves a session from the cache
func (cm *TieredUserCacheManager) GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)

	var userData map[string]interface{}
	err := cm.tieredCache.GetObject(ctx, key, "session", &userData)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

// InvalidateSession removes a session from the cache
func (cm *TieredUserCacheManager) InvalidateSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)
	return cm.tieredCache.Delete(ctx, key)
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

// WarmupCache warms up the cache with critical user data
func (cm *TieredUserCacheManager) WarmupCache(ctx context.Context) (*WarmupResult, error) {
	cm.logger.Info("Starting cache warm-up for critical user data")
	startTime := time.Now()
	result := &WarmupResult{}

	// This would typically warm up frequently accessed users, tokens, etc.
	// For now, we'll just return an empty result as a placeholder

	result.Duration = time.Since(startTime)
	return result, nil
}

// GetCacheMetrics returns metrics about the cache
func (cm *TieredUserCacheManager) GetCacheMetrics(ctx context.Context) (map[string]interface{}, error) {
	return cm.tieredCache.GetMetrics(), nil
}

// ResetCacheMetrics resets the cache metrics
func (cm *TieredUserCacheManager) ResetCacheMetrics() {
	cm.tieredCache.ResetMetrics()
}

// Close closes the cache manager
func (cm *TieredUserCacheManager) Close() error {
	return cm.tieredCache.Close()
}
