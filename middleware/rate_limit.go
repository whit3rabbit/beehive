package middleware

import (
	"sync"
	"time"
)

// RateLimiter struct to hold rate limiting parameters and attempt tracking
type RateLimiter struct {
	attempts      map[string]*RateLimit
	maxAttempts   int
	windowSecs    int
	blockoutMins  int
	mu            sync.RWMutex
}

// RateLimit struct to track individual client attempts
type RateLimit struct {
	count       int
	lastAttempt time.Time
	blockUntil  time.Time
}

// NewRateLimiter creates a new RateLimiter with the given parameters
func NewRateLimiter(maxAttempts int, windowSecs int, blockoutMins int) *RateLimiter {
	return &RateLimiter{
		attempts:      make(map[string]*RateLimit),
		maxAttempts:   maxAttempts,
		windowSecs:    windowSecs,
		blockoutMins:  blockoutMins,
	}
}

// CheckLimit checks if the given key is allowed to proceed, and returns the wait duration if not
func (rl *RateLimiter) CheckLimit(key string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit, exists := rl.attempts[key]

	if !exists {
		rl.attempts[key] = &RateLimit{count: 1, lastAttempt: now}
		return true, 0
	}

	// Check if in blockout period
	if !limit.blockUntil.IsZero() && now.Before(limit.blockUntil) {
		return false, time.Until(limit.blockUntil)
	}

	// Reset if window expired
	if now.Sub(limit.lastAttempt) > time.Duration(rl.windowSecs)*time.Second {
		limit.count = 1
		limit.lastAttempt = now
		limit.blockUntil = time.Time{}
		return true, 0
	}

	// Increment counter
	limit.count++
	limit.lastAttempt = now

	// Check if should block
	if limit.count > rl.maxAttempts {
		blockoutDuration := time.Duration(rl.blockoutMins) * time.Minute
		limit.blockUntil = now.Add(blockoutDuration)
		return false, blockoutDuration
	}

	return true, 0
}
