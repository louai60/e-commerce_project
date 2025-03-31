package middleware

import (
    "net/http"
    "strings"
    "os"
    "fmt"


    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt"
)

func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
            c.Abort()
            return
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
            c.Abort()
            return
        }

        token := bearerToken[1]
        claims, err := validateToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
            c.Abort()
            return
        }

        // Set user information in context
        c.Set("user_id", claims["user_id"])
        c.Set("user_role", claims["role"])
        c.Set("user_email", claims["email"])

        c.Next()
    }
}

func validateToken(tokenString string) (jwt.MapClaims, error) {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        return nil, fmt.Errorf("JWT_SECRET not set")
    }

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(secret), nil
    })

    if err != nil {
        if ve, ok := err.(*jwt.ValidationError); ok {
            if ve.Errors&jwt.ValidationErrorExpired != 0 {
                return nil, fmt.Errorf("token has expired")
            }
        }
        return nil, err
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token claims")
    }

    // Check if it's a refresh token
    if tokenType, ok := claims["type"].(string); ok && tokenType == "refresh" {
        // For refresh tokens, we only validate the user_id and type
        if claims["user_id"] == nil {
            return nil, fmt.Errorf("invalid refresh token")
        }
    } else {
        // For access tokens, we validate all required claims
        requiredClaims := []string{"user_id", "email", "username", "user_type", "role"}
        for _, claim := range requiredClaims {
            if claims[claim] == nil {
                return nil, fmt.Errorf("missing required claim: %s", claim)
            }
        }
    }

    return claims, nil
}




