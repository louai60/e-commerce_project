package cache

import (
	"errors"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	// CircuitClosed means the circuit is closed and operations are allowed
	CircuitClosed CircuitState = iota
	// CircuitOpen means the circuit is open and operations are not allowed
	CircuitOpen
	// CircuitHalfOpen means the circuit is testing if operations can be allowed again
	CircuitHalfOpen
)

var (
	// ErrCircuitOpen is returned when the circuit is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	state                CircuitState
	failureThreshold     int64
	resetTimeout         time.Duration
	halfOpenSuccessThreshold int64
	failureCount         int64
	successCount         int64
	lastStateChangeTime  time.Time
	mutex                sync.RWMutex
}

// CircuitBreakerOptions defines options for creating a circuit breaker
type CircuitBreakerOptions struct {
	FailureThreshold     int64
	ResetTimeout         time.Duration
	HalfOpenSuccessThreshold int64
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(opts CircuitBreakerOptions) *CircuitBreaker {
	if opts.FailureThreshold <= 0 {
		opts.FailureThreshold = 5
	}
	if opts.ResetTimeout <= 0 {
		opts.ResetTimeout = 30 * time.Second
	}
	if opts.HalfOpenSuccessThreshold <= 0 {
		opts.HalfOpenSuccessThreshold = 2
	}
	
	return &CircuitBreaker{
		state:                CircuitClosed,
		failureThreshold:     opts.FailureThreshold,
		resetTimeout:         opts.ResetTimeout,
		halfOpenSuccessThreshold: opts.HalfOpenSuccessThreshold,
		lastStateChangeTime:  time.Now(),
	}
}

// Execute executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.AllowRequest() {
		return ErrCircuitOpen
	}
	
	err := fn()
	
	cb.RecordResult(err == nil)
	
	return err
}

// AllowRequest checks if a request should be allowed based on the circuit state
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if reset timeout has elapsed
		if time.Since(cb.lastStateChangeTime) > cb.resetTimeout {
			// Transition to half-open state
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			defer cb.mutex.Unlock()
			
			// Double-check state after acquiring write lock
			if cb.state == CircuitOpen && time.Since(cb.lastStateChangeTime) > cb.resetTimeout {
				cb.state = CircuitHalfOpen
				cb.lastStateChangeTime = time.Now()
				cb.successCount = 0
			}
			return cb.state == CircuitHalfOpen
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordResult records the result of an operation
func (cb *CircuitBreaker) RecordResult(success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	switch cb.state {
	case CircuitClosed:
		if !success {
			cb.failureCount++
			if cb.failureCount >= cb.failureThreshold {
				cb.state = CircuitOpen
				cb.lastStateChangeTime = time.Now()
			}
		} else {
			// Reset failure count after a successful operation
			cb.failureCount = 0
		}
	case CircuitHalfOpen:
		if !success {
			// Any failure in half-open state transitions back to open
			cb.state = CircuitOpen
			cb.lastStateChangeTime = time.Now()
			cb.failureCount = cb.failureThreshold
		} else {
			cb.successCount++
			if cb.successCount >= cb.halfOpenSuccessThreshold {
				// Enough successes, transition back to closed
				cb.state = CircuitClosed
				cb.lastStateChangeTime = time.Now()
				cb.failureCount = 0
			}
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.state = CircuitClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastStateChangeTime = time.Now()
}

// GetMetrics returns metrics about the circuit breaker
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	stateStr := "closed"
	switch cb.state {
	case CircuitOpen:
		stateStr = "open"
	case CircuitHalfOpen:
		stateStr = "half-open"
	}
	
	return map[string]interface{}{
		"state":                 stateStr,
		"failure_count":         cb.failureCount,
		"success_count":         cb.successCount,
		"time_in_state_seconds": time.Since(cb.lastStateChangeTime).Seconds(),
	}
}
