package service

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/louai60/e-commerce_project/backend/user-service/models"
    "github.com/louai60/e-commerce_project/backend/user-service/repository"
    "go.uber.org/zap"
    "golang.org/x/crypto/bcrypt"
)

type UserService struct {
    repo         repository.UserRepository
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
    repo repository.UserRepository,
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

func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
    s.logger.Debug("Getting user by ID", zap.String("id", id))
    return s.repo.GetUser(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, page, limit int32) ([]*models.User, int64, error) {
    s.logger.Debug("Listing users", zap.Int32("page", page), zap.Int32("limit", limit))
    return s.repo.ListUsers(ctx, page, limit)
}

func (s *UserService) CreateUser(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
    s.logger.Debug("Creating new user", zap.String("email", req.Email))

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        s.logger.Error("Failed to hash password", zap.Error(err))
        return nil, errors.New("internal server error")
    }

    user := &models.User{
        ID:        uuid.New().String(),
        Email:     req.Email,
        Username:  req.Username,
        Password:  string(hashedPassword),
        FirstName: req.FirstName,
        LastName:  req.LastName,
        Role:      "user", // Default role
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    if err := s.repo.CreateUser(ctx, user); err != nil {
        s.logger.Error("Failed to create user", zap.Error(err))
        return nil, err
    }

    // Don't return the password hash
    user.Password = ""
    return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
    s.logger.Debug("Updating user", zap.String("id", user.ID))
    
    if err := s.repo.UpdateUser(ctx, user); err != nil {
        s.logger.Error("Failed to update user", zap.Error(err))
        return nil, err
    }

    return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
    s.logger.Debug("Deleting user", zap.String("id", id))
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
        []byte(user.Password), 
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
    user.Password = ""

    return &models.LoginResponse{
        Token:        token,
        RefreshToken: refreshToken,
        User:         user,
    }, nil
}

func (s *UserService) HealthCheck(ctx context.Context) error {
    return s.repo.Ping(ctx)
}




