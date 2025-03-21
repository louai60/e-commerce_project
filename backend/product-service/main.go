package main

import (
	"fmt"
	"log"
	"net"

	"github.com/louai60/e-commerce/product-service/repository"
	"github.com/louai60/e-commerce/product-service/service"
)

func main() {
	// Initialize repository
	repo := repository.NewMemoryRepository()

	// Initialize service
	productService := service.NewProductService(repo)

	// For now, just print a message
	fmt.Println("Product service initialized")
	fmt.Println("Repository:", repo)
	fmt.Println("Service:", productService)

	// In the future, we'll set up gRPC server here
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	fmt.Println("Server listening on :50051")
	
	// Placeholder for gRPC server
	// We'll implement this in the next step
	
	if err := lis.Close(); err != nil {
		log.Fatalf("Failed to close listener: %v", err)
	}
}