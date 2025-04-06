package service

import (
	"fmt"
	"time"
	"github.com/golang-jwt/jwt"
	"github.com/louai60/e-commerce_project/backend/user-service/models"
	"net/http"
)

type JWTManager struct {
	secretKey           string
	accessTokenDuration time.Duration
	refreshTokenDuration time.Duration
}

func NewJWTManager(secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:           secretKey,
		accessTokenDuration: accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
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

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(m.secretKey))
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

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(m.secretKey))
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
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
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
