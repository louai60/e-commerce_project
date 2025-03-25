package service

import (
    "context"
    "testing"
    "time"

    "github.com/louai60/e-commerce_project/backend/user-service/models"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "go.uber.org/zap"
)

// Mock implementations
type MockRepository struct {
    mock.Mock
}

type MockRateLimiter struct {
    mock.Mock
}

type MockTokenManager struct {
    mock.Mock
}

// MockRateLimiter implementations
func (m *MockRateLimiter) Allow(key string) error {
    args := m.Called(key)
    return args.Error(0)
}

func (m *MockRateLimiter) Record(key string) {
    m.Called(key)
}

// MockTokenManager implementations
func (m *MockTokenManager) GenerateTokenPair(user *models.User) (string, string, error) {
    args := m.Called(user)
    return args.String(0), args.String(1), args.Error(2)
}

func (m *MockTokenManager) ValidateToken(token string) (*models.User, error) {
    args := m.Called(token)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

// Repository mock methods (keep existing implementations)
func (m *MockRepository) GetUser(ctx context.Context, id string) (*models.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) ListUsers(ctx context.Context, page, limit int32) ([]*models.User, int64, error) {
    args := m.Called(ctx, page, limit)
    return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *models.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id string) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockRepository) Ping(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func TestUserService_GetUser(t *testing.T) {
    // Initialize all mocks
    mockRepo := new(MockRepository)
    mockRateLimiter := new(MockRateLimiter)
    mockTokenManager := new(MockTokenManager)
    logger := zap.NewNop()

    // Create service with all required dependencies
    service := NewUserService(
        mockRepo,
        logger,
        mockRateLimiter,
        mockTokenManager,
    )

    ctx := context.Background()
    expectedUser := &models.User{
        ID:        "test-id",
        Email:     "test@example.com",
        Username:  "testuser",
        FirstName: "Test",
        LastName:  "User",
        Role:      "user",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    mockRepo.On("GetUser", ctx, "test-id").Return(expectedUser, nil)

    user, err := service.GetUser(ctx, "test-id")

    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
    mockRepo.AssertExpectations(t)
}

// Example of a Login test
func TestUserService_Login(t *testing.T) {
    mockRepo := new(MockRepository)
    mockRateLimiter := new(MockRateLimiter)
    mockTokenManager := new(MockTokenManager)
    logger := zap.NewNop()

    service := NewUserService(
        mockRepo,
        logger,
        mockRateLimiter,
        mockTokenManager,
    )

    ctx := context.Background()
    credentials := &models.LoginCredentials{
        Email:    "test@example.com",
        Password: "password123",
    }

    storedUser := &models.User{
        ID:       "test-id",
        Email:    "test@example.com",
        Password: "$2a$10$somehashedpassword", // This should be a proper bcrypt hash
    }

    // Set up expectations
    mockRateLimiter.On("Allow", credentials.Email).Return(nil)
    mockRepo.On("GetUserByEmail", ctx, credentials.Email).Return(storedUser, nil)
    mockTokenManager.On("GenerateTokenPair", storedUser).Return("access_token", "refresh_token", nil)
    // Add the missing expectation for Record method
    mockRateLimiter.On("Record", credentials.Email).Return()

    response, err := service.Login(ctx, credentials)

    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.Equal(t, "access_token", response.Token)
    assert.Equal(t, "refresh_token", response.RefreshToken)

    mockRateLimiter.AssertExpectations(t)
    mockRepo.AssertExpectations(t)
    mockTokenManager.AssertExpectations(t)
}

