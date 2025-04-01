package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/go-redis/redis/v8"
)

type AdminKeyManager struct {
	redis *redis.Client
}

func generateSecureRandomKey(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err) // handle error appropriately in production
	}
	return hex.EncodeToString(bytes)
}

func (m *AdminKeyManager) RotateAdminKey() (string, error) {
	ctx := context.Background()
	newKey := generateSecureRandomKey(32)
	
	pipe := m.redis.Pipeline()
	// Store new key with 90-day expiration
	pipe.Set(ctx, "admin_key:current", newKey, 90*24*time.Hour)
	// Keep old key valid for 24 hours during transition
	pipe.Rename(ctx, "admin_key:current", "admin_key:previous")
	pipe.Expire(ctx, "admin_key:previous", 24*time.Hour)
	
	_, err := pipe.Exec(ctx)
	return newKey, err
}

func (m *AdminKeyManager) ValidateAdminKey(key string) bool {
	ctx := context.Background()
	// Check both current and previous keys
	currentKey := m.redis.Get(ctx, "admin_key:current").Val()
	previousKey := m.redis.Get(ctx, "admin_key:previous").Val()
	
	return key == currentKey || key == previousKey
}
