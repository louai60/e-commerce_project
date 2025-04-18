package handlers

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"github.com/louai60/e-commerce_project/backend/product-service/service"
)

type ProductHandler struct {
	pb.UnimplementedProductServiceServer
	service *service.ProductService
	logger  *zap.Logger
}

func NewProductHandler(service *service.ProductService, logger *zap.Logger) *ProductHandler {
	if service == nil {
		logger.Fatal("product service cannot be nil")
		return nil
	}
	
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
	if h.service == nil {
		h.logger.Error("product service is nil")
		return nil, status.Error(codes.Internal, "service not initialized")
	}

	if req == nil || req.Product == nil {
		h.logger.Error("invalid request: request or product is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	return h.service.CreateProduct(ctx, req)
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	h.logger.Info("Getting product", zap.Any("identifier", req.Identifier))
	return h.service.GetProduct(ctx, req)
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	h.logger.Info("Listing products", 
		zap.Int32("page", req.Page),
		zap.Int32("limit", req.Limit))
	return h.service.ListProducts(ctx, req)
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	if req == nil || req.Product == nil {
		h.logger.Error("invalid request: request or product is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Product.Id == "" {
		h.logger.Error("invalid request: product ID is required")
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	h.logger.Info("Updating product", zap.String("id", req.Product.Id))
	return h.service.UpdateProduct(ctx, req)
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	if req == nil || req.Id == "" {
		h.logger.Error("invalid request: request or product ID is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	h.logger.Info("Deleting product", zap.String("id", req.Id))
	response, err := h.service.DeleteProduct(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Brand methods
func (h *ProductHandler) CreateBrand(ctx context.Context, req *pb.CreateBrandRequest) (*pb.Brand, error) {
	h.logger.Info("Creating brand", zap.String("name", req.Brand.Name))
	return h.service.CreateBrand(ctx, req.Brand)
}

func (h *ProductHandler) GetBrand(ctx context.Context, req *pb.GetBrandRequest) (*pb.Brand, error) {
	if req == nil {
		h.logger.Error("invalid request: request is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Identifier == nil {
		h.logger.Error("invalid request: identifier is required")
		return nil, status.Error(codes.InvalidArgument, "identifier is required")
	}

	h.logger.Info("Getting brand", zap.Any("identifier", req.Identifier))
	return h.service.GetBrand(ctx, req)
}

func (h *ProductHandler) ListBrands(ctx context.Context, req *pb.ListBrandsRequest) (*pb.ListBrandsResponse, error) {
	if req == nil {
		h.logger.Error("invalid request: request is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	h.logger.Info("Listing brands", 
		zap.Int32("page", req.Page),
		zap.Int32("limit", req.Limit))
	return h.service.ListBrands(ctx, req)
}

// Category methods
func (h *ProductHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
	if req == nil || req.Category == nil {
		h.logger.Error("invalid request: request or category is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Category.Name == "" {
		h.logger.Error("invalid request: category name is required")
		return nil, status.Error(codes.InvalidArgument, "category name is required")
	}

	h.logger.Info("Creating category", zap.String("name", req.Category.Name))
	return h.service.CreateCategory(ctx, req)
}

func (h *ProductHandler) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.Category, error) {
	if req == nil {
		h.logger.Error("invalid request: request is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Identifier == nil {
		h.logger.Error("invalid request: identifier is required")
		return nil, status.Error(codes.InvalidArgument, "identifier is required")
	}

	h.logger.Info("Getting category", zap.Any("identifier", req.Identifier))
	return h.service.GetCategory(ctx, req)
}

func (h *ProductHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	if req == nil {
		h.logger.Error("invalid request: request is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	h.logger.Info("Listing categories", 
		zap.Int32("page", req.Page),
		zap.Int32("limit", req.Limit))
	return h.service.ListCategories(ctx, req)
}






