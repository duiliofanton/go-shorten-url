package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string]*client
	mu       sync.Mutex
	rate     int
	window   time.Duration
	done     chan struct{}
}

type client struct {
	count     int
	expiresAt time.Time
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*client),
		rate:     rate,
		window:   window,
		done:     make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for {
		select {
		case <-rl.done:
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for key, c := range rl.requests {
				if now.After(c.expiresAt) {
					delete(rl.requests, key)
				}
			}
			rl.mu.Unlock()
		}
	}
}

func (rl *RateLimiter) isAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	c, exists := rl.requests[key]
	if !exists || now.After(c.expiresAt) {
		rl.requests[key] = &client{
			count:     1,
			expiresAt: now.Add(rl.window),
		}
		return true
	}
	if c.count >= rl.rate {
		return false
	}
	c.count++
	return true
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.isAllowed(clientIP(r)) {
			writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}
