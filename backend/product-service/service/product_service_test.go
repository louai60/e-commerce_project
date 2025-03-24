package service

import (
    "context"
    "testing"
    "time"

    "github.com/louai60/e-commerce_project/backend/product-service/models"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of ProductRepository
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) GetProduct(ctx context.Context, id string) (*models.Product, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockRepository) ListProducts(ctx context.Context) ([]*models.Product, error) {
    args := m.Called(ctx)
    return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockRepository) CreateProduct(ctx context.Context, product *models.Product) error {
    args := m.Called(ctx, product)
    return args.Error(0)
}

func (m *MockRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
    args := m.Called(ctx, product)
    return args.Error(0)
}

func (m *MockRepository) DeleteProduct(ctx context.Context, id string) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

// Add to MockRepository struct
func (m *MockRepository) Ping(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func TestGetProduct(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewProductService(mockRepo)
    ctx := context.Background()

    expectedProduct := &models.Product{
        ID:          "123",
        Name:        "Test Product",
        Description: "Test Description",
        Price:       99.99,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    mockRepo.On("GetProduct", ctx, "123").Return(expectedProduct, nil)

    product, err := service.GetProduct(ctx, "123")

    assert.NoError(t, err)
    assert.Equal(t, expectedProduct, product)
    mockRepo.AssertExpectations(t)
}