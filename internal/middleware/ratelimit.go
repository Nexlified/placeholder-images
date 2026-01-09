package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// limiterEntry stores a rate limiter and its last access time
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// RateLimiter manages per-IP rate limiters
type RateLimiter struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
	rpm      int           // Requests per minute
	burst    int           // Burst size
	cleanup  time.Duration // Cleanup interval for stale entries
}

// NewRateLimiter creates a new rate limiter with the given requests per minute and burst size
func NewRateLimiter(rpm, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*limiterEntry),
		rpm:      rpm,
		burst:    burst,
		cleanup:  time.Minute * 10, // Clean up stale entries every 10 minutes
	}

	// Start cleanup goroutine
	go rl.cleanupStaleEntries()

	return rl
}

// getLimiter returns the rate limiter for the given IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[ip]
	if !exists {
		// Convert RPM to requests per second for rate.Limit
		rps := rate.Limit(float64(rl.rpm) / 60.0)
		entry = &limiterEntry{
			limiter:    rate.NewLimiter(rps, rl.burst),
			lastAccess: time.Now(),
		}
		rl.limiters[ip] = entry
	} else {
		// Update last access time
		entry.lastAccess = time.Now()
	}

	return entry.limiter
}

// cleanupStaleEntries periodically removes rate limiters that haven't been used recently
func (rl *RateLimiter) cleanupStaleEntries() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		// Remove entries that haven't been accessed in the last 10 minutes
		for ip, entry := range rl.limiters {
			if now.Sub(entry.lastAccess) > time.Minute*10 {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getIP extracts the client IP from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	// X-Forwarded-For format: client, proxy1, proxy2, ...
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Split by comma and take the first IP (original client)
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Check X-Real-IP header
	realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// Middleware creates an HTTP middleware that applies rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
