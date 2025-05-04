package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/louai60/e-commerce_project/backend/user-service/models"
)

type Repository interface {
	// User operations
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, page, limit int, where string, args ...interface{}) ([]*models.User, error)
	CountUsers(ctx context.Context, where string, args ...interface{}) (int64, error)
	UpdateRefreshTokenID(ctx context.Context, userID uuid.UUID, refreshTokenID string) error

	// Address operations
	CreateAddress(ctx context.Context, address *models.UserAddress) error
	GetAddresses(ctx context.Context, userID uuid.UUID) ([]models.UserAddress, error)
	UpdateAddress(ctx context.Context, address *models.UserAddress) error
	DeleteAddress(ctx context.Context, addressID uuid.UUID, userID uuid.UUID) error // Assuming addressID should also be UUID
	GetDefaultAddress(ctx context.Context, userID uuid.UUID) (*models.UserAddress, error)

	// Payment method operations
	CreatePaymentMethod(ctx context.Context, payment *models.PaymentMethod) error
	GetPaymentMethods(ctx context.Context, userID uuid.UUID) ([]models.PaymentMethod, error)
	UpdatePaymentMethod(ctx context.Context, payment *models.PaymentMethod) error
	DeletePaymentMethod(ctx context.Context, paymentID uuid.UUID, userID uuid.UUID) error // Assuming paymentID should also be UUID
	GetDefaultPaymentMethod(ctx context.Context, userID uuid.UUID) (*models.PaymentMethod, error)

	// Preferences operations
	CreatePreferences(ctx context.Context, prefs *models.UserPreferences) error
	GetPreferences(ctx context.Context, userID uuid.UUID) (*models.UserPreferences, error)
	UpdatePreferences(ctx context.Context, prefs *models.UserPreferences) error

	// Database health check
	Ping(ctx context.Context) error
}
