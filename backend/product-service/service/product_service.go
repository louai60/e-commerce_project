package service

import (
	"context"
	"fmt"

	"github.com/louai60/e-commerce_project/backend/product-service/cache"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
	"go.uber.org/zap"
)

// ProductService handles business logic for products
type ProductService struct {
	repo         repository.ProductRepository
	cacheManager *cache.CacheManager
	logger       *zap.Logger
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepository, cacheManager *cache.CacheManager, logger *zap.Logger) *ProductService {
	return &ProductService{
		repo:         repo,
		cacheManager: cacheManager,
		logger:       logger,
	}
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	// Try to get from cache first
	product, err := s.cacheManager.GetProduct(ctx, id)
	if err == nil {
		s.logger.Debug("Cache hit for product", zap.String("id", id))
		return product, nil
	}

	// Cache miss, get from database
	product, err = s.repo.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache for future requests
	if err := s.cacheManager.SetProduct(ctx, product); err != nil {
		s.logger.Warn("Failed to cache product", zap.Error(err))
	}

	return product, nil
}

// ListProducts returns all products
func (s *ProductService) ListProducts(ctx context.Context, page, limit int32) ([]*models.Product, int64, error) {
	// Generate cache key from pagination parameters
	filterKey := fmt.Sprintf("page:%d:limit:%d", page, limit)
	
	// Try to get from cache first
	products, err := s.cacheManager.GetProductList(ctx, filterKey)
	if err == nil {
		s.logger.Debug("Cache hit for product list", zap.String("filters", filterKey))
		return products, int64(len(products)), nil
	}

	// Cache miss, get from database
	products, total, err := s.repo.ListProducts(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Store in cache for future requests
	if err := s.cacheManager.SetProductList(ctx, filterKey, products); err != nil {
		s.logger.Warn("Failed to cache product list", zap.Error(err))
	}

	return products, total, nil
}

// CreateProduct adds a new product
func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	if err := s.repo.CreateProduct(ctx, product); err != nil {
		return err
	}

	// Invalidate product lists cache as the list has changed
	if err := s.cacheManager.InvalidateProductLists(ctx); err != nil {
		s.logger.Warn("Failed to invalidate product lists cache", zap.Error(err))
	}

	return nil
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		return err
	}

	// Invalidate both the specific product and product lists cache
	if err := s.cacheManager.InvalidateProduct(ctx, product.ID); err != nil {
		s.logger.Warn("Failed to invalidate product cache", zap.Error(err))
	}
	if err := s.cacheManager.InvalidateProductLists(ctx); err != nil {
		s.logger.Warn("Failed to invalidate product lists cache", zap.Error(err))
	}

	return nil
}

// DeleteProduct removes a product
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		return err
	}

	// Invalidate both the specific product and product lists cache
	if err := s.cacheManager.InvalidateProduct(ctx, id); err != nil {
		s.logger.Warn("Failed to invalidate product cache", zap.Error(err))
	}
	if err := s.cacheManager.InvalidateProductLists(ctx); err != nil {
		s.logger.Warn("Failed to invalidate product lists cache", zap.Error(err))
	}

	return nil
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
