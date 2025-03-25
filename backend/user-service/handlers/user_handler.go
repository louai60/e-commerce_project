package handlers

import (
	"context"
	"go.uber.org/zap"

	"github.com/louai60/e-commerce_project/backend/user-service/models"
	pb "github.com/louai60/e-commerce_project/backend/user-service/proto"
	"github.com/louai60/e-commerce_project/backend/user-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	service *service.UserService
	logger  *zap.Logger
}

func NewUserHandler(service *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := h.service.GetUser(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return convertUserToProto(user), nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, total, err := h.service.ListUsers(ctx, req.Page, req.Limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	response := &pb.ListUsersResponse{
		Users:      make([]*pb.UserResponse, len(users)),
		Total:      int32(total),
		Page:       req.Page,
		Limit:      req.Limit,
	}

	for i, user := range users {
		response.Users[i] = convertUserToProto(user)
	}

	return response, nil
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	registerReq := &models.RegisterRequest{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	user, err := h.service.CreateUser(ctx, registerReq)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return convertUserToProto(user), nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	user := &models.User{
		ID:        req.Id,
		Email:     req.Email,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
	}

	updatedUser, err := h.service.UpdateUser(ctx, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return convertUserToProto(updatedUser), nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	err := h.service.DeleteUser(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	h.logger.Info("Login attempt", zap.String("email", req.Email))

	credentials := &models.LoginCredentials{
		Email:    req.Email,
		Password: req.Password,
	}

	loginResponse, err := h.service.Login(ctx, credentials)
	if err != nil {
		h.logger.Error("Login failed", 
			zap.String("email", req.Email),
			zap.Error(err))
			
		if err.Error() == "invalid credentials" {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		if err.Error() == "too many attempts" {
			return nil, status.Error(codes.ResourceExhausted, "too many login attempts")
		}
		return nil, status.Error(codes.Internal, "login failed")
	}

	return &pb.LoginResponse{
		Token:        loginResponse.Token,
		// RefreshToken: loginResponse.RefreshToken,
		User:         convertUserToProto(loginResponse.User),
	}, nil
}

func (h *UserHandler) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	if err := h.service.HealthCheck(ctx); err != nil {
		return &pb.HealthCheckResponse{Status: "unhealthy"}, nil
	}
	return &pb.HealthCheckResponse{Status: "healthy"}, nil
}

func convertUserToProto(user *models.User) *pb.UserResponse {
	return &pb.UserResponse{
		Id:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}


