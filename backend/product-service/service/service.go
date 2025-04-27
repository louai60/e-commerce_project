package service

import (
	"context"

	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

// ProductService defines the interface for product service operations
type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error)
	GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error)
	GetProductFixed(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error)
	FixProductData(ctx context.Context, productID string) error
	ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error)
	UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error)
	DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error)
	GetBrand(ctx context.Context, req *pb.GetBrandRequest) (*pb.Brand, error)
	ListBrands(ctx context.Context, req *pb.ListBrandsRequest) (*pb.ListBrandsResponse, error)
	CreateBrand(ctx context.Context, req *pb.CreateBrandRequest) (*pb.Brand, error)
	// UpdateBrand(ctx context.Context, req *pb.UpdateBrandRequest) (*pb.Brand, error)
	// DeleteBrand(ctx context.Context, req *pb.DeleteBrandRequest) (*pb.DeleteBrandResponse, error)
	GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.Category, error)
	ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error)
	CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error)
	// UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error)
	// DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error)
}
