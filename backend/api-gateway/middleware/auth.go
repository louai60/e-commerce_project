package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    // "github.com/golang-jwt/jwt"
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
        // TODO: Implement proper JWT validation
        if !isValidToken(token) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        c.Next()
    }
}

func isValidToken(token string) bool {
    // TODO: Implement proper token validation
    return true
}