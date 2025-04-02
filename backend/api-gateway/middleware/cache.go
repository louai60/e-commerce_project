package middleware

import (
    "bytes"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "net/http"
    "sort"
    "strings"
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

// CacheConfig holds configuration for cache middleware
type CacheConfig struct {
    TTL            time.Duration
    ExcludeHeaders []string
    IgnoreParams   []string
}

func CacheMiddleware(cacheManager *cache.CacheManager, config CacheConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip caching for non-GET requests
        if c.Request.Method != "GET" {
            c.Next()
            return
        }

        // Generate cache key
        key := generateCacheKey(c.Request, config)

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

        // Only cache successful responses
        if c.Writer.Status() == http.StatusOK {
            // Set cache with configured TTL
            cacheManager.Set(c.Request.Context(), key, writer.body.Bytes(), config.TTL)
        }
    }
}

func generateCacheKey(r *http.Request, config CacheConfig) string {
    // Start with base components
    components := []string{
        r.Host,
        r.Method,
        r.URL.Path,
    }

    // Add sorted query parameters (excluding ignored ones)
    if len(r.URL.Query()) > 0 {
        params := []string{}
        for key, values := range r.URL.Query() {
            if !contains(config.IgnoreParams, key) {
                sort.Strings(values)
                params = append(params, fmt.Sprintf("%s=%s", key, strings.Join(values, ",")))
            }
        }
        sort.Strings(params)
        components = append(components, strings.Join(params, "&"))
    }

    // Add relevant headers
    relevantHeaders := []string{}
    for key, values := range r.Header {
        if !contains(config.ExcludeHeaders, key) {
            sort.Strings(values)
            relevantHeaders = append(relevantHeaders, fmt.Sprintf("%s=%s", key, strings.Join(values, ",")))
        }
    }
    if len(relevantHeaders) > 0 {
        sort.Strings(relevantHeaders)
        components = append(components, strings.Join(relevantHeaders, "&"))
    }

    // Generate SHA-256 hash of the combined string
    hasher := sha256.New()
    hasher.Write([]byte(strings.Join(components, "|")))
    return "cache:" + hex.EncodeToString(hasher.Sum(nil))
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if strings.EqualFold(s, item) {
            return true
        }
    }
    return false
}

// CacheInvalidator invalidates cache entries based on patterns
func CacheInvalidator(cacheManager *cache.CacheManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // Only invalidate cache on successful write operations
        if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
            switch c.Request.Method {
            case "POST", "PUT", "PATCH", "DELETE":
                pattern := fmt.Sprintf("cache:%s:*", c.Request.URL.Path)
                cacheManager.Clear(c.Request.Context(), pattern)
            }
        }
    }
}
