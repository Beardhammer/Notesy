package main

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	tokens int
	last   time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity int
	refill   time.Duration
}

func NewRateLimiter(capacity int, refill time.Duration) *RateLimiter {
	return &RateLimiter{buckets: map[string]*bucket{}, capacity: capacity, refill: refill}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{tokens: rl.capacity, last: now}
		rl.buckets[key] = b
	}
	elapsed := now.Sub(b.last)
	add := int(elapsed / rl.refill)
	if add > 0 {
		if b.tokens+add > rl.capacity {
			b.tokens = rl.capacity
		} else {
			b.tokens += add
		}
		b.last = now
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if !rl.Allow(ip) {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
