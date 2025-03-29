package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/louai60/e-commerce_project/backend/user-service/models"
	"github.com/louai60/e-commerce_project/backend/user-service/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

)

type UserService struct {
	repo         repository.Repository
	logger       *zap.Logger
	rateLimiter  RateLimiter
	tokenManager TokenManager
}

type RateLimiter interface {
	Allow(key string) error
	Record(key string)
}

type TokenManager interface {
	GenerateTokenPair(user *models.User) (string, string, error)
	ValidateToken(token string) (*models.User, error)
}

func NewUserService(
	repo repository.Repository,
	logger *zap.Logger,
	rateLimiter RateLimiter,
	tokenManager TokenManager,
) *UserService {
	return &UserService{
		repo:         repo,
		logger:       logger,
		rateLimiter:  rateLimiter,
		tokenManager: tokenManager,
	}
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	s.logger.Debug("Getting user by ID", zap.Int64("id", id))
	return s.repo.GetUser(ctx, id)
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
	// Validate email format
	if !strings.Contains(req.Email, "@") {
		return nil, fmt.Errorf("invalid email format")
	}

	// Check if email already exists
	existingUser, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Check if username already exists
	existingUser, err = s.repo.GetUserByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username already taken")
	}

	// Validate user type
	if req.UserType == "" {
		req.UserType = "customer" // Set default user type
	}

	// Validate role
	if req.Role == "" {
		req.Role = "user" // Set default role
	}

	if !models.IsValidRole(req.UserType, req.Role) {
		return nil, fmt.Errorf("invalid role '%s' for user type '%s'", req.Role, req.UserType)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to process password")
	}

	user := &models.User{
		Email:         strings.ToLower(req.Email),
		Username:      req.Username,
		PasswordHash:  string(hashedPassword),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		PhoneNumber:   req.PhoneNumber,
		UserType:      req.UserType,
		Role:          req.Role,
		AccountStatus: "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Error("Failed to create user in database", zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
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

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	s.logger.Debug("Deleting user", zap.Int64("id", id))
	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) Login(ctx context.Context, credentials *models.LoginCredentials) (*models.LoginResponse, error) {
	s.logger.Debug("Processing login request",
		zap.String("email", credentials.Email))

	// Add rate limiting
	if err := s.rateLimiter.Allow(credentials.Email); err != nil {
		s.logger.Warn("Rate limit exceeded",
			zap.String("email", credentials.Email))
		return nil, errors.New("too many attempts")
	}

	user, err := s.repo.GetUserByEmail(ctx, credentials.Email)
	if err != nil {
		s.logger.Debug("User not found",
			zap.String("email", credentials.Email),
			zap.Error(err))
		s.rateLimiter.Record(credentials.Email)
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(credentials.Password))
	if err != nil {
		s.logger.Debug("Password mismatch",
			zap.String("email", credentials.Email))
		s.rateLimiter.Record(credentials.Email)
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, refreshToken, err := s.tokenManager.GenerateTokenPair(user)
	if err != nil {
		s.logger.Error("Failed to generate token",
			zap.String("email", credentials.Email),
			zap.Error(err))
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login time
	now := time.Now()
	user.LastLogin = &now
	if _, err := s.UpdateUser(ctx, user); err != nil {
		s.logger.Warn("Failed to update last login time",
			zap.String("email", credentials.Email),
			zap.Error(err))
		// Don't return error as login was successful
	}

	// Clear sensitive data
	user.PasswordHash = ""

	return &models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
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



