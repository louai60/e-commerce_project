package repository

import (
	"context"

	"github.com/louai60/e-commerce_project/backend/user-service/models"
)

type UserRepository interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	ListUsers(ctx context.Context, page, limit int32) ([]*models.User, int64, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	Ping(ctx context.Context) error
}

// type PostgresRepository struct {
// 	db *sql.DB
// }

// Remove duplicate NewPostgresRepository function declaration
// func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
// 	if db == nil {
// 		return nil, fmt.Errorf("database connection is nil")
// 	}
// 	return &PostgresRepository{db: db}, nil
// }
