package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/louai60/e-commerce_project/backend/user-service/models"
)

const (
    UserKeyPrefix     = "user:"
    TokenKeyPrefix    = "token:"
    SessionKeyPrefix  = "session:"
    DefaultUserTTL    = 30 * time.Minute
    DefaultTokenTTL   = 24 * time.Hour
    DefaultSessionTTL = 7 * 24 * time.Hour
)

type UserCacheManager struct {
    client *redis.Client
}

func NewUserCacheManager(redisAddr string) (*UserCacheManager, error) {
    client := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    // Test connection
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    return &UserCacheManager{client: client}, nil
}

// User profile caching
func (cm *UserCacheManager) GetUser(ctx context.Context, userID string) (*models.User, error) {
    key := fmt.Sprintf("%s%s", UserKeyPrefix, userID)
    data, err := cm.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, err
    }

    var user models.User
    if err := json.Unmarshal(data, &user); err != nil {
        return nil, err
    }

    return &user, nil
}

func (cm *UserCacheManager) SetUser(ctx context.Context, user *models.User) error {
    key := fmt.Sprintf("%s%s", UserKeyPrefix, user.UserID)
    data, err := json.Marshal(user)
    if err != nil {
        return err
    }

    return cm.client.Set(ctx, key, data, DefaultUserTTL).Err()
}

// Token caching
func (cm *UserCacheManager) StoreToken(ctx context.Context, userID, tokenType, token string) error {
    key := fmt.Sprintf("%s%s:%s", TokenKeyPrefix, userID, tokenType)
    return cm.client.Set(ctx, key, token, DefaultTokenTTL).Err()
}

func (cm *UserCacheManager) GetToken(ctx context.Context, userID, tokenType string) (string, error) {
    key := fmt.Sprintf("%s%s:%s", TokenKeyPrefix, userID, tokenType)
    return cm.client.Get(ctx, key).Result()
}

func (cm *UserCacheManager) InvalidateToken(ctx context.Context, userID, tokenType string) error {
    key := fmt.Sprintf("%s%s:%s", TokenKeyPrefix, userID, tokenType)
    return cm.client.Del(ctx, key).Err()
}

// Session caching
func (cm *UserCacheManager) StoreSession(ctx context.Context, sessionID string, userData map[string]interface{}) error {
    key := fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)
    data, err := json.Marshal(userData)
    if err != nil {
        return err
    }

    return cm.client.Set(ctx, key, data, DefaultSessionTTL).Err()
}

func (cm *UserCacheManager) GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
    key := fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)
    data, err := cm.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, err
    }

    var userData map[string]interface{}
    if err := json.Unmarshal(data, &userData); err != nil {
        return nil, err
    }

    return userData, nil
}

func (cm *UserCacheManager) InvalidateSession(ctx context.Context, sessionID string) error {
    key := fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)
    return cm.client.Del(ctx, key).Err()
}