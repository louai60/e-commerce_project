package service

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"context" // Added for repository methods

	"github.com/golang-jwt/jwt"
	"github.com/louai60/e-commerce_project/backend/user-service/models"
	"github.com/louai60/e-commerce_project/backend/user-service/repository" // Added for repository interface
	"go.uber.org/zap"                                                       // Added for logging
)

type JWTManager struct {
	privateKey           *rsa.PrivateKey
	publicKey            *rsa.PublicKey
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	repo                 repository.Repository // Added repository field
	logger               *zap.Logger           // Added logger field
}

func NewJWTManager(
	privateKeyPath, publicKeyPath string,
	accessTokenDuration, refreshTokenDuration time.Duration,
	repo repository.Repository, // Added repository parameter
	logger *zap.Logger, // Added logger parameter
) (*JWTManager, error) {
	// Read private key
	privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Read public key
	publicKeyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		// Allow public key to be optional if not needed for validation within this service
		// Or handle error strictly depending on requirements
		fmt.Fprintf(os.Stderr, "Warning: failed to read public key file: %v\n", err)
	}
	var publicKey *rsa.PublicKey
	if len(publicKeyBytes) > 0 {
		publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	if logger == nil {
		// Fallback to a no-op logger if none is provided
		logger = zap.NewNop()
	}

	return &JWTManager{
		privateKey:           privateKey,
		publicKey:            publicKey, // Store the public key
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		repo:                 repo,   // Assign repository
		logger:               logger, // Assign logger
	}, nil
}

// GetPublicKey returns the RSA public key used by the manager.
// It returns an error if the public key was not loaded.
func (m *JWTManager) GetPublicKey() (*rsa.PublicKey, error) {
	if m.publicKey == nil {
		return nil, fmt.Errorf("public key not loaded or available")
	}
	return m.publicKey, nil
}

// generateRandomJTI generates a random JTI (JWT ID)
func generateRandomJTI() (string, error) {
	randomBytes := make([]byte, 32) // 32 bytes for a 256-bit key
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(randomBytes), nil
}

func (m *JWTManager) GenerateTokenPair(user *models.User) (string, string, string, *http.Cookie, error) {
	now := time.Now()

	// Access token claims
	accessClaims := jwt.MapClaims{
		"user_id":   user.UserID,
		"email":     user.Email,
		"username":  user.Username,
		"user_type": user.UserType,
		"role":      user.Role,
		"type":      "access",
		"iat":       now.Unix(),
		"exp":       now.Add(m.accessTokenDuration).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(m.privateKey)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate a unique JTI for the refresh token
	jti, err := generateRandomJTI()
	if err != nil {
		return "", "", "", nil, fmt.Errorf("failed to generate JTI: %w", err)
	}

	// Refresh token claims
	refreshClaims := jwt.MapClaims{
		"user_id": user.UserID,
		"type":    "refresh",
		"iat":     now.Unix(),
		"exp":     now.Add(m.refreshTokenDuration).Unix(),
		"jti":     jti, // Include the JTI
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(m.privateKey)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Set secure cookie options
	refreshTokenCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshTokenString,
		Path:     "/api/v1/users/refresh", // Specific path for the refresh endpoint
		HttpOnly: true,
		Secure:   true, // Ensure this is true in production (requires HTTPS)
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(m.refreshTokenDuration.Seconds()),
		Domain:   "", // Set domain appropriately in production, e.g., ".yourdomain.com"
	}

	return accessTokenString, refreshTokenString, jti, refreshTokenCookie, nil
}

// GetRefreshTokenDuration returns the configured duration for refresh tokens.
func (m *JWTManager) GetRefreshTokenDuration() time.Duration {
	return m.refreshTokenDuration
}

// ValidateToken validates a refresh token, checks its type, expiration, and JTI against the database.
func (m *JWTManager) ValidateToken(tokenString string) (*models.User, error) {
	m.logger.Debug("Validating token string")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			m.logger.Error("Unexpected signing method", zap.Any("alg", token.Header["alg"]))
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if m.publicKey == nil {
			m.logger.Error("Public key not loaded for validation")
			return nil, fmt.Errorf("public key not loaded for validation")
		}
		return m.publicKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				m.logger.Warn("Token has expired", zap.Error(err))
				return nil, fmt.Errorf("token has expired")
			}
		}
		m.logger.Error("Invalid token during parsing", zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		m.logger.Warn("Token marked as invalid")
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		m.logger.Error("Invalid token claims type")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check token type
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		m.logger.Warn("Invalid token type or type missing", zap.Any("type", claims["type"]))
		return nil, fmt.Errorf("invalid token type, expected 'refresh'")
	}

	// Check expiration (redundant check, Parse handles it, but good practice)
	if exp, ok := claims["exp"].(float64); ok {
		if float64(time.Now().Unix()) > exp {
			m.logger.Warn("Token expired (manual check)")
			return nil, fmt.Errorf("token has expired")
		}
	} else {
		m.logger.Error("Invalid expiration claim type")
		return nil, fmt.Errorf("invalid expiration claim")
	}

	// Extract user ID
	userIDFloat, ok := claims["user_id"].(float64) // JWT numbers are often float64
	if !ok {
		m.logger.Error("user_id claim missing or not a number", zap.Any("user_id", claims["user_id"]))
		return nil, fmt.Errorf("user_id claim missing or invalid")
	}
	userID := int64(userIDFloat)

	// Extract JTI
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		m.logger.Error("jti claim missing or empty", zap.Any("jti", claims["jti"]))
		return nil, fmt.Errorf("jti claim missing or invalid")
	}

	// Fetch user from database
	// Use a background context or pass one in if appropriate
	ctx := context.Background()
	user, err := m.repo.GetUser(ctx, userID)
	if err != nil {
		m.logger.Error("Failed to get user from repository", zap.Int64("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("user not found or db error: %w", err)
	}

	// *** The core rotation check: Compare token JTI with stored RefreshTokenID ***
	if user.RefreshTokenID == "" {
		m.logger.Error("User does not have a refresh token ID set", zap.Int64("userID", userID))
		return nil, fmt.Errorf("refresh token validation failed: no token ID stored for user")
	}
	if user.RefreshTokenID != jti {
		m.logger.Warn("Refresh token JTI mismatch",
			zap.Int64("userID", userID),
			zap.String("token_jti", jti),
			zap.String("stored_jti", user.RefreshTokenID))
		// This indicates the token might be stolen or an old one is being reused.
		// Depending on security policy, you might want to revoke all tokens for this user here.
		return nil, fmt.Errorf("invalid refresh token: ID mismatch")
	}

	m.logger.Info("Refresh token validated successfully", zap.Int64("userID", userID), zap.String("jti", jti))
	return user, nil // Return the user object on successful validation
}
