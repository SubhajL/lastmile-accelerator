package security

import (
	"sync"
	"time"
)

// RateLimiter is a simple token-bucket limiter keyed by string.
// It is safe for concurrent use.
// rate: tokens added per second; burst: maximum tokens per bucket.

type RateLimiter struct {
	mu    sync.Mutex
	bkt   map[string]*bucket
	rate  float64
	burst float64
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewRateLimiter(rate float64, burst int) *RateLimiter {
	if rate <= 0 {
		rate = 10
	}
	if burst <= 0 {
		burst = 20
	}
	return &RateLimiter{bkt: make(map[string]*bucket), rate: rate, burst: float64(burst)}
}

// Allow returns true if a token is available for the given key.
func (r *RateLimiter) Allow(key string) bool {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bkt[key]
	if !ok {
		b = &bucket{tokens: r.burst - 1, last: now}
		r.bkt[key] = b
		return true
	}
	// refill
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * r.rate
	if b.tokens > r.burst {
		b.tokens = r.burst
	}
	b.last = now
	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}
