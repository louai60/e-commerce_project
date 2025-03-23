package service

import (
	"context"

	"github.com/louai60/e-commerce_project/backend/common/logger"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
	"go.uber.org/zap"
)

// ProductService handles business logic for products
type ProductService struct {
	repo repository.ProductRepository
	log  *zap.Logger
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{
		repo: repo,
		log:  logger.GetLogger(),
	}
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	s.log.Info("Getting product", zap.String("id", id))
	
	product, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		s.log.Error("Failed to get product", 
			zap.String("id", id),
			zap.Error(err))
		return nil, err
	}
	
	return product, nil
}

// ListProducts returns all products
func (s *ProductService) ListProducts(ctx context.Context) ([]*models.Product, error) {
	return s.repo.ListProducts(ctx)
}

// CreateProduct adds a new product
func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	s.log.Info("Creating product",
		zap.String("name", product.Name),
		zap.Float64("price", product.Price))
	
	err := s.repo.CreateProduct(ctx, product)
	if err != nil {
		s.log.Error("Failed to create product",
			zap.String("name", product.Name),
			zap.Error(err))
		return err
	}
	
	s.log.Info("Product created successfully", zap.String("id", product.ID))
	return nil
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
	return s.repo.UpdateProduct(ctx, product)
}

// DeleteProduct removes a product
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.DeleteProduct(ctx, id)
}


// ProductServiceInterface defines the methods for product service
type ProductServiceInterface interface {
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	ListProducts(ctx context.Context) ([]*models.Product, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
}

// Ensure ProductService implements ProductServiceInterface
var _ ProductServiceInterface = (*ProductService)(nil)
