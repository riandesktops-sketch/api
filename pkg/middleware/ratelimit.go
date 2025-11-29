package middleware

import (
	"sync"
	"time"
	"zodiac-ai-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter implements sliding window rate limiting
// Reference: DDIA Ch. 11 - Sliding window provides fair rate limiting with O(1) complexity
type RateLimiter struct {
	requests map[string]*userRequests
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type userRequests struct {
	timestamps []time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*userRequests),
		limit:    limit,
		window:   window,
	}

	// Cleanup goroutine to prevent memory leaks
	go rl.cleanup()

	return rl
}

// RateLimitMiddleware creates rate limiting middleware
func (rl *RateLimiter) RateLimitMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user identifier (IP or user_id if authenticated)
		identifier := c.IP()
		if userID := GetUserID(c); userID != "" {
			identifier = userID
		}

		// Check rate limit
		if !rl.allow(identifier) {
			return response.TooManyRequests(c, "Rate limit exceeded. Please try again later.")
		}

		return c.Next()
	}
}

// allow checks if request is allowed based on rate limit
func (rl *RateLimiter) allow(identifier string) bool {
	rl.mu.Lock()
	userReq, exists := rl.requests[identifier]
	if !exists {
		userReq = &userRequests{
			timestamps: make([]time.Time, 0),
		}
		rl.requests[identifier] = userReq
	}
	rl.mu.Unlock()

	userReq.mu.Lock()
	defer userReq.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Remove timestamps outside the window (sliding window)
	validTimestamps := make([]time.Time, 0)
	for _, ts := range userReq.timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	userReq.timestamps = validTimestamps

	// Check if limit exceeded
	if len(userReq.timestamps) >= rl.limit {
		return false
	}

	// Add current timestamp
	userReq.timestamps = append(userReq.timestamps, now)
	return true
}

// cleanup removes old entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for identifier, userReq := range rl.requests {
			userReq.mu.Lock()
			// Remove if no recent requests
			if len(userReq.timestamps) == 0 || userReq.timestamps[len(userReq.timestamps)-1].Before(windowStart) {
				delete(rl.requests, identifier)
			}
			userReq.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}
