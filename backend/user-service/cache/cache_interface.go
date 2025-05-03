package cache

import (
	"context"

	"github.com/louai60/e-commerce_project/backend/user-service/models"
)

// CacheInterface defines the interface for user cache managers
type CacheInterface interface {
	// User methods
	GetUser(ctx context.Context, userID string) (*models.User, error)
	SetUser(ctx context.Context, user *models.User) error

	// Token methods
	StoreToken(ctx context.Context, userID, tokenType, token string) error
	GetToken(ctx context.Context, userID, tokenType string) (string, error)
	InvalidateToken(ctx context.Context, userID, tokenType string) error

	// Session methods
	StoreSession(ctx context.Context, sessionID string, userData map[string]interface{}) error
	GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error)
	InvalidateSession(ctx context.Context, sessionID string) error

	// Close connection
	Close() error
}
