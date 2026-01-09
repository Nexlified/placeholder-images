package middleware

import (
"net"
"net/http"
"sync"
"time"

"golang.org/x/time/rate"
)

// RateLimiter manages per-IP rate limiters
type RateLimiter struct {
limiters map[string]*rate.Limiter
mu       sync.RWMutex
rpm      int           // Requests per minute
burst    int           // Burst size
cleanup  time.Duration // Cleanup interval for stale entries
}

// NewRateLimiter creates a new rate limiter with the given requests per minute and burst size
func NewRateLimiter(rpm, burst int) *RateLimiter {
rl := &RateLimiter{
limiters: make(map[string]*rate.Limiter),
rpm:      rpm,
burst:    burst,
cleanup:  time.Minute * 5, // Clean up stale entries every 5 minutes
}

// Start cleanup goroutine
go rl.cleanupStaleEntries()

return rl
}

// getLimiter returns the rate limiter for the given IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
rl.mu.Lock()
defer rl.mu.Unlock()

limiter, exists := rl.limiters[ip]
if !exists {
// Convert RPM to requests per second for rate.Limit
rps := rate.Limit(float64(rl.rpm) / 60.0)
limiter = rate.NewLimiter(rps, rl.burst)
rl.limiters[ip] = limiter
}

return limiter
}

// cleanupStaleEntries periodically removes rate limiters that haven't been used
func (rl *RateLimiter) cleanupStaleEntries() {
ticker := time.NewTicker(rl.cleanup)
defer ticker.Stop()

for range ticker.C {
rl.mu.Lock()
// Simple cleanup: remove all entries periodically
// More sophisticated approach would track last access time
rl.limiters = make(map[string]*rate.Limiter)
rl.mu.Unlock()
}
}

// getIP extracts the client IP from the request
func getIP(r *http.Request) string {
// Check X-Forwarded-For header first (for proxies)
forwarded := r.Header.Get("X-Forwarded-For")
if forwarded != "" {
// X-Forwarded-For can contain multiple IPs, take the first one
if ip, _, err := net.SplitHostPort(forwarded); err == nil {
return ip
}
return forwarded
}

// Check X-Real-IP header
realIP := r.Header.Get("X-Real-IP")
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
