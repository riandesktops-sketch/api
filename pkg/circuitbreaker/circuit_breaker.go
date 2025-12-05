package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
	
	// ErrTooManyRequests is returned when too many requests in half-open state
	ErrTooManyRequests = errors.New("too many requests")
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed - normal operation, requests pass through
	StateClosed State = iota
	
	// StateOpen - circuit is open, requests fail fast
	StateOpen
	
	// StateHalfOpen - testing if service recovered
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements circuit breaker pattern
// Prevents cascade failures by failing fast when service is down
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	halfOpenMaxReqs int
	
	mu              sync.RWMutex
	state           State
	failures        int
	lastFailureTime time.Time
	halfOpenReqs    int
}

// Config holds circuit breaker configuration
type Config struct {
	MaxFailures     int           // Number of failures before opening circuit
	ResetTimeout    time.Duration // Time to wait before attempting recovery
	HalfOpenMaxReqs int           // Max requests allowed in half-open state
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config Config) *CircuitBreaker {
	if config.MaxFailures <= 0 {
		config.MaxFailures = 5
	}
	if config.ResetTimeout <= 0 {
		config.ResetTimeout = 60 * time.Second
	}
	if config.HalfOpenMaxReqs <= 0 {
		config.HalfOpenMaxReqs = 1
	}
	
	return &CircuitBreaker{
		maxFailures:     config.MaxFailures,
		resetTimeout:    config.ResetTimeout,
		halfOpenMaxReqs: config.HalfOpenMaxReqs,
		state:           StateClosed,
	}
}

// Execute wraps a function call with circuit breaker logic
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if we can proceed
	if err := cb.beforeRequest(); err != nil {
		return err
	}
	
	// Execute the function
	err := fn()
	
	// Update state based on result
	cb.afterRequest(err)
	
	return err
}

// beforeRequest checks if request can proceed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	switch cb.state {
	case StateClosed:
		return nil
		
	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenReqs = 0
			return nil
		}
		return ErrCircuitOpen
		
	case StateHalfOpen:
		// Limit requests in half-open state
		if cb.halfOpenReqs >= cb.halfOpenMaxReqs {
			return ErrTooManyRequests
		}
		cb.halfOpenReqs++
		return nil
		
	default:
		return ErrCircuitOpen
	}
}

// afterRequest updates circuit breaker state after request
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if err != nil {
		// Request failed
		cb.failures++
		cb.lastFailureTime = time.Now()
		
		if cb.state == StateHalfOpen {
			// Failed in half-open, go back to open
			cb.state = StateOpen
		} else if cb.failures >= cb.maxFailures {
			// Too many failures, open the circuit
			cb.state = StateOpen
		}
	} else {
		// Request succeeded
		if cb.state == StateHalfOpen {
			// Success in half-open, close the circuit
			cb.state = StateClosed
			cb.failures = 0
			cb.halfOpenReqs = 0
		} else if cb.state == StateClosed {
			// Reset failure count on success
			cb.failures = 0
		}
	}
}

// State returns current circuit breaker state
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	return cb.state
}

// IsOpen returns true if circuit is open
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.State() == StateOpen
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.state = StateClosed
	cb.failures = 0
	cb.halfOpenReqs = 0
}

// Failures returns current failure count
func (cb *CircuitBreaker) Failures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	return cb.failures
}
