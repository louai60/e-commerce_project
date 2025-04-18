package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"time"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// 	"google.golang.org/protobuf/types/known/wrapperspb"

// 	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
// )

// func main() {
// 	// Set up a connection to the server
// 	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("Failed to connect: %v", err)
// 	}
// 	defer conn.Close()

// 	// Create a client
// 	client := pb.NewProductServiceClient(conn)

// 	// Create a context with timeout
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	// Create a timestamp for unique values
// 	timestamp := time.Now().Unix()

// 	// Create a product with images
// 	product := &pb.Product{
// 		Title:            fmt.Sprintf("Test Product with Images %d", timestamp),
// 		Slug:             fmt.Sprintf("test-product-images-%d", timestamp),
// 		Description:      "This is a test product with images",
// 		ShortDescription: "Test product with images",
// 		IsPublished:      true,
// 		InventoryStatus:  "in_stock",
// 		Weight: &wrapperspb.DoubleValue{
// 			Value: 1.5,
// 		},
// 		Images: []*pb.ProductImage{
// 			{
// 				Url:      "https://example.com/image1.jpg",
// 				AltText:  "Test Image 1",
// 				Position: 0,
// 			},
// 			{
// 				Url:      "https://example.com/image2.jpg",
// 				AltText:  "Test Image 2",
// 				Position: 1,
// 			},
// 		},
// 		Variants: []*pb.ProductVariant{
// 			{
// 				Title:        "Red Variant",
// 				Sku:          fmt.Sprintf("TEST-SKU-RED-%d", timestamp),
// 				Price:        99.99,
// 				InventoryQty: 50,
// 				Attributes: []*pb.VariantAttributeValue{
// 					{
// 						Name:  "Color",
// 						Value: "Red",
// 					},
// 				},
// 			},
// 			{
// 				Title:        "Blue Variant",
// 				Sku:          fmt.Sprintf("TEST-SKU-BLUE-%d", timestamp),
// 				Price:        99.99,
// 				InventoryQty: 50,
// 				Attributes: []*pb.VariantAttributeValue{
// 					{
// 						Name:  "Color",
// 						Value: "Blue",
// 					},
// 				},
// 			},
// 		},
// 		Tags: []*pb.ProductTag{
// 			{Tag: "test"},
// 			{Tag: "images"},
// 		},
// 		Attributes: []*pb.ProductAttribute{
// 			{
// 				Name:  "Material",
// 				Value: "Aluminum",
// 			},
// 			{
// 				Name:  "Size",
// 				Value: "Medium",
// 			},
// 		},
// 		Specifications: []*pb.ProductSpecification{
// 			{
// 				Name:  "Dimensions",
// 				Value: "10 x 5 x 2",
// 				Unit:  "inches",
// 			},
// 			{
// 				Name:  "Weight",
// 				Value: "1.5",
// 				Unit:  "kg",
// 			},
// 		},
// 	}

// 	// Create the request
// 	req := &pb.CreateProductRequest{
// 		Product: product,
// 	}

// 	// Call the CreateProduct method
// 	resp, err := client.CreateProduct(ctx, req)
// 	if err != nil {
// 		log.Fatalf("Failed to create product: %v", err)
// 	}

// 	log.Printf("Product created successfully with ID: %s", resp.Id)

// 	// Get the product to verify it was created correctly
// 	getReq := &pb.GetProductRequest{
// 		Identifier: &pb.GetProductRequest_Id{
// 			Id: resp.Id,
// 		},
// 	}

// 	getResp, err := client.GetProduct(ctx, getReq)
// 	if err != nil {
// 		log.Fatalf("Failed to get product: %v", err)
// 	}

// 	// Print product details
// 	log.Printf("Retrieved product: %s", getResp.Product.Title)
// 	log.Printf("Number of images: %d", len(getResp.Product.Images))
// 	for i, img := range getResp.Product.Images {
// 		log.Printf("Image %d: URL=%s, AltText=%s", i+1, img.Url, img.AltText)
// 	}
// 	log.Printf("Number of variants: %d", len(getResp.Product.Variants))
// 	for i, variant := range getResp.Product.Variants {
// 		log.Printf("Variant %d: SKU=%s, Title=%s", i+1, variant.Sku, variant.Title)
// 	}
// 	log.Printf("Number of attributes: %d", len(getResp.Product.Attributes))
// 	log.Printf("Number of specifications: %d", len(getResp.Product.Specifications))
// 	log.Printf("Number of tags: %d", len(getResp.Product.Tags))

// 	// Now check the database to verify the images were saved
// 	log.Printf("To verify images in the database, run: go run tests/check_product_images_in_db.go")
// }
