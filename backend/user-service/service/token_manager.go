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
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(m.refreshTokenDuration.Seconds()),
		Domain:   ".nexcart.com", // Update with your domain
	}

	return accessTokenString, refreshTokenString, refreshTokenCookie, nil
}

func (m *JWTManager) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		return nil, jwt.ValidationError{Errors: jwt.ValidationErrorExpired}
	}

	// For refresh tokens, we only need the user ID
	if claims["type"] == "refresh" {
		return &models.User{
			UserID: claims["user_id"].(int64),
		}, nil
	}

	// For access tokens, we need all user details
	return &models.User{
		UserID:   claims["user_id"].(int64),
		Email:    claims["email"].(string),
		Username: claims["username"].(string),
		UserType: claims["user_type"].(string),
		Role:     claims["role"].(string),
	}, nil
}
