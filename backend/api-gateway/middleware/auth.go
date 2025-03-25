package middleware

import (
    "net/http"
    "strings"

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
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
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
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Make sure to use the same secret key as in user-service
        return []byte("your_jwt_secret_key_here"), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return claims, nil
    }

    return nil, err
}

