package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func AdminRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get user role from context (set during auth middleware)
        role, exists := c.Get("user_role")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }

        // Check if user is admin
        if role.(string) != "admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
            c.Abort()
            return
        }

        c.Next()
    }
}