package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	adminpb "github.com/louai60/e-commerce_project/backend/admin-service/proto"
	"github.com/louai60/e-commerce_project/backend/admin-service/handlers"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, using environment variables")
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	// Get service addresses and port from environment variables
	productServiceAddr := os.Getenv("PRODUCT_SERVICE_ADDR")
	if productServiceAddr == "" {
		logger.Fatal("PRODUCT_SERVICE_ADDR environment variable is required")
	}
	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		logger.Fatal("USER_SERVICE_ADDR environment variable is required")
	}
	port := os.Getenv("ADMIN_SERVICE_PORT")
	if port == "" {
		port = "8085" // Default port
		logger.Warn("ADMIN_SERVICE_PORT not set, using default", zap.String("port", port))
	}

	// Set up TCP listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err), zap.String("port", port))
	}

	// Create a new gRPC server
	s := grpc.NewServer()

	// Create and register the admin handler
	adminHandler, err := handlers.NewAdminHandler(logger, productServiceAddr, userServiceAddr)
	if err != nil {
		logger.Fatal("Failed to create admin handler", zap.Error(err))
	}
	adminpb.RegisterAdminServiceServer(s, adminHandler)

	// Set up channel for graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a separate goroutine
	go func() {
		logger.Info("Admin gRPC server listening", zap.String("address", lis.Addr().String()))
		if err := s.Serve(lis); err != nil {
			logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	// Wait for termination signal (Ctrl+C or kill)
	sigReceived := <-stopChan
	logger.Info("Received signal", zap.String("signal", sigReceived.String()))

	// Graceful shutdown
	logger.Info("Shutting down the server...")
	s.GracefulStop()

	// Close the admin handler gRPC connections
	adminHandler.Close()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)

	logger.Info("Server gracefully shut down")
}
