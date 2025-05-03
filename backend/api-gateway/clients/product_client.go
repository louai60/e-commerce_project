package clients

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/louai60/e-commerce_project/backend/api-gateway/config"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

// ProductInfo represents basic product information
type ProductInfo struct {
	Id     string      `json:"id"`
	Title  string      `json:"title"`
	Slug   string      `json:"slug"`
	Images []ImageInfo `json:"images,omitempty"`
}

// ImageInfo represents basic image information
type ImageInfo struct {
	Url string `json:"url"`
}

// ProductClient handles communication with the product service
type ProductClient struct {
	client pb.ProductServiceClient
	conn   *grpc.ClientConn
	logger *zap.Logger
}

// NewProductClient creates a new product service client
func NewProductClient(cfg *config.Config, logger *zap.Logger) (*ProductClient, error) {
	// Get product service address from config
	productAddr := fmt.Sprintf("%s:%s", cfg.Services.Product.Host, cfg.Services.Product.Port)
	logger.Info("Connecting to product service", zap.String("address", productAddr))

	// Set up connection with retry logic
	var conn *grpc.ClientConn
	var err error
	maxRetries := 3
	retryInterval := time.Second * 2

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		conn, err = grpc.DialContext(
			ctx,
			productAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		cancel()

		if err == nil {
			break
		}

		logger.Warn("Failed to connect to product service, retrying...",
			zap.String("address", productAddr),
			zap.Error(err),
			zap.Int("attempt", i+1),
			zap.Int("maxRetries", maxRetries))

		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	if err != nil {
		logger.Error("Failed to connect to product service after retries",
			zap.String("address", productAddr),
			zap.Error(err))
		return nil, fmt.Errorf("failed to connect to product service: %w", err)
	}

	return &ProductClient{
		client: pb.NewProductServiceClient(conn),
		conn:   conn,
		logger: logger,
	}, nil
}

// Close closes the gRPC connection
func (c *ProductClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetProduct retrieves a product by ID
func (c *ProductClient) GetProduct(ctx context.Context, id string) (*pb.Product, error) {
	c.logger.Info("Getting product by ID", zap.String("id", id))

	// Create the request
	req := &pb.GetProductRequest{
		Identifier: &pb.GetProductRequest_Id{
			Id: id,
		},
	}

	// Call the product service
	resp, err := c.client.GetProduct(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get product", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return resp, nil
}

// GetProducts retrieves a list of products
func (c *ProductClient) GetProducts(ctx context.Context, page, limit int, filters map[string]string) (*pb.ListProductsResponse, error) {
	c.logger.Info("Getting products", zap.Int("page", page), zap.Int("limit", limit))

	// Create the request
	req := &pb.ListProductsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	}

	// Call the product service
	resp, err := c.client.ListProducts(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get products", zap.Error(err))
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	return resp, nil
}
