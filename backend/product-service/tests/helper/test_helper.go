package helper

import (
    "context"
    "time"

    "github.com/stretchr/testify/suite"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

type TestSuite struct {
    suite.Suite
    ctx     context.Context
    conn    *grpc.ClientConn
    client  pb.ProductServiceClient
    cleanup func()
}

func (s *TestSuite) SetupSuite() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    s.cleanup = cancel
    s.ctx = ctx

    // Connect to the service
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    s.Require().NoError(err)
    s.conn = conn
    s.client = pb.NewProductServiceClient(conn)
}

func (s *TestSuite) TearDownSuite() {
    s.cleanup()
    if s.conn != nil {
        s.conn.Close()
    }
}

// Helper function to create a test product
func (s *TestSuite) CreateTestProduct() *pb.Product {
    product, err := s.client.CreateProduct(s.ctx, &pb.CreateProductRequest{
        Product: &pb.Product{
            Title:       "Test Product",
            Price:       99.99,
            Sku:        "TEST-SKU-" + time.Now().Format("20060102150405"),
            Description: "Test Description",
        },
    })
    s.Require().NoError(err)
    return product
}

// Helper function to create a test brand
func (s *TestSuite) CreateTestBrand() *pb.Brand {
    brand, err := s.client.CreateBrand(s.ctx, &pb.CreateBrandRequest{
        Brand: &pb.Brand{
            Name:        "Test Brand",
            Description: "Test Brand Description",
        },
    })
    s.Require().NoError(err)
    return brand
}

// Helper function to create a test category
func (s *TestSuite) CreateTestCategory() *pb.Category {
    category, err := s.client.CreateCategory(s.ctx, &pb.CreateCategoryRequest{
        Category: &pb.Category{
            Name:        "Test Category",
            Description: "Test Category Description",
        },
    })
    s.Require().NoError(err)
    return category
}