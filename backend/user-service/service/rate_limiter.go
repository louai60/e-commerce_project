package service

import (
    "sync"
    "time"
    "errors"
)

type SimpleRateLimiter struct {
    attempts map[string][]time.Time
    mu       sync.RWMutex
    maxAttempts int
    window     time.Duration
}

func NewSimpleRateLimiter(maxAttempts int, window time.Duration) *SimpleRateLimiter {
    return &SimpleRateLimiter{
        attempts:    make(map[string][]time.Time),
        maxAttempts: maxAttempts,
        window:     window,
    }
}

func (rl *SimpleRateLimiter) Allow(key string) error {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    windowStart := now.Add(-rl.window)

    // Clean old attempts
    attempts := rl.attempts[key]
    valid := attempts[:0]
    for _, t := range attempts {
        if t.After(windowStart) {
            valid = append(valid, t)
        }
    }
    rl.attempts[key] = valid

    if len(valid) >= rl.maxAttempts {
        return errors.New("rate limit exceeded")
    }

    return nil
}

func (rl *SimpleRateLimiter) Record(key string) {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    rl.attempts[key] = append(rl.attempts[key], time.Now())
}