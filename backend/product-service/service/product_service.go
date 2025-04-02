package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
	"github.com/louai60/e-commerce_project/backend/shared/cache"
)

// ProductService handles business logic for products
type ProductService struct {
	repo    repository.ProductRepository
	cache   *cache.CacheManager
	logger  *zap.Logger
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepository, logger *zap.Logger, cacheManager *cache.CacheManager) *ProductService {
	return &ProductService{
		repo:    repo,
		cache:   cacheManager,
		logger:  logger,
	}
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product:%s", id)
	
	// Try cache first
	var product *models.Product
	err := s.cache.Get(ctx, cacheKey, &product)
	if err == nil {
		return product, nil
	}

	// Cache miss, get from database
	product, err = s.repo.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.Set(ctx, cacheKey, product, 1*time.Hour); err != nil {
		s.logger.Warn("Failed to cache product", zap.Error(err))
	}
	
	return product, nil
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
	err := s.repo.UpdateProduct(ctx, product)
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("product:%s", product.ID)
	s.cache.Delete(ctx, cacheKey)
	
	return nil
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
