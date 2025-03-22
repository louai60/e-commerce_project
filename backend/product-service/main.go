package main

import (
	"fmt"
	"log"
	"net"

	"github.com/louai60/e-commerce_project/backend/product-service/handlers"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
	"github.com/louai60/e-commerce_project/backend/product-service/service"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"google.golang.org/grpc"
)

func main() {
	// Initialize repository
	repo := repository.NewMemoryRepository()

	// Initialize service
	productService := service.NewProductService(repo)

	// Initialize handler
	productHandler := handlers.NewProductHandler(productService)

	// Set up gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, productHandler)

	fmt.Println("Product service initialized")
	fmt.Println("Server listening on :50051")

	// Start the server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}