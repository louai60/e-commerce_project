package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	pb "github.com/louai60/e-commerce_project/backend/user-service/proto"
	"github.com/louai60/e-commerce_project/backend/user-service/repository"
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
	cacheManager cache.CacheInterface
}

type RateLimiter interface {
	Allow(key string) error
	Record(key string)
}

type TokenManager interface {
	GenerateTokenPair(user *models.User) (string, string, string, *http.Cookie, error)
	ValidateToken(token string) (*models.User, error)
	GetRefreshTokenDuration() time.Duration
}

type UserServiceI interface {
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	ListUsers(ctx context.Context, page, limit int32, filters map[string]any) ([]*models.User, int64, error)
	CreateUser(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	UpdatePassword(ctx context.Context, email string, newPassword string) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
	HealthCheck(ctx context.Context) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	AddAddress(ctx context.Context, userID uuid.UUID, address *models.UserAddress) (*models.UserAddress, error)
	RefreshToken(ctx context.Context, refreshToken string) (*pb.RefreshTokenResponse, error)

	Authenticate(ctx context.Context, email, password string) (*models.User, error)
	UpdateRefreshTokenID(ctx context.Context, userID uuid.UUID, refreshTokenID string) error
	ValidateRefreshTokenID(ctx context.Context, userID uuid.UUID, refreshTokenID string) (bool, error)
	RotateRefreshTokenID(ctx context.Context, userID uuid.UUID, oldRefreshTokenID, newRefreshTokenID string) error
}

func NewUserService(
	repo repository.Repository,
	cache cache.CacheInterface,
	logger *zap.Logger,
	rateLimiter RateLimiter,
	tokenManager *JWTManager,
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

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	s.logger.Info("Authenticating user", zap.String("email", email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Failed to find user",
			zap.String("email", email),
			zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		s.logger.Error("Password mismatch",
			zap.String("email", email),
			zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	return user, nil
}

func (s *UserService) UpdateRefreshTokenID(ctx context.Context, userID uuid.UUID, refreshTokenID string) error {
	s.logger.Info("Updating refresh token ID",
		zap.String("userID", userID.String()))

	user, err := s.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user",
			zap.String("userID", userID.String()),
			zap.Error(err))
		return err
	}

	user.RefreshTokenID = refreshTokenID

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		s.logger.Error("Failed to update user",
			zap.String("userID", userID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *UserService) ValidateRefreshTokenID(ctx context.Context, userID uuid.UUID, refreshTokenID string) (bool, error) {
	s.logger.Info("Validating refresh token ID", zap.String("userID", userID.String()))

	user, err := s.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.String("userID", userID.String()), zap.Error(err))
		return false, err
	}

	return user.RefreshTokenID == refreshTokenID, nil
}

func (s *UserService) RotateRefreshTokenID(ctx context.Context, userID uuid.UUID, oldRefreshTokenID, newRefreshTokenID string) error {
	s.logger.Info("Rotating refresh token ID", zap.String("userID", userID.String()))

	user, err := s.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.String("userID", userID.String()), zap.Error(err))
		return err
	}

	if user.RefreshTokenID != oldRefreshTokenID {
		s.logger.Warn("Old refresh token ID does not match", zap.String("userID", userID.String()))
		return fmt.Errorf("old refresh token ID does not match")
	}

	user.RefreshTokenID = newRefreshTokenID

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		s.logger.Error("Failed to update user", zap.String("userID", userID.String()), zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	// Try to get from cache first
	user, err := s.cacheManager.GetUser(ctx, id.String())
	if err == nil {
		s.logger.Debug("Cache hit for user", zap.String("id", id.String()))
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

func (s *UserService) ListUsers(ctx context.Context, page, limit int32, filters map[string]any) ([]*models.User, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Build query conditions
	var conditions []string
	var args []any

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
		zap.String("email", req.Email), // Log email
		zap.String("userType", req.UserType),
		zap.String("role", req.Role))

	// Validate user type and role
	if !models.IsValidUserType(req.UserType) {
		s.logger.Error("Invalid user type provided", zap.String("userType", req.UserType))
		return nil, status.Errorf(codes.InvalidArgument, "invalid user type: %s", req.UserType)
	}

	if !models.IsValidRole(req.UserType, req.Role) {
		s.logger.Error("Invalid role for user type", zap.String("role", req.Role), zap.String("userType", req.UserType))
		return nil, status.Errorf(codes.InvalidArgument, "invalid role %s for user type %s", req.Role, req.UserType)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email: strings.ToLower(req.Email),
		// Username will be set to email by default in repository if empty
		HashedPassword: string(hashedPassword),
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		// PhoneNumber is omitted
		UserType:      req.UserType, // Use provided UserType
		Role:          req.Role,     // Use provided Role
		AccountStatus: "active",     // Default AccountStatus
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.logger.Info("Attempting to create user in repository",
		zap.String("email", user.Email),
		zap.String("userType", user.UserType),
		zap.String("role", user.Role))

	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Error("Failed to create user in repository", zap.Error(err))
		// Check for specific errors like duplicate email/username
		if strings.Contains(err.Error(), "already exists") {
			// Use the specific error message from the repository if available
			return nil, status.Errorf(codes.AlreadyExists, "user already exists: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err.Error())
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	s.logger.Debug("Updating user", zap.String("id", user.UserID.String()))

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

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("Deleting user", zap.String("id", id.String()))
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
	accessToken, _, refreshTokenID, refreshTokenCookie, err := s.tokenManager.GenerateTokenPair(user) // Use blank identifier for refreshToken string
	if err != nil {
		s.logger.Error("Failed to generate tokens",
			zap.String("email", req.Email),
			zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate tokens")
	}

	// Update user object with new RefreshTokenID and LastLogin time
	user.RefreshTokenID = refreshTokenID
	user.LastLogin = sql.NullTime{Time: time.Now(), Valid: true}

	// *** Perform a single update for both RefreshTokenID and LastLogin ***
	// We need a repository function that updates these specific fields.
	// Let's assume a function UpdateUserLoginDetails exists or modify UpdateUser.
	// For now, let's modify the existing UpdateUser call to include RefreshTokenID.
	// We'll need to adjust the UpdateUser function in the repository accordingly later if needed.
	// NOTE: This assumes UpdateUser will be modified to handle RefreshTokenID and LastLogin.
	// If UpdateUser cannot be modified, a new specific repository function is required.
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		s.logger.Error("Failed to update user details (RefreshTokenID, LastLogin)",
			zap.String("userID", user.UserID.String()),
			zap.Error(err))
		// If storing the token state is critical, return an error.
		return nil, status.Errorf(codes.Internal, "failed to update user state after login: %s", err.Error())
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
		Token: accessToken,
		// RefreshToken: refreshToken, // Removed - Handled by cookie
		User:   convertUserToProto(user),
		Cookie: cookieInfo,
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

func (s *UserService) AddAddress(ctx context.Context, userID uuid.UUID, address *models.UserAddress) (*models.UserAddress, error) {
	s.logger.Debug("adding address for user",
		zap.String("user_id", userID.String()),
		zap.String("address_type", address.AddressType))

	// Verify user exists
	_, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error("user not found when adding address",
			zap.String("user_id", userID.String()),
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
				zap.String("user_id", userID.String()),
				zap.Error(err))
		}
	}

	// Add address to database
	if err := s.repo.CreateAddress(ctx, address); err != nil {
		s.logger.Error("failed to add address to database",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to add address: %w", err)
	}

	s.logger.Info("successfully added address",
		zap.String("address_id", address.AddressID.String()),
		zap.String("address_type", address.AddressType))
	return address, nil
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*pb.RefreshTokenResponse, error) {
	// Validate the refresh token
	user, err := s.tokenManager.ValidateToken(refreshToken)
	if err != nil {
		s.logger.Error("Invalid refresh token provided", zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired refresh token: %s", err.Error())
	}

	// Generate NEW token pair (this includes a new JTI)
	accessToken, newRefreshTokenString, newRefreshTokenID, newRefreshTokenCookie, err := s.tokenManager.GenerateTokenPair(user)
	if err != nil {
		s.logger.Error("Failed to generate new token pair during refresh", zap.String("userID", user.UserID.String()), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %s", err.Error())
	}

	// *** Store the NEW refresh token ID, rotating the old one ***
	if err := s.repo.UpdateRefreshTokenID(ctx, user.UserID, newRefreshTokenID); err != nil {
		s.logger.Error("Failed to store new refresh token ID during refresh",
			zap.String("userID", user.UserID.String()),
			zap.Error(err))
		// This is critical. If we can't store the new ID, the user might be locked out after the old token expires.
		return nil, status.Errorf(codes.Internal, "failed to update refresh token state")
	}

	// Prepare CookieInfo for the new refresh token
	newCookieInfo := &pb.CookieInfo{
		Name:     newRefreshTokenCookie.Name, // Use the generated cookie details
		Value:    newRefreshTokenString,      // Use the generated refresh token string
		MaxAge:   int32(newRefreshTokenCookie.MaxAge),
		Path:     newRefreshTokenCookie.Path,
		Domain:   newRefreshTokenCookie.Domain,
		Secure:   newRefreshTokenCookie.Secure,
		HttpOnly: newRefreshTokenCookie.HttpOnly,
	}

	return &pb.RefreshTokenResponse{
		Token:  accessToken,
		User:   convertUserToProto(user),
		Cookie: newCookieInfo,
		// RefreshToken field remains removed
	}, nil
}

func convertUserToProto(user *models.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		UserId:        user.UserID.String(),
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
		LastLogin:     user.LastLogin.Time.Format(time.RFC3339), // Use the potentially empty string
	}
}
