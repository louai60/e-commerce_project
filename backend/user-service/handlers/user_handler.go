package handlers

import (
	"context"
	"fmt"
	// "net/http" // Removed unused import
	// "strconv" // Removed unused import
	"strings"
	"time"

	"github.com/golang-jwt/jwt"

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
	// Use req.UserId directly as int64, assuming proto definition is int64
	userID := req.UserId

	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		// Consider differentiating between not found and other errors if the service layer provides more detail
		h.logger.Warn("User not found", zap.Int64("userID", userID), zap.Error(err)) // Use zap.Int64
		return nil, status.Error(codes.NotFound, "user not found")
	}
	// Removed duplicated error check block

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
		zap.String("username", req.Username),
		zap.String("userType", req.UserType),
		zap.String("role", req.Role))

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
	// Use req.UserId directly as int64
	userID := req.UserId

	// TODO: Consider fetching the existing user first to apply partial updates
	// or ensure the service layer handles partial updates correctly.
	user := &models.User{
		UserID:    userID,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	updatedUser, err := h.service.UpdateUser(ctx, user)
	if err != nil {
		// TODO: Add more specific error handling based on potential service layer errors (e.g., not found, validation)
		h.logger.Error("Failed to update user", zap.Int64("userID", userID), zap.Error(err)) // Already using zap.Int64 here, which is correct
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return &pb.UserResponse{
		User: convertUserToProto(updatedUser),
	}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteResponse, error) {
	// Use req.UserId directly as int64
	userID := req.UserId

	err := h.service.DeleteUser(ctx, userID)
	if err != nil {
		// TODO: Add more specific error handling based on potential service layer errors (e.g., not found)
		h.logger.Error("Failed to delete user", zap.Int64("userID", userID), zap.Error(err)) // Use zap.Int64
		return nil, status.Error(codes.Internal, "failed to delete user")
	}
	// Removed duplicated error check block

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

	// Store the refresh token ID in the database
	err = h.service.UpdateRefreshTokenID(ctx, user.UserID, refreshTokenID)
	if err != nil {
		h.logger.Error("Failed to store refresh token ID", zap.Error(err))
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

	token, err := jwt.ParseWithClaims(req.RefreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
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

	// Ensure user_id claim is treated as float64 as per standard JWT number encoding, then convert to int64
	userIDClaim, ok := claims["user_id"].(float64)
	if !ok {
		h.logger.Error("Invalid user_id claim type in token", zap.Any("claim", claims["user_id"]))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id claim type in token")
	}
	userID := int64(userIDClaim)

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid jti in token")
	}

	// Validate the JTI against the stored JTI
	valid, err := h.service.ValidateRefreshTokenID(ctx, userID, jti)
	if err != nil || !valid {
		h.logger.Error("Invalid refresh token ID", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token ID")
	}

	// Get the user by the ID from the token
	user, err := h.service.GetUser(ctx, userID) // Use the parsed userID (int64)
	if err != nil {
		h.logger.Error("User not found during token refresh", zap.Int64("userID", userID), zap.Error(err))
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Generate a new token pair
	accessToken, refreshToken, refreshTokenID, cookie, err := h.tokenManager.GenerateTokenPair(user)
	if err != nil {
		h.logger.Error("Failed to generate new token pair", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate new token pair")
	}

	// Store the new refresh token ID and invalidate the old one
	err = h.service.RotateRefreshTokenID(ctx, userID, jti, refreshTokenID)
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
	lastLoginStr := ""
	// Check if LastLogin is valid (not NULL) before formatting
	if user.LastLogin.Valid {
		lastLoginStr = user.LastLogin.Time.Format(time.RFC3339)
	}
	return &pb.User{
		UserId:        user.UserID, // Assign int64 directly
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
		LastLogin:     lastLoginStr, // Use the potentially empty string
	}
}
