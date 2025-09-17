package circuitbreaker

import (
	"log"
	"sync"
	"time"
)

type CircuitBreaker struct {
	state        string // "closed", "open", "half-open"
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
	mu           sync.Mutex
}

func New(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        "closed",
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if cb.state == "open" {
		if now.Sub(cb.lastFailure) < cb.resetTimeout {
			return false
		}
		// Timeout expired, transition to half-open
		cb.state = "half-open"
		log.Printf("Circuit breaker transitioning to half-open state")
	}
	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "half-open" {
		log.Printf("Circuit breaker test request succeeded, closing circuit")
	}
	cb.state = "closed"
	cb.failures = 0
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	cb.failures++
	cb.lastFailure = now

	if cb.state == "half-open" {
		// Immediately open the circuit on failure in half-open state
		log.Printf("Circuit breaker test request failed, reopening circuit")
		cb.state = "open"
	} else if cb.failures >= cb.maxFailures {
		log.Printf("Circuit breaker opened due to %d consecutive failures", cb.failures)
		cb.state = "open"
	}
}
