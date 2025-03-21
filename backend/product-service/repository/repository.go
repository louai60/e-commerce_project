package repository

import (
	"context"

	"github.com/louai60/e-commerce/product-service/models"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	ListProducts(ctx context.Context) ([]*models.Product, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
}