package service

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"context" 
	"github.com/google/uuid" 
	"github.com/golang-jwt/jwt"
	"github.com/louai60/e-commerce_project/backend/user-service/models"
	"github.com/louai60/e-commerce_project/backend/user-service/repository"
	"go.uber.org/zap"
)

// JWTManager handles all JWT token operations including generation, validation,
// and token pair management for secure authentication.
type JWTManager struct {
	privateKey           *rsa.PrivateKey  // RSA private key for signing tokens
	publicKey            *rsa.PublicKey   // RSA public key for verifying tokens
	accessTokenDuration  time.Duration    // Lifetime of access tokens
	refreshTokenDuration time.Duration    // Lifetime of refresh tokens
	repo                 repository.Repository  // User data repository
	logger               *zap.Logger      // Structured logger for operational insights
}

// NewJWTManager initializes a new JWT token manager with cryptographic keys and configuration.
func NewJWTManager(
	privateKeyPath, publicKeyPath string,
	accessTokenDuration, refreshTokenDuration time.Duration,
	repo repository.Repository,
	logger *zap.Logger,
) (*JWTManager, error) {
	// Load and parse the RSA private key for token signing
	privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read private key file: %w", err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse private key: %w", err)
	}

	// Load and parse the RSA public key (optional for some implementations)
	var publicKey *rsa.PublicKey
	if publicKeyBytes, err := ioutil.ReadFile(publicKeyPath); err == nil {
		if publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes); err != nil {
			return nil, fmt.Errorf("could not parse public key: %w", err)
		}
	}

	// Initialize with default no-op logger if none provided
	if logger == nil {
		logger = zap.NewNop()
	}

	return &JWTManager{
		privateKey:           privateKey,
		publicKey:            publicKey,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		repo:                 repo,
		logger:               logger,
	}, nil
}

// GetPublicKey provides access to the public key for external token validation.
func (m *JWTManager) GetPublicKey() (*rsa.PublicKey, error) {
	if m.publicKey == nil {
		return nil, fmt.Errorf("public key not available")
	}
	return m.publicKey, nil
}

// GenerateTokenPair creates a new set of access and refresh tokens for user authentication.
func (m *JWTManager) GenerateTokenPair(user *models.User) (string, string, string, *http.Cookie, error) {
	// Create unique identifier for refresh token tracking
	refreshTokenID := uuid.New().String()

	// Base claims common to both token types
	commonClaims := jwt.MapClaims{
		"user_id":   user.UserID,
		"email":     user.Email,
		"username":  user.Username,
		"role":      user.Role,
		"user_type": user.UserType,
		"iat":       time.Now().Unix(),  // Issued at timestamp
	}

	// Generate access token with shorter lifespan
	accessTokenString, err := m.generateToken("access", commonClaims)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("access token generation failed: %w", err)
	}

	// Generate refresh token with extended lifespan and tracking ID
	refreshTokenString, err := m.generateToken("refresh", commonClaims, refreshTokenID)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("refresh token generation failed: %w", err)
	}

	// Configure secure HTTP cookie for refresh token storage
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshTokenString,
		Path:     "/api/v1/users/refresh",
		HttpOnly: true,       // Prevent JavaScript access
		Secure:   true,       // Require HTTPS in production
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(m.refreshTokenDuration.Seconds()),
	}

	return accessTokenString, refreshTokenString, refreshTokenID, refreshCookie, nil
}

// ValidateToken thoroughly checks a refresh token's validity and ownership.
func (m *JWTManager) ValidateToken(tokenString string) (*models.User, error) {
	// Verify token signature and basic validity
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	// Validate token claims structure
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Confirm this is a refresh token specifically
	if claims["type"] != "refresh" {
		return nil, fmt.Errorf("incorrect token type")
	}

	// Extract and validate required claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id claim is not a string or is missing")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id format in token: %w", err)
	}

	tokenID, ok := claims["jti"].(string)
	if !ok || tokenID == "" {
		return nil, fmt.Errorf("missing token identifier")
	}

	// Retrieve user data and verify token against stored record
	user, err := m.repo.GetUser(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("user verification failed: %w", err)
	}

	// Critical security check - ensures token hasn't been revoked
	if user.RefreshTokenID != tokenID {
		return nil, fmt.Errorf("token no longer valid")
	}

	return user, nil
}

// generateToken constructs and signs a JWT with specified properties
func (m *JWTManager) generateToken(tokenType string, baseClaims jwt.MapClaims, extra ...string) (string, error) {
	claims := jwt.MapClaims{}
	for k, v := range baseClaims {
		claims[k] = v
	}

	// Set token-specific properties
	claims["exp"] = m.getTokenExpiration(tokenType).Unix()
	claims["type"] = tokenType

	// Add refresh token identifier if provided
	if tokenType == "refresh" && len(extra) > 0 {
		claims["jti"] = extra[0]
	}

	// Debug logging
	m.logger.Debug("Generating token with claims",
		zap.String("type", tokenType),
		zap.Any("claims", claims))

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

// getTokenExpiration calculates the expiry time based on token type
func (m *JWTManager) getTokenExpiration(tokenType string) time.Time {
	switch tokenType {
	case "access":
		return time.Now().Add(m.accessTokenDuration)
	case "refresh":
		return time.Now().Add(m.refreshTokenDuration)
	default:
		return time.Now().Add(time.Hour) // Default fallback
	}
}

// GetRefreshTokenDuration exposes the configured refresh token lifespan
func (m *JWTManager) GetRefreshTokenDuration() time.Duration {
	return m.refreshTokenDuration
}
