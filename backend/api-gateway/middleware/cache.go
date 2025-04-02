package middleware

import (
    "bytes"
    "fmt"
    // "io"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/louai60/e-commerce_project/backend/shared/cache"
)

type CachedWriter struct {
    gin.ResponseWriter
    body *bytes.Buffer
}

func (w *CachedWriter) Write(b []byte) (int, error) {
    w.body.Write(b)
    return w.ResponseWriter.Write(b)
}

func CacheMiddleware(cacheManager *cache.CacheManager, ttl time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip caching for non-GET requests
        if c.Request.Method != "GET" {
            c.Next()
            return
        }

        // Generate cache key
        key := generateCacheKey(c.Request)

        // Try to get from cache
        var cachedResponse []byte
        err := cacheManager.Get(c.Request.Context(), key, &cachedResponse)
        if err == nil {
            c.Data(http.StatusOK, "application/json", cachedResponse)
            c.Abort()
            return
        }

        // Cache miss, wrap response writer
        writer := &CachedWriter{
            ResponseWriter: c.Writer,
            body:          &bytes.Buffer{},
        }
        c.Writer = writer

        c.Next()

        // Cache the response if status is 200
        if c.Writer.Status() == http.StatusOK {
            cacheManager.Set(c.Request.Context(), key, writer.body.Bytes(), ttl)
        }
    }
}

func generateCacheKey(r *http.Request) string {
    return fmt.Sprintf("%s:%s:%s", r.Host, r.Method, r.URL.Path)
}