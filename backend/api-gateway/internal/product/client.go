package product

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/louai60/e-commerce_project/backend/api-gateway/config"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

type Client struct {
	client pb.ProductServiceClient
	conn   *grpc.ClientConn
	cfg    *config.Config
}

func NewClient(cfg *config.Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		cfg.Services.Product.Host+":"+cfg.Services.Product.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: pb.NewProductServiceClient(conn),
		conn:   conn,
		cfg:    cfg,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetProduct(ctx context.Context, id string) (*pb.Product, error) {
	req := &pb.GetProductRequest{
		Identifier: &pb.GetProductRequest_Id{
			Id: id,
		},
	}
	return c.client.GetProduct(ctx, req)
}

func (c *Client) ListProducts(ctx context.Context, page, limit int32) (*pb.ListProductsResponse, error) {
	return c.client.ListProducts(ctx, &pb.ListProductsRequest{
		Page:  page,
		Limit: limit,
	})
}

func (c *Client) CreateProduct(ctx context.Context, product *pb.Product) (*pb.Product, error) {
	return c.client.CreateProduct(ctx, &pb.CreateProductRequest{
		Product: product,
	})
}

func (c *Client) UpdateProduct(ctx context.Context, product *pb.Product) (*pb.Product, error) {
	return c.client.UpdateProduct(ctx, &pb.UpdateProductRequest{
		Product: product,
	})
}

func (c *Client) DeleteProduct(ctx context.Context, id string) error {
	_, err := c.client.DeleteProduct(ctx, &pb.DeleteProductRequest{
		Id: id,
	})
	return err
}

func (c *Client) GetBrand(ctx context.Context, id string) (*pb.Brand, error) {
	req := &pb.GetBrandRequest{
		Identifier: &pb.GetBrandRequest_Id{
			Id: id,
		},
	}
	return c.client.GetBrand(ctx, req)
}

func (c *Client) ListBrands(ctx context.Context, page, limit int32) (*pb.ListBrandsResponse, error) {
	return c.client.ListBrands(ctx, &pb.ListBrandsRequest{
		Page:  page,
		Limit: limit,
	})
}

func (c *Client) CreateBrand(ctx context.Context, brand *pb.Brand) (*pb.Brand, error) {
	return c.client.CreateBrand(ctx, &pb.CreateBrandRequest{
		Brand: brand,
	})
}

func (c *Client) GetCategory(ctx context.Context, id string) (*pb.Category, error) {
	req := &pb.GetCategoryRequest{
		Identifier: &pb.GetCategoryRequest_Id{
			Id: id,
		},
	}
	return c.client.GetCategory(ctx, req)
}

func (c *Client) ListCategories(ctx context.Context, page, limit int32) (*pb.ListCategoriesResponse, error) {
	return c.client.ListCategories(ctx, &pb.ListCategoriesRequest{
		Page:  page,
		Limit: limit,
	})
}

func (c *Client) CreateCategory(ctx context.Context, category *pb.Category) (*pb.Category, error) {
	return c.client.CreateCategory(ctx, &pb.CreateCategoryRequest{
		Category: category,
	})
}

func (c *Client) GenerateSKUPreview(ctx context.Context, brandName, categoryName, color, size string) (*pb.GenerateSKUPreviewResponse, error) {
	return c.client.GenerateSKUPreview(ctx, &pb.GenerateSKUPreviewRequest{
		BrandName:    brandName,
		CategoryName: categoryName,
		Color:        color,
		Size:         size,
	})
}
