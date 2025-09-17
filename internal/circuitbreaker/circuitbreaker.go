package circuitbreaker

import (
	"context"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreaker wraps gobreaker.CircuitBreaker with context support
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

// New creates a new CircuitBreaker with the given name and settings
func New(name string, st gobreaker.Settings) *CircuitBreaker {
	return &CircuitBreaker{
		cb: gobreaker.NewCircuitBreaker(st),
	}
}

// Execute runs the given function with circuit breaker protection
func (c *CircuitBreaker) Execute(ctx context.Context, req func() (interface{}, error)) (interface{}, error) {
	// Check if context is already canceled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		res, err := c.cb.Execute(func() (interface{}, error) {
			return req()
		})
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- res
	}()

	select {
	case res := <-resultChan:
		return res, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// DefaultSettings returns sensible default circuit breaker settings
func DefaultSettings(name string) gobreaker.Settings {
	return gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Log state changes for monitoring
		},
	}
}
