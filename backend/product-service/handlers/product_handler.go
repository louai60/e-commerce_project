package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"github.com/louai60/e-commerce_project/backend/product-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
    "go.uber.org/zap"
    "github.com/louai60/e-commerce_project/backend/common/logger"
)

// ProductHandler handles gRPC requests for products
type ProductHandler struct {
	pb.UnimplementedProductServiceServer
	productService service.ProductServiceInterface
	log            *zap.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService service.ProductServiceInterface) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		log:            logger.GetLogger(),
	}
}

// GetProduct retrieves a product by ID
func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
	if req.Id == "" {
		h.log.Warn("Empty product ID in request")
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	product, err := h.productService.GetProduct(ctx, req.Id)
	if err != nil {
		h.log.Error("Failed to get product", zap.String("product_id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "failed to get product: %v", err)
	}

	return convertProductToProto(product), nil
}

// ListProducts returns all products
func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, err := h.productService.ListProducts(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	response := &pb.ListProductsResponse{
		Products: make([]*pb.ProductResponse, 0, len(products)),
	}

	for _, product := range products {
		response.Products = append(response.Products, convertProductToProto(product))
	}

	return response, nil
}

// CreateProduct adds a new product
func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	if req.Name == "" || req.Price <= 0 {
		h.log.Warn("Invalid create product request", zap.Any("request", req))
		return nil, status.Error(codes.InvalidArgument, "name and positive price are required")
	}

	product := &models.Product{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageUrl,
		CategoryID:  req.CategoryId,
		Stock:       int(req.Stock),
	}

	err := h.productService.CreateProduct(ctx, product)
	if err != nil {
		h.log.Error("Failed to create product", zap.Any("request", req), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	return convertProductToProto(product), nil
}

// UpdateProduct updates an existing product
func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductResponse, error) {
	product, err := h.productService.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	// Update fields
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.ImageURL = req.ImageUrl
	product.CategoryID = req.CategoryId
	product.Stock = int(req.Stock)

	err = h.productService.UpdateProduct(ctx, product)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	return convertProductToProto(product), nil
}

// DeleteProduct removes a product
func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	err := h.productService.DeleteProduct(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	return &pb.DeleteProductResponse{Success: true}, nil
}

// Helper function to convert a product model to a proto response
func convertProductToProto(product *models.Product) *pb.ProductResponse {
	return &pb.ProductResponse{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		ImageUrl:    product.ImageURL,
		CategoryId:  product.CategoryID,
		Stock:       int32(product.Stock),
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}
}
