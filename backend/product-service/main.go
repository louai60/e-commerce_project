package main

import (
	"fmt"
	"net"

	"github.com/louai60/e-commerce_project/backend/common/logger"
	"github.com/louai60/e-commerce_project/backend/product-service/handlers"
	"github.com/louai60/e-commerce_project/backend/product-service/repository"
	"github.com/louai60/e-commerce_project/backend/product-service/service"
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"google.golang.org/grpc"
)

func main() {
	// Initialize logger
	logger.Initialize("development")
	log := logger.GetLogger()
	defer log.Sync()

	// Initialize repository
	repo := repository.NewMemoryRepository()

	// Initialize service
	productService := service.NewProductService(repo)

	// Initialize handler
	productHandler := handlers.NewProductHandler(productService)

	// Set up gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("Failed to listen",
			zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, productHandler)

	log.Info("Product service initialized",
		zap.String("port", "50051"))

	// Start the server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve",
			zap.Error(err))
	}
}
