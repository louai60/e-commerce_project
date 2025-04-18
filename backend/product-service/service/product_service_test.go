package service

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/go-redis/redis/v8"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"go.uber.org/zap"

// 	"github.com/louai60/e-commerce_project/backend/product-service/models"
// 	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
// )

// // MockProductRepository implements repository.ProductRepository
// type MockProductRepository struct {
// 	mock.Mock
// }

// func (m *MockProductRepository) CreateProduct(ctx context.Context, product *models.Product) error {
// 	args := m.Called(ctx, product)
// 	return args.Error(0)
// }

// func (m *MockProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
// 	args := m.Called(ctx, id)
// 	if product, ok := args.Get(0).(*models.Product); ok {
// 		return product, args.Error(1)
// 	}
// 	return nil, args.Error(1)
// }

// func (m *MockProductRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
// 	args := m.Called(ctx, slug)
// 	if product, ok := args.Get(0).(*models.Product); ok {
// 		return product, args.Error(1)
// 	}
// 	return nil, args.Error(1)
// }

// func (m *MockProductRepository) ListProducts(ctx context.Context, filters models.ProductFilters) ([]*models.Product, int64, error) {
// 	args := m.Called(ctx, filters)
// 	return args.Get(0).([]*models.Product), args.Get(1).(int64), args.Error(2)
// }

// func (m *MockProductRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
// 	args := m.Called(ctx, product)
// 	return args.Error(0)
// }

// func (m *MockProductRepository) DeleteProduct(ctx context.Context, id string) error {
// 	args := m.Called(ctx, id)
// 	return args.Error(0)
// }

// func (m *MockProductRepository) AddCategories(ctx context.Context, productID string, categoryIDs []string) error {
// 	args := m.Called(ctx, productID, categoryIDs)
// 	return args.Error(0)
// }

// // MockCacheManager implements the cache manager interface
// type MockCacheManager struct {
// 	mock.Mock
// }

// func (m *MockCacheManager) GetProduct(ctx context.Context, id string) (*models.Product, error) {
// 	args := m.Called(ctx, id)
// 	if product, ok := args.Get(0).(*models.Product); ok {
// 		return product, args.Error(1)
// 	}
// 	return nil, args.Error(1)
// }

// func (m *MockCacheManager) SetProduct(ctx context.Context, product *models.Product) error {
// 	args := m.Called(ctx, product)
// 	return args.Error(0)
// }

// func (m *MockCacheManager) InvalidateProduct(ctx context.Context, id string) error {
// 	args := m.Called(ctx, id)
// 	return args.Error(0)
// }

// func (m *MockCacheManager) InvalidateProductLists(ctx context.Context) error {
// 	args := m.Called(ctx)
// 	return args.Error(0)
// }
// func (m *MockProductRepository) AddImage(ctx context.Context, productID string, imageURL string) error {
//     args := m.Called(ctx, productID, imageURL)
//     return args.Error(0)
// }

// func TestGetProduct(t *testing.T) {
// 	mockRepo := new(MockProductRepository)
// 	mockCache := new(MockCacheManager)
// 	logger := zap.NewNop()
// 	service := NewProductService(mockRepo, mockRepo, mockRepo, mockCache, logger)
// 	ctx := context.Background()

// 	expectedProduct := &models.Product{
// 		ID:          "123",
// 		Title:       "Test Product",
// 		Description: "Test Description",
// 		Price:       99.99,
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 	}

// 	// Test cache hit
// 	mockCache.On("GetProduct", ctx, "123").Return(expectedProduct, nil)
	
// 	req := &pb.GetProductRequest{
// 		Identifier: &pb.GetProductRequest_Id{Id: "123"},
// 	}
	
// 	product, err := service.GetProduct(ctx, req)
	
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedProduct.ID, product.Id)
// 	assert.Equal(t, expectedProduct.Title, product.Title)
// 	mockCache.AssertExpectations(t)

// 	// Test cache miss
// 	mockCache.On("GetProduct", ctx, "456").Return(nil, redis.Nil)
// 	mockRepo.On("GetByID", ctx, "456").Return(expectedProduct, nil)
// 	mockCache.On("SetProduct", ctx, expectedProduct).Return(nil)

	
// 	req = &pb.GetProductRequest{
// 		Identifier: &pb.GetProductRequest_Id{Id: "456"},
// 	}
	
// 	product, err = service.GetProduct(ctx, req)
	
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedProduct.ID, product.Id)
// 	mockRepo.AssertExpectations(t)
// 	mockCache.AssertExpectations(t)
// }


