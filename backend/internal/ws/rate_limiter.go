package ws

import (
	"sync"
	"time"
)

type fixedWindowRateLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	maxHits int
	buckets map[string]rateBucket
}

type rateBucket struct {
	windowStart time.Time
	hits        int
}

func newFixedWindowRateLimiter(window time.Duration, maxHits int) *fixedWindowRateLimiter {
	if window <= 0 {
		window = 10 * time.Second
	}
	if maxHits <= 0 {
		maxHits = 20
	}

	return &fixedWindowRateLimiter{
		window:  window,
		maxHits: maxHits,
		buckets: make(map[string]rateBucket),
	}
}

func (r *fixedWindowRateLimiter) Allow(key string, now time.Time) bool {
	if r == nil || key == "" {
		return true
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	b := r.buckets[key]
	if b.windowStart.IsZero() || now.Sub(b.windowStart) >= r.window {
		r.buckets[key] = rateBucket{
			windowStart: now,
			hits:        1,
		}
		return true
	}

	if b.hits >= r.maxHits {
		return false
	}

	b.hits++
	r.buckets[key] = b
	return true
}
