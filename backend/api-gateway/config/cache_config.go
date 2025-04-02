package config

import "time"

type CacheConfig struct {
    // Default TTL for cached items
    DefaultTTL time.Duration `yaml:"defaultTTL"`

    // Headers to exclude from cache key generation
    ExcludeHeaders []string `yaml:"excludeHeaders"`

    // Query parameters to ignore in cache key generation
    IgnoreParams []string `yaml:"ignoreParams"`

    // Paths to exclude from caching
    ExcludePaths []string `yaml:"excludePaths"`

    // Maximum size of cached items in bytes
    MaxItemSize int `yaml:"maxItemSize"`

    // Redis configuration
    Redis struct {
        Host     string `yaml:"host"`
        Port     string `yaml:"port"`
        Password string `yaml:"password"`
        DB       int    `yaml:"db"`
    } `yaml:"redis"`
}

func GetDefaultCacheConfig() CacheConfig {
    return CacheConfig{
        DefaultTTL: 15 * time.Minute,
        ExcludeHeaders: []string{
            "Authorization",
            "Cookie",
            "Set-Cookie",
        },
        IgnoreParams: []string{
            "token",
            "api_key",
        },
        ExcludePaths: []string{
            "/api/v1/auth",
            "/api/v1/webhook",
        },
        MaxItemSize: 1 << 20, // 1MB
    }
}