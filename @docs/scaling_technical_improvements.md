# Technical Improvements for Scaling to 100K+ Users

This document outlines critical technical improvements to ensure our NextJS + Go microservices architecture can scale efficiently to handle hundreds of thousands of users without becoming expensive or problematic.

## 1. Optimize Caching Strategy

### Current Issues
- Redis cache implementation has a fixed 15-minute TTL
- Cache invalidation issues (as seen in the product deletion flow)
- Client-side SWR caching isn't fully optimized

### Recommendations

#### Implement Tiered Caching
```go
// In your cache manager
func (c *CacheManager) Get(ctx context.Context, key string, result interface{}) error {
    // Try L1 cache first (memory)
    if c.memoryCache.Has(key) {
        return c.memoryCache.Get(key, result)
    }
    
    // Then try Redis
    if err := c.redisCache.Get(ctx, key, result); err == nil {
        // Populate memory cache for future requests
        c.memoryCache.Set(key, result, 30*time.Second)
        return nil
    }
    
    return cache.ErrCacheMiss
}
```

#### Dynamic TTL Based on Data Type
```go
// Different TTLs for different data types
var cacheTTLs = map[string]time.Duration{
    "product_detail": 1 * time.Hour,
    "product_list":   5 * time.Minute,
    "category_list":  24 * time.Hour,
    "user_data":      15 * time.Minute,
}
```

#### Implement Cache Stampede Protection
```go
// Use a mutex to prevent multiple identical requests
var cacheMutexes = sync.Map{}

func (c *CacheManager) GetOrSet(ctx context.Context, key string, ttl time.Duration, fetchFn func() (interface{}, error)) (interface{}, error) {
    // Try to get from cache first
    var result interface{}
    if err := c.Get(ctx, key, &result); err == nil {
        return result, nil
    }
    
    // Lock to prevent multiple fetches
    mutexI, _ := cacheMutexes.LoadOrStore(key, &sync.Mutex{})
    mutex := mutexI.(*sync.Mutex)
    mutex.Lock()
    defer mutex.Unlock()
    
    // Check cache again after acquiring lock
    if err := c.Get(ctx, key, &result); err == nil {
        return result, nil
    }
    
    // Fetch data
    data, err := fetchFn()
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    c.Set(ctx, key, data, ttl)
    return data, nil
}
```

## 2. Database Optimization

### Current Issues
- Soft deletes without proper indexing can slow queries as data grows
- No evidence of database sharding strategy
- Potential N+1 query issues in product listing with related data

### Recommendations

#### Add Composite Indexes for Soft Deletes
```sql
-- Add to your migration files
CREATE INDEX idx_products_deleted_at_created_at ON products (deleted_at, created_at);
CREATE INDEX idx_products_category_deleted_at ON products (category_id, deleted_at);
```

#### Implement Read Replicas Configuration
```go
// In your database configuration
type DBConfig struct {
    Master  *sql.DB
    Replicas []*sql.DB
    ReplicaSelector func([]*sql.DB) *sql.DB
}

func NewRepository(config DBConfig) *PostgresRepository {
    return &PostgresRepository{
        master:   config.Master,
        replicas: config.Replicas,
        selector: config.ReplicaSelector,
    }
}

func (r *PostgresRepository) GetProduct(ctx context.Context, id string) (*models.Product, error) {
    // Use replica for reads
    db := r.selector(r.replicas)
    // Fall back to master if needed
    if db == nil {
        db = r.master
    }
    // Execute query...
}
```

#### Prepare for Horizontal Partitioning
```go
// Add tenant/shard ID to your models
type Product struct {
    ID        string
    ShardKey  string // Could be tenant ID, category, or date-based
    // Other fields...
}

// Sharding router
func (r *ShardedRepository) getDBForProduct(shardKey string) *sql.DB {
    shardID := r.shardingStrategy.GetShardID(shardKey)
    return r.shardConnections[shardID]
}
```

## 3. API Design and Performance

### Current Issues
- Authentication token handling is causing errors (as seen in product deletion)
- API responses aren't consistently structured
- No rate limiting or throttling mechanisms

### Recommendations

#### Standardize API Response Structure
```go
// In your API handlers
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

type Meta struct {
    Page       int `json:"page,omitempty"`
    PerPage    int `json:"per_page,omitempty"`
    TotalItems int `json:"total_items,omitempty"`
    TotalPages int `json:"total_pages,omitempty"`
}

func RespondWithJSON(c *gin.Context, statusCode int, response APIResponse) {
    c.JSON(statusCode, response)
}
```

#### Implement API Versioning
```go
// In your routes setup
v1 := router.Group("/api/v1")
v2 := router.Group("/api/v2")

// Ensure backward compatibility
v2.GET("/products/:id", handlers.GetProductV2)
v1.GET("/products/:id", handlers.GetProductV1)
```

#### Add Rate Limiting Middleware
```go
func RateLimiter() gin.HandlerFunc {
    store := memory.NewStore()
    rate := limiter.Rate{
        Period: 1 * time.Minute,
        Limit:  100,
    }
    
    return mgin.NewMiddleware(limiter.New(store, rate))
}

// Apply to routes
api := router.Group("/api")
api.Use(RateLimiter())
```

## 4. Frontend Optimization

### Current Issues
- SWR configuration isn't optimized for high-traffic scenarios
- No code splitting or lazy loading for admin dashboard components
- Duplicate API calls (as seen in product deletion)

### Recommendations

#### Optimize SWR Configuration
```typescript
// In your SWR config
export const swrConfig: SWRConfiguration = {
  dedupingInterval: 5000,
  focusThrottleInterval: 5000,
  loadingTimeout: 3000,
  errorRetryCount: 3,
  errorRetryInterval: 5000,
  revalidateOnFocus: false,
  revalidateIfStale: true,
  revalidateOnReconnect: true,
  shouldRetryOnError: (err) => {
    // Only retry on network errors, not 4xx/5xx
    return !err.status || err.status < 400;
  },
  onErrorRetry: (error, key, config, revalidate, { retryCount }) => {
    // Custom retry logic with exponential backoff
    if (retryCount >= 3) return;
    setTimeout(() => revalidate({ retryCount }), 2 ** retryCount * 1000);
  }
};
```

#### Implement Code Splitting
```tsx
// In your page components
import dynamic from 'next/dynamic';

// Lazy load heavy components
const ProductTable = dynamic(() => import('@/components/products/ProductTable'), {
  loading: () => <TableSkeleton />,
  ssr: false // If not needed for SEO
});

const ProductForm = dynamic(() => import('@/components/products/ProductForm'), {
  loading: () => <FormSkeleton />,
  ssr: false
});
```

#### Add Request Deduplication
```typescript
// In your API client
const pendingRequests = new Map();

api.interceptors.request.use(
  (config) => {
    const requestKey = `${config.method}:${config.url}:${JSON.stringify(config.params)}`;
    
    if (pendingRequests.has(requestKey)) {
      const controller = new AbortController();
      config.signal = controller.signal;
      controller.abort('Duplicate request canceled');
    }
    
    pendingRequests.set(requestKey, true);
    
    return config;
  },
  (error) => Promise.reject(error)
);

api.interceptors.response.use(
  (response) => {
    const requestKey = `${response.config.method}:${response.config.url}:${JSON.stringify(response.config.params)}`;
    pendingRequests.delete(requestKey);
    return response;
  },
  (error) => {
    if (error.config) {
      const requestKey = `${error.config.method}:${error.config.url}:${JSON.stringify(error.config.params)}`;
      pendingRequests.delete(requestKey);
    }
    return Promise.reject(error);
  }
);
```

## 5. Authentication and Security

### Current Issues
- JWT token handling is causing issues in API requests
- No refresh token rotation mechanism
- Authentication middleware lacks proper error handling

### Recommendations

#### Implement Token Refresh Strategy
```typescript
// In your API client
let isRefreshing = false;
let refreshSubscribers = [];

const onRefreshed = (token) => {
  refreshSubscribers.forEach(callback => callback(token));
  refreshSubscribers = [];
};

const addRefreshSubscriber = (callback) => {
  refreshSubscribers.push(callback);
};

api.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;
    
    // If error is not 401 or request already retried, reject
    if (error.response?.status !== 401 || originalRequest._retry) {
      return Promise.reject(error);
    }
    
    if (isRefreshing) {
      // Wait for token refresh
      return new Promise((resolve, reject) => {
        addRefreshSubscriber(token => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          resolve(api(originalRequest));
        });
      });
    }
    
    originalRequest._retry = true;
    isRefreshing = true;
    
    try {
      const { data } = await api.post('/auth/refresh');
      const { access_token } = data;
      
      // Update token in storage
      localStorage.setItem('access_token', access_token);
      
      // Update header for current request
      originalRequest.headers.Authorization = `Bearer ${access_token}`;
      
      // Notify waiting requests
      onRefreshed(access_token);
      
      return api(originalRequest);
    } catch (refreshError) {
      // Force logout on refresh failure
      localStorage.removeItem('access_token');
      window.location.href = '/login';
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  }
);
```

#### Enhance Auth Middleware
```go
// In your auth middleware
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error": &APIError{
                    Code:    "auth_required",
                    Message: "Authentication required",
                },
            })
            return
        }
        
        claims, err := validateToken(token)
        if err != nil {
            // Detailed error responses
            if errors.Is(err, ErrTokenExpired) {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                    "success": false,
                    "error": &APIError{
                        Code:    "token_expired",
                        Message: "Authentication token has expired",
                    },
                })
                return
            }
            
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error": &APIError{
                    Code:    "invalid_token",
                    Message: "Invalid authentication token",
                },
            })
            return
        }
        
        // Set user info in context
        c.Set("user_id", claims["user_id"])
        c.Set("user_role", claims["role"])
        
        c.Next()
    }
}
```

## 6. Monitoring and Observability

### Current Issues
- Limited error tracking in backend services
- No performance monitoring for frontend
- No centralized logging system

### Recommendations

#### Implement Structured Logging
```go
// In your service layer
type LoggedService struct {
    service ProductService
    logger  *zap.Logger
    tracer  trace.Tracer
}

func (s *LoggedService) GetProduct(ctx context.Context, id string) (product *models.Product, err error) {
    ctx, span := s.tracer.Start(ctx, "GetProduct")
    defer span.End()
    
    defer func(start time.Time) {
        latency := time.Since(start)
        fields := []zap.Field{
            zap.String("method", "GetProduct"),
            zap.String("product_id", id),
            zap.Duration("latency", latency),
        }
        
        if err != nil {
            fields = append(fields, zap.Error(err))
            span.SetStatus(codes.Error, err.Error())
            s.logger.Error("failed to get product", fields...)
        } else {
            s.logger.Info("product retrieved", fields...)
        }
    }(time.Now())
    
    return s.service.GetProduct(ctx, id)
}
```

#### Add Frontend Performance Monitoring
```typescript
// In your _app.tsx
import { init as initApm } from '@elastic/apm-rum';

const apm = initApm({
  serviceName: 'admin-dashboard',
  serverUrl: process.env.NEXT_PUBLIC_APM_SERVER_URL,
  environment: process.env.NODE_ENV,
  distributedTracingOrigins: [process.env.NEXT_PUBLIC_API_URL],
});

// Custom hook for component performance
function useComponentPerformance(componentName) {
  useEffect(() => {
    const transaction = apm.startTransaction(`render:${componentName}`, 'component');
    
    return () => {
      transaction.end();
    };
  }, [componentName]);
}
```

#### Implement Health Checks
```go
// In your API server
func setupHealthChecks(router *gin.Engine) {
    health := router.Group("/health")
    
    // Basic liveness probe
    health.GET("/live", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "up"})
    })
    
    // Readiness probe with dependency checks
    health.GET("/ready", func(c *gin.Context) {
        status := http.StatusOK
        checks := map[string]string{
            "database": "up",
            "redis":    "up",
        }
        
        // Check database
        if err := checkDBConnection(); err != nil {
            status = http.StatusServiceUnavailable
            checks["database"] = "down"
        }
        
        // Check Redis
        if err := checkRedisConnection(); err != nil {
            status = http.StatusServiceUnavailable
            checks["redis"] = "down"
        }
        
        c.JSON(status, gin.H{
            "status": status == http.StatusOK ? "ready" : "not ready",
            "checks": checks,
        })
    })
}
```

## 7. Infrastructure Preparation

### Current Issues
- No clear containerization strategy
- No evidence of infrastructure-as-code
- Limited configuration for different environments

### Recommendations

#### Containerize All Services
```dockerfile
# Example Dockerfile for product service
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o product-service ./cmd/product-service

FROM alpine:3.16
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/product-service .
COPY --from=builder /app/configs ./configs

# Use environment variables for configuration
ENV DB_HOST=postgres \
    DB_PORT=5432 \
    REDIS_HOST=redis \
    REDIS_PORT=6379 \
    LOG_LEVEL=info

EXPOSE 8080
CMD ["./product-service"]
```

#### Implement Infrastructure as Code
```yaml
# docker-compose.yml for local development
version: '3.8'

services:
  api-gateway:
    build:
      context: ./backend/api-gateway
    ports:
      - "8080:8080"
    environment:
      - PRODUCT_SERVICE_HOST=product-service
      - PRODUCT_SERVICE_PORT=8081
      - USER_SERVICE_HOST=user-service
      - USER_SERVICE_PORT=8082
    depends_on:
      - product-service
      - user-service

  product-service:
    build:
      context: ./backend/product-service
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  # Other services...

  postgres:
    image: postgres:14-alpine
    environment:
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_USER=postgres
    volumes:
      - postgres-data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis-data:/data

volumes:
  postgres-data:
  redis-data:
```

#### Environment Configuration
```go
// In your config package
type Config struct {
    Environment string `envconfig:"ENVIRONMENT" default:"development"`
    
    Database struct {
        Host     string `envconfig:"DB_HOST" default:"localhost"`
        Port     int    `envconfig:"DB_PORT" default:"5432"`
        User     string `envconfig:"DB_USER" default:"postgres"`
        Password string `envconfig:"DB_PASSWORD" default:"postgres"`
        Name     string `envconfig:"DB_NAME" default:"ecommerce"`
        SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
        
        // Scale settings
        MaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
        MaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"25"`
        ConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"15m"`
    }
    
    Redis struct {
        Host     string `envconfig:"REDIS_HOST" default:"localhost"`
        Port     int    `envconfig:"REDIS_PORT" default:"6379"`
        Password string `envconfig:"REDIS_PASSWORD" default:""`
        DB       int    `envconfig:"REDIS_DB" default:"0"`
        
        // Scale settings
        PoolSize     int           `envconfig:"REDIS_POOL_SIZE" default:"100"`
        MinIdleConns int           `envconfig:"REDIS_MIN_IDLE_CONNS" default:"10"`
        MaxConnAge   time.Duration `envconfig:"REDIS_MAX_CONN_AGE" default:"30m"`
    }
    
    // Other configuration sections...
}
```

## Implementation Priority

To ensure a smooth scaling journey, implement these improvements in the following order:

1. **Database Optimization** - This forms the foundation of your system's performance
2. **Caching Strategy** - Reduces database load and improves response times
3. **API Design and Performance** - Ensures consistent client-server communication
4. **Authentication and Security** - Critical for system integrity at scale
5. **Frontend Optimization** - Improves user experience as traffic grows
6. **Monitoring and Observability** - Provides visibility into system performance
7. **Infrastructure Preparation** - Enables horizontal scaling when needed

By implementing these technical improvements, your NextJS + Go architecture will be well-positioned to handle hundreds of thousands of users efficiently and cost-effectively.
