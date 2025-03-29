package middleware

import (
    "os"
    "github.com/gin-gonic/gin"
)

func AdminKeyRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        adminKey := c.GetHeader("X-Admin-Key")
        if adminKey != os.Getenv("ADMIN_CREATE_KEY") {
            c.JSON(401, gin.H{"error": "Invalid admin key"})
            c.Abort()
            return
        }
        c.Next()
    }
}