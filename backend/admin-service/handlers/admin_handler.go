package handlers

import (

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	adminpb "github.com/louai60/e-commerce_project/backend/admin-service/proto"
	productpb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	userpb "github.com/louai60/e-commerce_project/backend/user-service/proto"
)

// AdminHandler implements the AdminServiceServer interface.
type AdminHandler struct {
	adminpb.UnimplementedAdminServiceServer // Embed for forward compatibility

	logger        *zap.Logger
	productClient productpb.ProductServiceClient
	userClient    userpb.UserServiceClient
	productConn   *grpc.ClientConn // save connection to close later
	userConn      *grpc.ClientConn
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(logger *zap.Logger, productServiceAddr, userServiceAddr string) (*AdminHandler, error) {
	// Connect to Product Service
	productConn, err := grpc.Dial(productServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to connect to product service", zap.String("address", productServiceAddr), zap.Error(err))
		return nil, err
	}
	productClient := productpb.NewProductServiceClient(productConn)

	// Connect to User Service
	userConn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to connect to user service", zap.String("address", userServiceAddr), zap.Error(err))
		productConn.Close() // Close already-opened product connection
		return nil, err
	}
	userClient := userpb.NewUserServiceClient(userConn)

	return &AdminHandler{
		logger:        logger,
		productClient: productClient,
		userClient:    userClient,
		productConn:   productConn,
		userConn:      userConn,
	}, nil
}

// Close closes the gRPC connections when shutting down
func (h *AdminHandler) Close() {
	if h.productConn != nil {
		h.productConn.Close()
	}
	if h.userConn != nil {
		h.userConn.Close()
	}
}
