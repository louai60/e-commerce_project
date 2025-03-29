package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	user, err := h.service.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.UserResponse{
		User: convertUserToProto(user),
	}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
    users, total, err := h.service.ListUsers(ctx, req.Page, req.Limit, map[string]interface{}{})
    
    if err != nil {
        h.logger.Error("Failed to list users",
            zap.Int32("page", req.Page),
            zap.Int32("limit", req.Limit),
            zap.Error(err))
        return nil, status.Error(codes.Internal, "failed to list users")
    }

    response := &pb.ListUsersResponse{
        Users:      make([]*pb.User, len(users)),
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
	h.logger.Info("Received create user request",
		zap.String("email", req.Email),
		zap.String("username", req.Username))

	registerReq := &models.RegisterRequest{
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		UserType:    req.UserType,
		Role:        req.Role,
		PhoneNumber: req.PhoneNumber,
	}

	user, err := h.service.CreateUser(ctx, registerReq)
	if err != nil {
		h.logger.Error("Failed to create user", 
			zap.String("email", req.Email),
			zap.String("username", req.Username),
			zap.Error(err))
		
		switch {
		case strings.Contains(err.Error(), "email already registered"):
			return nil, status.Errorf(codes.AlreadyExists, "email already registered")
		case strings.Contains(err.Error(), "username already taken"):
			return nil, status.Errorf(codes.AlreadyExists, "username already taken")
		default:
			return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
		}
	}

	return &pb.UserResponse{
		User: convertUserToProto(user),
	}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	user := &models.User{
		UserID:    req.UserId,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	updatedUser, err := h.service.UpdateUser(ctx, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return &pb.UserResponse{
		User: convertUserToProto(updatedUser),
	}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteResponse, error) {
	err := h.service.DeleteUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "user deleted successfully",
	}, nil
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
			
		switch err.Error() {
		case "invalid credentials":
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		default:
			return nil, status.Error(codes.Internal, "login failed")
		}
	}

	return &pb.LoginResponse{
		Token:        loginResponse.Token,
		RefreshToken: loginResponse.RefreshToken,
		User:         convertUserToProto(loginResponse.User),
	}, nil
}

func (h *UserHandler) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	if err := h.service.HealthCheck(ctx); err != nil {
		return &pb.HealthCheckResponse{Status: "unhealthy"}, nil
	}
	return &pb.HealthCheckResponse{Status: "healthy"}, nil
}

func (h *UserHandler) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.UserResponse, error) {
	user, err := h.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("user not found: %v", err))
	}
	return &pb.UserResponse{
		User: convertUserToProto(user),
	}, nil
}

func convertUserToProto(user *models.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		UserId:        user.UserID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		PhoneNumber:   user.PhoneNumber,
		UserType:      user.UserType,
		Role:          user.Role,
		AccountStatus: user.AccountStatus,
		EmailVerified: user.EmailVerified,
		PhoneVerified: user.PhoneVerified,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     user.UpdatedAt.Format(time.RFC3339),
		LastLogin:     user.LastLogin.Format(time.RFC3339),
	}
}