package repository

import (
	"context"

	"github.com/louai60/e-commerce_project/backend/user-service/models"
)

type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, userID int64) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, userID int64) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	ListUsers(ctx context.Context, page, limit int, where string, args ...interface{}) ([]*models.User, error)
	CountUsers(ctx context.Context, where string, args ...interface{}) (int64, error)

	// Address operations
	CreateAddress(ctx context.Context, address *models.UserAddress) error
	GetAddresses(ctx context.Context, userID int64) ([]models.UserAddress, error)
	UpdateAddress(ctx context.Context, address *models.UserAddress) error
	DeleteAddress(ctx context.Context, addressID, userID int64) error
	GetDefaultAddress(ctx context.Context, userID int64) (*models.UserAddress, error)

	// Payment method operations
	CreatePaymentMethod(ctx context.Context, payment *models.PaymentMethod) error
	GetPaymentMethods(ctx context.Context, userID int64) ([]models.PaymentMethod, error)
	UpdatePaymentMethod(ctx context.Context, payment *models.PaymentMethod) error
	DeletePaymentMethod(ctx context.Context, paymentID, userID int64) error
	GetDefaultPaymentMethod(ctx context.Context, userID int64) (*models.PaymentMethod, error)

	// Preferences operations
	CreatePreferences(ctx context.Context, prefs *models.UserPreferences) error
	GetPreferences(ctx context.Context, userID int64) (*models.UserPreferences, error)
	UpdatePreferences(ctx context.Context, prefs *models.UserPreferences) error

	// Database health check
	Ping(ctx context.Context) error
}

