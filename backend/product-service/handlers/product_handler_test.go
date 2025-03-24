package handlers

import (
    "context"
    "testing"

    "github.com/louai60/e-commerce_project/backend/product-service/models"
    pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
    "github.com/louai60/e-commerce_project/backend/product-service/service"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockProductService implements ProductServiceInterface
type MockProductService struct {
    mock.Mock
}

var _ service.ProductServiceInterface = (*MockProductService)(nil)

func (m *MockProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductService) ListProducts(ctx context.Context) ([]*models.Product, error) {
    args := m.Called(ctx)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductService) CreateProduct(ctx context.Context, product *models.Product) error {
    args := m.Called(ctx, product)
    return args.Error(0)
}

func (m *MockProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
    args := m.Called(ctx, product)
    return args.Error(0)
}

func (m *MockProductService) DeleteProduct(ctx context.Context, id string) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

// Add to MockProductService struct
func (m *MockProductService) HealthCheck(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func TestCreateProduct(t *testing.T) {
    mockService := new(MockProductService)
    handler := NewProductHandler(mockService)
    ctx := context.Background()

    req := &pb.CreateProductRequest{
        Name:        "Test Product",
        Description: "Test Description",
        Price:       99.99,
        ImageUrl:    "http://example.com/image.jpg",
        CategoryId:  "cat123",
        Stock:       100,
    }

    mockService.On("CreateProduct", ctx, mock.AnythingOfType("*models.Product")).Return(nil)

    response, err := handler.CreateProduct(ctx, req)

    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.Equal(t, req.Name, response.Name)
    assert.Equal(t, req.Description, response.Description)
    assert.Equal(t, req.Price, response.Price)
    mockService.AssertExpectations(t)
}