package service

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
)

// ProductService handles business logic for products
type ProductService struct {
	repo   repository.ProductRepository
	logger *zap.Logger
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepository, logger *zap.Logger) *ProductService {
	return &ProductService{
		repo:   repo,
		logger: logger,
	}
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	s.logger.Info("Getting product", zap.String("id", id))
	return s.repo.GetProduct(ctx, id)
}

// ListProducts returns all products
func (s *ProductService) ListProducts(ctx context.Context, page, limit int32) ([]*models.Product, int64, error) {
	s.logger.Info("Listing products", zap.Int32("page", page), zap.Int32("limit", limit))
	return s.repo.ListProducts(ctx, page, limit)
}

// CreateProduct adds a new product
func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	product.ID = uuid.New().String()
	s.logger.Info("Creating product", 
		zap.String("id", product.ID),
		zap.String("name", product.Name),
	)
	return s.repo.CreateProduct(ctx, product)
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
	s.logger.Info("Updating product", 
		zap.String("id", product.ID),
		zap.String("name", product.Name),
	)
	return s.repo.UpdateProduct(ctx, product)
}

// DeleteProduct removes a product
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	s.logger.Info("Deleting product", zap.String("id", id))
	return s.repo.DeleteProduct(ctx, id)
}

// HealthCheck verifies the service is healthy by checking database connectivity
func (s *ProductService) HealthCheck(ctx context.Context) error {
	// Perform a simple database ping or lightweight query
return nil // TODO: Implement proper health check mechanism
}

// ProductServiceInterface defines the methods for product service
type ProductServiceInterface interface {
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	ListProducts(ctx context.Context, page, limit int32) ([]*models.Product, int64, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
	HealthCheck(ctx context.Context) error
}

// Ensure ProductService implements ProductServiceInterface
var _ ProductServiceInterface = (*ProductService)(nil)
