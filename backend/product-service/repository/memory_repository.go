package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/louai60/e-commerce/product-service/models"
)

// MemoryRepository implements ProductRepository with an in-memory store
type MemoryRepository struct {
	products map[string]*models.Product
	mutex    sync.RWMutex
}

// NewMemoryRepository creates a new in-memory product repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		products: make(map[string]*models.Product),
	}
}

// GetProduct retrieves a product by ID
func (r *MemoryRepository) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	product, exists := r.products[id]
	if !exists {
		return nil, errors.New("product not found")
	}
	return product, nil
}

// ListProducts returns all products
func (r *MemoryRepository) ListProducts(ctx context.Context) ([]*models.Product, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	products := make([]*models.Product, 0, len(r.products))
	for _, product := range r.products {
		products = append(products, product)
	}
	return products, nil
}

// CreateProduct adds a new product
func (r *MemoryRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.products[product.ID]; exists {
		return errors.New("product already exists")
	}

	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	r.products[product.ID] = product
	return nil
}

// UpdateProduct updates an existing product
func (r *MemoryRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.products[product.ID]; !exists {
		return errors.New("product not found")
	}

	product.UpdatedAt = time.Now()
	r.products[product.ID] = product
	return nil
}

// DeleteProduct removes a product
func (r *MemoryRepository) DeleteProduct(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.products[id]; !exists {
		return errors.New("product not found")
	}

	delete(r.products, id)
	return nil
}