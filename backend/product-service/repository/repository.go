package repository

import (
	"context"
	"database/sql"

	"github.com/louai60/e-commerce_project/backend/product-service/models"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	ListProducts(ctx context.Context, page, limit int32) ([]*models.Product, int64, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
}

// PostgresRepository implements ProductRepository
type PostgresRepository struct {
	db *sql.DB
}

// Ping checks database connectivity
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
