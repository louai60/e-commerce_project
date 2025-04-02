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
            "http://127.0.0.1:3000",
        },
        AllowMethods: []string{
            "GET",
            "POST",
            "PUT",
            "PATCH",
            "DELETE",
            "OPTIONS",
        },
        AllowHeaders: []string{
            "Origin",
            "Content-Type",
            "Content-Length",
            "Accept",
            "Authorization",
            "X-Requested-With",
            "X-Admin-Key",
        },
        ExposeHeaders: []string{
            "Content-Length",
        },
        AllowCredentials: true,
        MaxAge: 12 * time.Hour,
    })
}
