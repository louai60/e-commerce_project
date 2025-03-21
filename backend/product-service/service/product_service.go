package service

import (
	"context"

	"github.com/louai60/e-commerce/product-service/models"
	"github.com/louai60/e-commerce/product-service/repository"
)

// ProductService handles business logic for products
type ProductService struct {
	repo repository.ProductRepository
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	return s.repo.GetProduct(ctx, id)
}

// ListProducts returns all products
func (s *ProductService) ListProducts(ctx context.Context) ([]*models.Product, error) {
	return s.repo.ListProducts(ctx)
}

// CreateProduct adds a new product
func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	return s.repo.CreateProduct(ctx, product)
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
	return s.repo.UpdateProduct(ctx, product)
}

// DeleteProduct removes a product
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.DeleteProduct(ctx, id)
}