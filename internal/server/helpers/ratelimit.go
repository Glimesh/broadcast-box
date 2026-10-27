package helpers

import (
	"sync"
	"time"
)

// RateLimiter is a simple fixed-window limiter keyed by caller-supplied strings.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateEntry
}

type rateEntry struct {
	count  int
	reset  time.Time
}

// NewRateLimiter creates an empty rate limiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{entries: map[string]rateEntry{}}
}

// DefaultRateLimiter is used by HTTP handlers.
var DefaultRateLimiter = NewRateLimiter()

// Allow reports whether the key may perform another action in the current window.
func (r *RateLimiter) Allow(key string, limit int, window time.Duration) bool {
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.entries[key]
	if !ok || now.After(entry.reset) {
		entry = rateEntry{count: 0, reset: now.Add(window)}
	}

	if entry.count >= limit {
		r.entries[key] = entry
		return false
	}

	entry.count++
	r.entries[key] = entry
	return true
}
