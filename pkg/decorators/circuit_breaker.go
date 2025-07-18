package decorators

import (
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreakerImpl implements the circuit breaker pattern
type CircuitBreakerImpl struct {
	state           CircuitBreakerState
	failureCount    int
	lastFailureTime time.Time
	lastSuccessTime time.Time

	// Configuration
	failureThreshold int
	recoveryTimeout  time.Duration

	mu sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, recoveryTimeout time.Duration) *CircuitBreakerImpl {
	return &CircuitBreakerImpl{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		recoveryTimeout:  recoveryTimeout,
	}
}

// IsOpen checks if the circuit breaker is open
func (cb *CircuitBreakerImpl) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateOpen:
		// Check if recovery timeout has passed
		if time.Since(cb.lastFailureTime) >= cb.recoveryTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.mu.Unlock()
			cb.mu.RLock()
		}
		return cb.state == StateOpen

	case StateHalfOpen:
		return false // Allow one request to test

	case StateClosed:
		return false

	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreakerImpl) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastSuccessTime = time.Now()

	switch cb.state {
	case StateHalfOpen:
		// Success in half-open state, close the circuit
		cb.state = StateClosed
		cb.failureCount = 0
	case StateClosed:
		// Already closed, just update success time
		cb.failureCount = 0
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreakerImpl) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()
	cb.failureCount++

	switch cb.state {
	case StateClosed:
		// Check if threshold reached
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		// Failure in half-open state, open the circuit
		cb.state = StateOpen
	}
}

// GetState returns the current state as a string
func (cb *CircuitBreakerImpl) GetState() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreakerImpl) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.GetState(),
		"failure_count":     cb.failureCount,
		"last_failure":      cb.lastFailureTime,
		"last_success":      cb.lastSuccessTime,
		"failure_threshold": cb.failureThreshold,
		"recovery_timeout":  cb.recoveryTimeout.String(),
	}
}

// createCircuitBreaker creates a circuit breaker from configuration
func createCircuitBreaker(config *ProxyConfig) CircuitBreaker {
	failureThreshold := config.FailureThreshold
	if failureThreshold == 0 {
		failureThreshold = DefaultFailureThreshold
	}

	recoveryTimeout, _ := time.ParseDuration(config.CircuitBreaker)
	if recoveryTimeout == 0 {
		recoveryTimeout = 30 * time.Second
	}

	return NewCircuitBreaker(failureThreshold, recoveryTimeout)
}
