package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/louai60/e-commerce_project/backend/user-service/repository"
	pb "github.com/louai60/e-commerce_project/backend/user-service/proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/louai60/e-commerce_project/backend/user-service/cache"
	"github.com/louai60/e-commerce_project/backend/user-service/models"
	"go.uber.org/zap"
)

type UserService struct {
	repo         repository.Repository
	logger       *zap.Logger
	rateLimiter  RateLimiter
	tokenManager TokenManager
	cacheManager *cache.UserCacheManager
}

type RateLimiter interface {
	Allow(key string) error
	Record(key string)
}

type TokenManager interface {
	GenerateTokenPair(user *models.User) (string, string, *http.Cookie, error)
	ValidateToken(token string) (*models.User, error)
}

func NewUserService(
	repo repository.Repository,
	cache *cache.UserCacheManager,
	logger *zap.Logger,
	rateLimiter RateLimiter,
	tokenManager TokenManager,
) *UserService {
	repoWithLogger, ok := repo.(*repository.PostgresRepository)
	if !ok {
		panic("Repository is not PostgresRepository")
	}
	repoWithLogger.Logger = logger
	return &UserService{
		repo:         repo,
		logger:       logger,
		rateLimiter:  rateLimiter,
		tokenManager: tokenManager,
		cacheManager: cache,
	}
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	// Try to get from cache first
	user, err := s.cacheManager.GetUser(ctx, fmt.Sprintf("%d", id))
	if err == nil {
		s.logger.Debug("Cache hit for user", zap.Int64("id", id))
		return user, nil
	}

	// Cache miss, get from database
	user, err = s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache for future requests
	if err := s.cacheManager.SetUser(ctx, user); err != nil {
		s.logger.Warn("Failed to cache user", zap.Error(err))
	}

	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context, page, limit int32, filters map[string]interface{}) ([]*models.User, int64, error) {
	// Validate pagination
	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 10 }

	// Build query conditions
	var conditions []string
	var args []interface{}
	
	if userType, ok := filters["user_type"]; ok {
		conditions = append(conditions, "user_type = ?")
		args = append(args, userType)
	}
	if role, ok := filters["role"]; ok {
		conditions = append(conditions, "role = ?")
		args = append(args, role)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	users, err := s.repo.ListUsers(ctx, int(page), int(limit), where, args...)
	if err != nil {
		s.logger.Error("ListUsers failed", zap.Error(err))
		return nil, 0, err
	}

	total, err := s.repo.CountUsers(ctx, where, args...)
	if err != nil {
		s.logger.Error("CountUsers failed", zap.Error(err))
		return nil, 0, err
	}

	return users, total, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	s.logger.Info("Creating user with type and role", 
		zap.String("userType", req.UserType),
		zap.String("role", req.Role))

	// Validate user type and role
	if !models.IsValidUserType(req.UserType) {
		return nil, fmt.Errorf("invalid user type: %s", req.UserType)
	}

	if !models.IsValidRole(req.UserType, req.Role) {
		return nil, fmt.Errorf("invalid role %s for user type %s", req.Role, req.UserType)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:         strings.ToLower(req.Email),
		Username:      req.Username,
		HashedPassword: string(hashedPassword),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		PhoneNumber:   req.PhoneNumber,
		UserType:      req.UserType,
		Role:         req.Role,
		AccountStatus: "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.logger.Info("Attempting to create user", 
		zap.String("email", user.Email),
		zap.String("userType", user.UserType),
		zap.String("role", user.Role))

	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	s.logger.Debug("Updating user", zap.Int64("id", user.UserID))

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdatePassword(ctx context.Context, email string, newPassword string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return status.Errorf(codes.NotFound, "user not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return fmt.Errorf("failed to process password")
	}

	user.HashedPassword = string(hashedPassword)

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		s.logger.Error("Failed to update user password", zap.Error(err))
		return err
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	s.logger.Debug("Deleting user", zap.Int64("id", id))
	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	s.logger.Info("Login attempt", 
		zap.String("email", req.Email))

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to find user", 
			zap.String("email", req.Email),
			zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
	if err != nil {
		s.logger.Error("Password mismatch",
			zap.String("email", req.Email),
			zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// Generate token pair
	accessToken, refreshToken, refreshTokenCookie, err := s.tokenManager.GenerateTokenPair(user)
	if err != nil {
		s.logger.Error("Failed to generate tokens",
			zap.String("email", req.Email),
			zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate tokens")
	}

	// Update last login
	user.LastLogin = time.Now()
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		s.logger.Error("Failed to update last login",
			zap.String("email", req.Email),
			zap.Error(err))
		// Don't return error as login was successful
	}

	// Prepare CookieInfo for gRPC response
	cookieInfo := &pb.CookieInfo{
		Name:     refreshTokenCookie.Name,
		Value:    refreshTokenCookie.Value,
		MaxAge:   int32(refreshTokenCookie.MaxAge),
		Path:     refreshTokenCookie.Path,
		Domain:   refreshTokenCookie.Domain,
		Secure:   refreshTokenCookie.Secure,
		HttpOnly: refreshTokenCookie.HttpOnly,
	}

	return &pb.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken, // Keep sending the raw refresh token for potential non-cookie clients
		User:         convertUserToProto(user),
		Cookie:       cookieInfo,
	}, nil
}

func (s *UserService) HealthCheck(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	s.logger.Debug("Retrieving user by email", zap.String("email", email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Failed to retrieve user by email", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// Add Address Ads a new address to User's Profile

func (s *UserService) AddAddress(ctx context.Context, userID int64, address *models.UserAddress) (*models.UserAddress, error) {
	s.logger.Debug("adding address for user", 
		zap.Int64("user_id", userID), 
		zap.String("address_type", address.AddressType))

	// Verify user exists
	_, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error("user not found when adding address", 
			zap.Int64("user_id", userID), 
			zap.Error(err))
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Set address metadata
	address.UserID = userID
	address.CreatedAt = time.Now()
	address.UpdatedAt = time.Now()

	// If this is set as default, update any existing default addresses
	if address.IsDefault {
		if err := s.repo.UpdateAddress(ctx, address); err != nil {
			s.logger.Warn("failed to update existing default addresses", 
				zap.Int64("user_id", userID),
				zap.Error(err))
		}
	}

	// Add address to database
	if err := s.repo.CreateAddress(ctx, address); err != nil {
		s.logger.Error("failed to add address to database",
			zap.Int64("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to add address: %w", err)
	}

	s.logger.Info("successfully added address",
		zap.Int64("address_id", address.AddressID), 
		zap.String("address_type", address.AddressType))
	return address, nil
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*pb.RefreshTokenResponse, error) {
	// Validate the refresh token
	user, err := s.tokenManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token")
	}

	// Generate new token pair
	accessToken, newRefreshToken, _, err := s.tokenManager.GenerateTokenPair(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	return &pb.RefreshTokenResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		User:         convertUserToProto(user),
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
