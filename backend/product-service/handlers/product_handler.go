package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"github.com/louai60/e-commerce_project/backend/product-service/service"
	"go.uber.org/zap"
)

type ProductHandler struct {
	pb.UnimplementedProductServiceServer
	service *service.ProductService
	logger  *zap.Logger
}

func NewProductHandler(service *service.ProductService, logger *zap.Logger) *ProductHandler {
	if service == nil {
		logger.Error("service is nil in NewProductHandler")
		return nil
	}
	if logger == nil {
		// Can't log if logger is nil
		return nil
	}

	logger.Info("Initializing product handler")
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
	if req == nil {
		h.logger.Error("invalid request: request is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Brand == nil {
		h.logger.Error("invalid request: brand is nil")
		return nil, status.Error(codes.InvalidArgument, "brand details are required")
	}

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

// Image methods
func (h *ProductHandler) UploadImage(ctx context.Context, req *pb.UploadImageRequest) (*pb.UploadImageResponse, error) {
	if req == nil {
		h.logger.Error("invalid request: request is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.File) == 0 {
		h.logger.Error("invalid request: file is required")
		return nil, status.Error(codes.InvalidArgument, "file is required")
	}

	h.logger.Info("Uploading image",
		zap.String("folder", req.Folder),
		zap.String("filename", req.Filename))

	return h.service.UploadImage(ctx, req)
}

func (h *ProductHandler) DeleteImage(ctx context.Context, req *pb.DeleteImageRequest) (*pb.DeleteImageResponse, error) {
	if req == nil {
		h.logger.Error("invalid request: request is nil")
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.PublicId == "" {
		h.logger.Error("invalid request: public_id is required")
		return nil, status.Error(codes.InvalidArgument, "public_id is required")
	}

	h.logger.Info("Deleting image", zap.String("public_id", req.PublicId))
	return h.service.DeleteImage(ctx, req)
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
