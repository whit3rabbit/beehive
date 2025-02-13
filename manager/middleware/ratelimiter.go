package middleware

import (
	"sync"
	"time"
)

type rateLimiter struct {
	maxAttempts     int
	window          time.Duration
	blockoutPeriod  time.Duration
	attempts        map[string][]time.Time
	blockedUntil    map[string]time.Time
	mu             sync.RWMutex
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(maxAttempts int, window, blockoutPeriod time.Duration) *rateLimiter {
	return &rateLimiter{
		maxAttempts:    maxAttempts,
		window:         window,
		blockoutPeriod: blockoutPeriod,
		attempts:       make(map[string][]time.Time),
		blockedUntil:   make(map[string]time.Time),
	}
}

// CheckLimit checks if the key has exceeded its rate limit
func (rl *rateLimiter) CheckLimit(key string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Check if key is blocked
	if blockedUntil, exists := rl.blockedUntil[key]; exists {
		if now.Before(blockedUntil) {
			return false, time.Until(blockedUntil)
		}
		delete(rl.blockedUntil, key)
	}

	// Clean up old attempts
	windowStart := now.Add(-rl.window)
	attempts := rl.attempts[key]
	validAttempts := make([]time.Time, 0)

	for _, attempt := range attempts {
		if attempt.After(windowStart) {
			validAttempts = append(validAttempts, attempt)
		}
	}

	// Update attempts
	rl.attempts[key] = append(validAttempts, now)

	// Check if limit exceeded
	if len(rl.attempts[key]) > rl.maxAttempts {
		rl.blockedUntil[key] = now.Add(rl.blockoutPeriod)
		return false, rl.blockoutPeriod
	}

	return true, 0
}