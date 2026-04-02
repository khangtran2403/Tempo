package connector

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState trạng thái của circuit breaker
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Bình thường
	CircuitOpen                         // Tạm dừng vì quá nhiều lỗi
	CircuitHalfOpen                     // Thử lại
)

type CircuitBreaker struct {
	mu              sync.Mutex
	state           CircuitState
	failures        int
	lastFailureTime time.Time
	maxFailures     int
	timeout         time.Duration
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       CircuitClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	cb.mu.Lock()

	switch cb.state {
	case CircuitOpen:

		if time.Since(cb.lastFailureTime) > cb.timeout {

			cb.state = CircuitHalfOpen
			cb.mu.Unlock()
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker is open")
		}

	case CircuitHalfOpen:
		cb.mu.Unlock()

	case CircuitClosed:
		cb.mu.Unlock()
	}

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {

		cb.failures++
		cb.lastFailureTime = time.Now()

		if cb.failures >= cb.maxFailures {

			cb.state = CircuitOpen
			return fmt.Errorf("circuit breaker opened: %w", err)
		}

		return err
	}

	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
	}
	cb.failures = 0

	return nil
}

// GetState trả về trạng thái hiện tại
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset đặt lại circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.failures = 0
}
