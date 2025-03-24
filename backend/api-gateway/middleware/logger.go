package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

func Logger(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        if query != "" {
            path = path + "?" + query
        }

        logger.Info("Request",
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.Int("status", status),
            zap.Duration("latency", latency),
            zap.String("ip", c.ClientIP()),
        )
    }
}