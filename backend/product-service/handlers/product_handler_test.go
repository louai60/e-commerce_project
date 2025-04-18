package handlers

// import (
// 	"context"
// 	"testing"
// 	"errors"

// 	"github.com/louai60/e-commerce_project/backend/product-service/models"
// 	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
// 	"github.com/louai60/e-commerce_project/backend/product-service/service"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"go.uber.org/zap"
// )

// // MockProductService implements ProductServiceInterface
// type MockProductService struct {
// 	mock.Mock
// }

// var _ service.ProductServiceInterface = (*MockProductService)(nil)

// func (m *MockProductService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*models.Product, error) {
// 	args := m.Called(ctx, req)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*models.Product), args.Error(1)
// }

// func (m *MockProductService) ListProducts(ctx context.Context, page, limit int32) ([]*models.Product, int64, error) {
// 	args := m.Called(ctx, page, limit)
// 	if args.Get(0) == nil {
// 		return nil, 0, args.Error(2)
// 	}
// 	return args.Get(0).([]*models.Product), args.Get(1).(int64), args.Error(2)
// }

// func (m *MockProductService) CreateProduct(ctx context.Context, product *models.Product) error {
// 	args := m.Called(ctx, product)
// 	return args.Error(0)
// }

// func (m *MockProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
// 	args := m.Called(ctx, product)
// 	return args.Error(0)
// }

// func (m *MockProductService) DeleteProduct(ctx context.Context, id string) error {
// 	args := m.Called(ctx, id)
// 	return args.Error(0)
// }

// // Add to MockProductService struct
// func (m *MockProductService) HealthCheck(ctx context.Context) error {
// 	args := m.Called(ctx)
// 	return args.Error(0)
// }

// func setupTest() (*ProductHandler, *MockProductService, *zap.Logger) {
// 	logger := zap.NewNop()
// 	mockService := new(MockProductService)
// handler := NewProductHandler(service.ProductServiceInterface(mockService), logger)
// 	return handler, mockService, logger
// }

// func TestCreateProduct(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		req     *pb.CreateProductRequest
// 		setup   func(*MockProductService)
// 		wantErr bool
// 	}{
// 		{
// 			name: "successful creation",
// 			req: &pb.CreateProductRequest{
// 				Product: &pb.Product{
// 					Title:            "Test Product",
// 					Description:      "Test Description",
// 					Price:           99.99,
// 					Sku:             "SKU123",
// 					InventoryQty:    100,
// 					IsPublished:     true,
// 				},
// 			},
// 			setup: func(ms *MockProductService) {
// 				ms.On("CreateProduct", mock.Anything, mock.AnythingOfType("*pb.CreateProductRequest")).Return(&pb.Product{}, nil)
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "service error",
// 			req: &pb.CreateProductRequest{
// 				Product: &pb.Product{
// 					Title: "Test Product",
// 				},
// 			},
// 			setup: func(ms *MockProductService) {
// 				ms.On("CreateProduct", mock.Anything, mock.AnythingOfType("*pb.CreateProductRequest")).Return(nil, errors.New("service error"))
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "invalid request - missing title",
// 			req: &pb.CreateProductRequest{
// 				Product: &pb.Product{
// 					Description: "Test Description",
// 				},
// 			},
// 			setup: func(ms *MockProductService) {
// 				// No setup needed as validation should fail before service call
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			handler, mockService, _ := setupTest()
// 			if tt.setup != nil {
// 				tt.setup(mockService)
// 			}

// 			resp, err := handler.CreateProduct(context.Background(), tt.req)

// 			if tt.wantErr {
// 				assert.Error(t, err)
// 				assert.Nil(t, resp)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, resp)
// 				mockService.AssertExpectations(t)
// 			}
// 		})
// 	}
// }

// func TestGetProduct(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		id      string
// 		setup   func(*MockProductService)
// 		wantErr bool
// 	}{
// 		{
// 			name: "successful retrieval",
// 			id:   "valid-id",
// 			setup: func(ms *MockProductService) {
// 				ms.On("GetProduct", mock.Anything, mock.AnythingOfType("*pb.GetProductRequest")).Return(&models.Product{
// 					Title: "Test Product",
// 					Price: 99.99,
// 				}, nil)
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "not found",
// 			id:   "invalid-id",
// 			setup: func(ms *MockProductService) {
// 				ms.On("GetProduct", mock.Anything, mock.AnythingOfType("*pb.GetProductRequest")).Return(nil, errors.New("not found"))
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			handler, mockService, _ := setupTest()
// 			if tt.setup != nil {
// 				tt.setup(mockService)
// 			}

// 			req := &pb.GetProductRequest{
// 				Identifier: &pb.GetProductRequest_Id{Id: tt.id},
// 			}
// 			resp, err := handler.GetProduct(context.Background(), req)

// 			if tt.wantErr {
// 				assert.Error(t, err)
// 				assert.Nil(t, resp)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, resp)
// 				mockService.AssertExpectations(t)
// 			}
// 		})
// 	}
// }

