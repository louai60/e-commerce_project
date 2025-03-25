package service

import (
    "time"
    "github.com/golang-jwt/jwt"
    "github.com/louai60/e-commerce_project/backend/user-service/models"
)

type JWTManager struct {
    secretKey     string
    tokenDuration time.Duration
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
    return &JWTManager{
        secretKey:     secretKey,
        tokenDuration: tokenDuration,
    }
}

func (m *JWTManager) GenerateTokenPair(user *models.User) (string, string, error) {
    claims := jwt.MapClaims{
        "user_id":  user.ID,
        "email":    user.Email,
        "role":     user.Role,
        "exp":      time.Now().Add(m.tokenDuration).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    
    // Generate access token
    accessToken, err := token.SignedString([]byte(m.secretKey))
    if err != nil {
        return "", "", err
    }

    // Generate refresh token with longer expiration
    refreshClaims := jwt.MapClaims{
        "user_id": user.ID,
        "exp":     time.Now().Add(m.tokenDuration * 24 * 7).Unix(), // 7 days
    }
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenString, err := refreshToken.SignedString([]byte(m.secretKey))
    if err != nil {
        return "", "", err
    }

    return accessToken, refreshTokenString, nil
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

    return &models.User{
        ID:    claims["user_id"].(string),
        Email: claims["email"].(string),
        Role:  claims["role"].(string),
    }, nil
}