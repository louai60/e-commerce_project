package service

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/louai60/e-commerce_project/backend/user-service/models"
)

type JWTManager struct {
	privateKey           *rsa.PrivateKey
	publicKey            *rsa.PublicKey
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewJWTManager(privateKeyPath, publicKeyPath string, accessTokenDuration, refreshTokenDuration time.Duration) (*JWTManager, error) {
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


	return &JWTManager{
		privateKey:           privateKey,
		publicKey:            publicKey, // Store the public key
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}, nil
}

func (m *JWTManager) GenerateTokenPair(user *models.User) (string, string, *http.Cookie, error) {
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
		return "", "", nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token claims
	refreshClaims := jwt.MapClaims{
		"user_id": user.UserID,
		"type":    "refresh",
		"iat":     now.Unix(),
		"exp":     now.Add(m.refreshTokenDuration).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(m.privateKey)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to sign refresh token: %w", err)
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

	return accessTokenString, refreshTokenString, refreshTokenCookie, nil
}

// GetRefreshTokenDuration returns the configured duration for refresh tokens.
func (m *JWTManager) GetRefreshTokenDuration() time.Duration {
	return m.refreshTokenDuration
}

func (m *JWTManager) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Check if public key is loaded before returning
		if m.publicKey == nil {
			return nil, fmt.Errorf("public key not loaded for validation")
		}
		return m.publicKey, nil
	})

	if err != nil {
		// Check specifically for expired token error
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, fmt.Errorf("token has expired: %w", err)
			}
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check expiration again (though Parse should handle it)
	if exp, ok := claims["exp"].(float64); ok {
		if float64(time.Now().Unix()) > exp {
			return nil, fmt.Errorf("token has expired")
		}
	} else {
		return nil, fmt.Errorf("invalid expiration claim")
	}

	// Extract user ID safely
	userIDClaim, ok := claims["user_id"]
	if !ok {
		return nil, fmt.Errorf("user_id claim missing")
	}
	
	var userID int64
	switch v := userIDClaim.(type) {
	case float64:
		userID = int64(v)
	case int64:
		userID = v
	default:
		return nil, fmt.Errorf("invalid user_id claim type")
	}


	// For refresh tokens, we only need the user ID
	if tokenType, ok := claims["type"].(string); ok && tokenType == "refresh" {
		return &models.User{
			UserID: userID,
		}, nil
	}

	// For access tokens, extract all user details safely
	email, _ := claims["email"].(string)
	username, _ := claims["username"].(string)
	userType, _ := claims["user_type"].(string)
	role, _ := claims["role"].(string)

	return &models.User{
		UserID:   userID,
		Email:    email,
		Username: username,
		UserType: userType,
		Role:     role,
	}, nil
}
