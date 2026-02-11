package auth

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple per-client token-bucket rate limiter.
// It is safe for concurrent use and self-cleans stale entries on each Allow call.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*bucket
	rate     int           // tokens added per interval
	interval time.Duration // refill interval
	burst    int           // max burst size
}

type bucket struct {
	tokens   int
	lastSeen time.Time
}

// NewRateLimiter creates a RateLimiter.
//
//	rate     – requests allowed per interval per client
//	interval – the refill window (e.g. time.Minute for rate/min)
//	burst    – maximum accumulated requests above rate (0 = same as rate)
func NewRateLimiter(rate int, interval time.Duration, burst int) *RateLimiter {
	if burst <= 0 {
		burst = rate
	}
	return &RateLimiter{
		clients:  make(map[string]*bucket),
		rate:     rate,
		interval: interval,
		burst:    burst,
	}
}

// Allow reports whether the given client key (IP, token subject, etc.) may
// proceed. Returns the number of remaining tokens in this window.
func (rl *RateLimiter) Allow(key string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.clients[key]
	if !ok {
		b = &bucket{tokens: rl.burst, lastSeen: now}
		rl.clients[key] = b
	}

	// Refill tokens proportional to elapsed time
	elapsed := now.Sub(b.lastSeen)
	refill := int(elapsed / rl.interval) * rl.rate
	if refill > 0 {
		b.tokens += refill
		if b.tokens > rl.burst {
			b.tokens = rl.burst
		}
		b.lastSeen = now
	}

	rl.evict(now)

	if b.tokens <= 0 {
		return false, 0
	}
	b.tokens--
	return true, b.tokens
}

// evict removes entries that haven't been seen for > 10 intervals.
// Must be called with rl.mu held.
func (rl *RateLimiter) evict(now time.Time) {
	cutoff := now.Add(-10 * rl.interval)
	for k, b := range rl.clients {
		if b.lastSeen.Before(cutoff) {
			delete(rl.clients, k)
		}
	}
}

// Middleware returns an HTTP middleware that enforces the rate limit.
// The client key is derived from the X-Forwarded-For or RemoteAddr header.
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := clientKey(r)
			ok, remaining := rl.Allow(key)
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			if !ok {
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientKey(r *http.Request) string {
	// Honour proxy headers first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first address (closest client)
		if idx := len(xff); idx > 0 {
			for i, c := range xff {
				if c == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}
	// Fall back to RemoteAddr (strip port)
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}
