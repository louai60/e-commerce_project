package middleware

import (
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "time"
)

func CORSMiddleware() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins: []string{
            "http://localhost:3000",
            "https://yourdomain.com",
            "https://api.yourdomain.com",
            "https://admin.yourdomain.com",
        },
        AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders: []string{
            "Origin",
            "Content-Type",
            "Accept",
            "Authorization",
            "X-Requested-With",
        },
        AllowCredentials: true,
        MaxAge: 12 * time.Hour,
    })
}
