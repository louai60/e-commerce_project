package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// CacheMetrics tracks cache performance metrics
type CacheMetrics struct {
	hits              int64
	misses            int64
	errors            int64
	totalLatencyNs    int64
	operationCount    int64
	lastResetTime     time.Time
	slowOperationsMap sync.Map // Maps operation key to count
	mu                sync.RWMutex
}

// NewCacheMetrics creates a new metrics collector
func NewCacheMetrics() *CacheMetrics {
	return &CacheMetrics{
		lastResetTime: time.Now(),
	}
}

// RecordHit records a cache hit
func (m *CacheMetrics) RecordHit() {
	atomic.AddInt64(&m.hits, 1)
}

// RecordMiss records a cache miss
func (m *CacheMetrics) RecordMiss() {
	atomic.AddInt64(&m.misses, 1)
}

// RecordError records a cache error
func (m *CacheMetrics) RecordError() {
	atomic.AddInt64(&m.errors, 1)
}

// RecordLatency records operation latency
func (m *CacheMetrics) RecordLatency(latencyNs int64) {
	atomic.AddInt64(&m.totalLatencyNs, latencyNs)
	atomic.AddInt64(&m.operationCount, 1)
	
	// Record slow operations (>10ms)
	if latencyNs > 10_000_000 {
		m.recordSlowOperation()
	}
}

// recordSlowOperation increments the slow operation counter
func (m *CacheMetrics) recordSlowOperation() {
	// Get current time truncated to minute for bucketing
	now := time.Now().Truncate(time.Minute)
	key := now.Format(time.RFC3339)
	
	// Load or initialize counter
	val, _ := m.slowOperationsMap.LoadOrStore(key, int64(0))
	// Increment counter
	m.slowOperationsMap.Store(key, val.(int64)+1)
	
	// Clean up old entries (older than 1 hour)
	threshold := time.Now().Add(-1 * time.Hour).Truncate(time.Minute)
	m.slowOperationsMap.Range(func(k, v interface{}) bool {
		keyTime, err := time.Parse(time.RFC3339, k.(string))
		if err == nil && keyTime.Before(threshold) {
			m.slowOperationsMap.Delete(k)
		}
		return true
	})
}

// GetHitRate returns the cache hit rate as a percentage
func (m *CacheMetrics) GetHitRate() float64 {
	hits := atomic.LoadInt64(&m.hits)
	misses := atomic.LoadInt64(&m.misses)
	total := hits + misses
	
	if total == 0 {
		return 0
	}
	
	return float64(hits) / float64(total) * 100
}

// GetAverageLatencyMs returns the average operation latency in milliseconds
func (m *CacheMetrics) GetAverageLatencyMs() float64 {
	totalLatency := atomic.LoadInt64(&m.totalLatencyNs)
	count := atomic.LoadInt64(&m.operationCount)
	
	if count == 0 {
		return 0
	}
	
	// Convert nanoseconds to milliseconds
	return float64(totalLatency) / float64(count) / 1_000_000
}

// GetMetrics returns all metrics as a map
func (m *CacheMetrics) GetMetrics() map[string]interface{} {
	hits := atomic.LoadInt64(&m.hits)
	misses := atomic.LoadInt64(&m.misses)
	errors := atomic.LoadInt64(&m.errors)
	total := hits + misses
	
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}
	
	avgLatency := 0.0
	totalLatency := atomic.LoadInt64(&m.totalLatencyNs)
	count := atomic.LoadInt64(&m.operationCount)
	if count > 0 {
		avgLatency = float64(totalLatency) / float64(count) / 1_000_000 // ms
	}
	
	// Count slow operations in the last hour
	slowOpsCount := int64(0)
	m.slowOperationsMap.Range(func(_, v interface{}) bool {
		slowOpsCount += v.(int64)
		return true
	})
	
	m.mu.RLock()
	uptime := time.Since(m.lastResetTime).Seconds()
	m.mu.RUnlock()
	
	return map[string]interface{}{
		"hits":              hits,
		"misses":            misses,
		"errors":            errors,
		"hit_rate":          hitRate,
		"avg_latency_ms":    avgLatency,
		"operation_count":   count,
		"slow_ops_count":    slowOpsCount,
		"uptime_seconds":    uptime,
		"operations_per_sec": float64(count) / uptime,
	}
}

// Reset resets all metrics
func (m *CacheMetrics) Reset() {
	atomic.StoreInt64(&m.hits, 0)
	atomic.StoreInt64(&m.misses, 0)
	atomic.StoreInt64(&m.errors, 0)
	atomic.StoreInt64(&m.totalLatencyNs, 0)
	atomic.StoreInt64(&m.operationCount, 0)
	
	m.mu.Lock()
	m.lastResetTime = time.Now()
	m.mu.Unlock()
	
	// Clear slow operations map
	m.slowOperationsMap = sync.Map{}
}
