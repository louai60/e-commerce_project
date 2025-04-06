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
        // Handle specific validation errors
        if ve, ok := err.(*jwt.ValidationError); ok {
            if ve.Errors&jwt.ValidationErrorMalformed != 0 {
                return nil, fmt.Errorf("malformed token")
            } else if ve.Errors&jwt.ValidationErrorExpired != 0 {
                // Token is expired
                return nil, fmt.Errorf("token has expired")
            } else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
                // Token not active yet
                return nil, fmt.Errorf("token not active yet")
            } else {
                return nil, fmt.Errorf("couldn't handle this token: %w", err)
            }
        }
        // Other parsing errors
        return nil, fmt.Errorf("couldn't parse token: %w", err)
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token or claims")
    }

    // Ensure it's an access token
    tokenType, ok := claims["type"].(string)
    if !ok || tokenType != "access" {
        return nil, fmt.Errorf("invalid token type: expected 'access'")
    }

    // Validate required claims for an access token
    requiredClaims := []string{"user_id", "email", "username", "user_type", "role"}
    for _, claim := range requiredClaims {
        if claims[claim] == nil {
            return nil, fmt.Errorf("missing required claim: %s", claim)
        }
    }

    // Check claim types (optional but recommended for robustness)
    if _, ok := claims["user_id"].(float64); !ok { // JWT numbers are often float64
         if _, ok := claims["user_id"].(int64); !ok { // Allow int64 as well
            return nil, fmt.Errorf("invalid type for user_id claim")
         }
    }
    if _, ok := claims["role"].(string); !ok {
        return nil, fmt.Errorf("invalid type for role claim")
    }
     if _, ok := claims["email"].(string); !ok {
        return nil, fmt.Errorf("invalid type for email claim")
    }


    return claims, nil
}
