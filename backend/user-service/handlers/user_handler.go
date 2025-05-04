package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/louai60/e-commerce_project/backend/user-service/models"
	pb "github.com/louai60/e-commerce_project/backend/user-service/proto"
	"github.com/louai60/e-commerce_project/backend/user-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	service      *service.UserService
	logger       *zap.Logger
	tokenManager *service.JWTManager
}

func NewUserHandler(service *service.UserService, logger *zap.Logger, tokenManager *service.JWTManager) *UserHandler {
	return &UserHandler{
		service:      service,
		logger:       logger,
		tokenManager: tokenManager,
	}
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.Error("Invalid UUID format", zap.String("userID", req.UserId), zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		h.logger.Warn("User not found", zap.String("userID", userID.String()), zap.Error(err))
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.UserResponse{
		User: convertUserToProto(user),
	}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, total, err := h.service.ListUsers(ctx, req.Page, req.Limit, map[string]any{})

	if err != nil {
		h.logger.Error("Failed to list users",
			zap.Int32("page", req.Page),
			zap.Int32("limit", req.Limit),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	response := &pb.ListUsersResponse{
		Users: make([]*pb.User, len(users)),
		Total: int32(total),
		Page:  req.Page,
		Limit: req.Limit,
	}

	for i, user := range users {
		response.Users[i] = convertUserToProto(user)
	}

	return response, nil
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	h.logger.Info("Received create user request",
		zap.String("email", req.Email),
		zap.String("firstName", req.FirstName),
		zap.String("lastName", req.LastName),
		zap.String("userType", req.UserType), // Add logging for userType
		zap.String("role", req.Role))         // Add logging for role

	// Map the gRPC request to the service layer's RegisterRequest model
	registerReq := &models.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		UserType:  req.UserType, // Map UserType from request
		Role:      req.Role,     // Map Role from request
		// Username and PhoneNumber are still omitted
	}

	user, err := h.service.CreateUser(ctx, registerReq)
	if err != nil {
		h.logger.Error("Failed to create user",
			zap.String("email", req.Email),
			zap.Error(err))

		// Check for specific errors returned by the service/repository
		// The repository now handles username uniqueness based on email.
		if status.Code(err) == codes.AlreadyExists || strings.Contains(err.Error(), "already exists") {
			// Return a generic "already exists" error, as it could be email or the derived username
			return nil, status.Errorf(codes.AlreadyExists, "email or username already exists")
		}
		// Handle other potential errors (e.g., validation errors if added later)
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err.Error())
	}

	return &pb.UserResponse{
		User: convertUserToProto(user),
	}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.Error("Invalid UUID format", zap.String("userID", req.UserId), zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	user := &models.User{
		UserID:    userID,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	updatedUser, err := h.service.UpdateUser(ctx, user)
	if err != nil {
		h.logger.Error("Failed to update user",
			zap.String("userID", userID.String()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return &pb.UserResponse{
		User: convertUserToProto(updatedUser),
	}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.Error("Invalid UUID format", zap.String("userID", req.UserId), zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	err = h.service.DeleteUser(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.String("userID", userID.String()), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "user deleted successfully",
	}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	h.logger.Info("Login attempt", zap.String("email", req.Email))

	user, err := h.service.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Login failed",
			zap.String("email", req.Email),
			zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	accessToken, refreshToken, refreshTokenID, cookie, err := h.tokenManager.GenerateTokenPair(user)
	if err != nil {
		h.logger.Error("Failed to generate token pair", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	err = h.service.UpdateRefreshTokenID(ctx, user.UserID, refreshTokenID)
	if err != nil {
		h.logger.Error("Failed to store refresh token ID",
			zap.String("userID", user.UserID.String()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to store refresh token ID")
	}

	return &pb.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         convertUserToProto(user),
		Cookie: &pb.CookieInfo{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HttpOnly,
			MaxAge:   int32(cookie.MaxAge),
		},
	}, nil
}

func (h *UserHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	h.logger.Info("Attempting to refresh token")

	// Extract claims from the refresh token
	claims := jwt.MapClaims{}
	publicKey, err := h.tokenManager.GetPublicKey()
	if err != nil {
		h.logger.Error("Failed to get public key for token validation", zap.Error(err))
		return nil, status.Error(codes.Internal, "token validation setup error")
	}

	token, err := jwt.ParseWithClaims(req.RefreshToken, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}
		return publicKey, nil
	})

	// Explicit expiration check
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				h.logger.Warn("Refresh token has expired")
				return nil, status.Error(codes.Unauthenticated, "refresh token has expired")
			}
		}
		h.logger.Error("Invalid refresh token", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	if !token.Valid {
		h.logger.Error("Token validation failed")
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	// Double-check expiration explicitly
	if exp, ok := claims["exp"].(float64); !ok || float64(time.Now().Unix()) > exp {
		h.logger.Warn("Refresh token has expired (manual check)")
		return nil, status.Error(codes.Unauthenticated, "refresh token has expired")
	}

	// Extract user_id claim as string and parse to UUID
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		h.logger.Error("user_id claim is not a string or is missing", zap.Any("claim", claims["user_id"]))
		return nil, status.Error(codes.InvalidArgument, "user_id claim is not a string or is missing")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.Error("Invalid user_id format in token", zap.String("userIDStr", userIDStr), zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id format in token: %v", err)
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid jti in token")
	}

	// Validate the JTI against the stored JTI
	valid, err := h.service.ValidateRefreshTokenID(ctx, userID, jti) // Use parsed uuid.UUID
	if err != nil || !valid {
		h.logger.Error("Invalid refresh token ID", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token ID")
	}

	// Get the user by the ID from the token
	user, err := h.service.GetUser(ctx, userID) // Use parsed uuid.UUID
	if err != nil {
		h.logger.Error("User not found during token refresh", zap.String("userID", userID.String()), zap.Error(err))
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Generate a new token pair
	accessToken, refreshToken, refreshTokenID, cookie, err := h.tokenManager.GenerateTokenPair(user)
	if err != nil {
		h.logger.Error("Failed to generate new token pair", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate new token pair")
	}

	// Store the new refresh token ID and invalidate the old one
	err = h.service.RotateRefreshTokenID(ctx, userID, jti, refreshTokenID) // Use parsed uuid.UUID
	if err != nil {
		h.logger.Error("Failed to rotate refresh token ID", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to rotate refresh token ID")
	}

	return &pb.RefreshTokenResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         convertUserToProto(user),
		Cookie: &pb.CookieInfo{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HttpOnly,
			MaxAge:   int32(cookie.MaxAge),
		},
	}, nil
}

func (h *UserHandler) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	if err := h.service.HealthCheck(ctx); err != nil {
		return &pb.HealthCheckResponse{Status: "unhealthy"}, nil
	}
	return &pb.HealthCheckResponse{Status: "healthy"}, nil
}

// Removed convertSameSiteToString function as it's not needed

func (h *UserHandler) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.UserResponse, error) {
	user, err := h.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.logger.Error("User not found by email", zap.String("email", req.Email), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}
	return &pb.UserResponse{
		User: convertUserToProto(user),
	}, nil
}

func convertUserToProto(user *models.User) *pb.User {
	if user == nil {
		return nil
	}
	lastLoginStr := ""
	if user.LastLogin.Valid {
		lastLoginStr = user.LastLogin.Time.Format(time.RFC3339)
	}
	return &pb.User{
		UserId:        user.UserID.String(), // Convert UUID to string
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
		LastLogin:     lastLoginStr,
	}
}
