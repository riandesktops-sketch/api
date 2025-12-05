package ratelimiter

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter implements token bucket algorithm for rate limiting
// Thread-safe for concurrent access
type RateLimiter struct {
	limiter     *rate.Limiter
	maxRequests int
	interval    time.Duration
	mu          sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
// maxRequests: maximum number of requests allowed per interval
// interval: time window for rate limiting (e.g., 1 second)
func NewRateLimiter(maxRequests int, interval time.Duration) *RateLimiter {
	// Calculate rate per second
	rps := float64(maxRequests) / interval.Seconds()
	
	return &RateLimiter{
		limiter:     rate.NewLimiter(rate.Limit(rps), maxRequests),
		maxRequests: maxRequests,
		interval:    interval,
	}
}

// Allow checks if a request is allowed without blocking
// Returns true if request can proceed, false otherwise
func (rl *RateLimiter) Allow() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	return rl.limiter.Allow()
}

// Wait blocks until the request can proceed or context is cancelled
// Returns error if context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.RLock()
	limiter := rl.limiter
	rl.mu.RUnlock()
	
	return limiter.Wait(ctx)
}

// WaitN blocks until n requests can proceed or context is cancelled
func (rl *RateLimiter) WaitN(ctx context.Context, n int) error {
	rl.mu.RLock()
	limiter := rl.limiter
	rl.mu.RUnlock()
	
	return limiter.WaitN(ctx, n)
}

// Reserve reserves a request and returns a Reservation
// The caller must call Cancel on the returned Reservation if the request is not used
func (rl *RateLimiter) Reserve() *rate.Reservation {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	return rl.limiter.Reserve()
}

// SetLimit dynamically updates the rate limit
// Useful for adjusting limits based on API tier or runtime conditions
func (rl *RateLimiter) SetLimit(maxRequests int, interval time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rps := float64(maxRequests) / interval.Seconds()
	rl.limiter.SetLimit(rate.Limit(rps))
	rl.limiter.SetBurst(maxRequests)
	rl.maxRequests = maxRequests
	rl.interval = interval
}

// GetLimit returns current rate limit configuration
func (rl *RateLimiter) GetLimit() (maxRequests int, interval time.Duration) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	return rl.maxRequests, rl.interval
}

// Tokens returns the number of available tokens
func (rl *RateLimiter) Tokens() float64 {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	return rl.limiter.Tokens()
}
