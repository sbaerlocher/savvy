// Package middleware contains Echo middleware for authentication, sessions, and observability.
package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// limiterEntry stores a rate limiter with its creation timestamp
type limiterEntry struct {
	limiter   *rate.Limiter
	createdAt time.Time
}

// IPRateLimiter manages rate limiters per IP address
type IPRateLimiter struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
	r        rate.Limit // requests per second
	b        int        // burst size
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewIPRateLimiter creates a new rate limiter with the specified rate and burst
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	rl := &IPRateLimiter{
		limiters: make(map[string]*limiterEntry),
		r:        r,
		b:        b,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start periodic cleanup goroutine
	go rl.cleanup()
	return rl
}

// cleanup periodically removes old limiters to prevent memory leaks
func (i *IPRateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-i.ctx.Done():
			return // Cleanup on shutdown
		case <-ticker.C:
			i.mu.Lock()
			now := time.Now()
			for ip, entry := range i.limiters {
				if now.Sub(entry.createdAt) > 1*time.Hour {
					delete(i.limiters, ip)
				}
			}
			i.mu.Unlock()
		}
	}
}

// maxLimiters is the maximum number of IP entries to prevent unbounded memory growth.
const maxLimiters = 10000

// GetLimiter returns the rate limiter for the given IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	if entry, exists := i.limiters[ip]; exists {
		return entry.limiter
	}

	// Evict oldest entries if at capacity
	if len(i.limiters) >= maxLimiters {
		var oldestIP string
		var oldestTime time.Time
		for ip, entry := range i.limiters {
			if oldestIP == "" || entry.createdAt.Before(oldestTime) {
				oldestIP = ip
				oldestTime = entry.createdAt
			}
		}
		delete(i.limiters, oldestIP)
	}

	limiter := rate.NewLimiter(i.r, i.b)
	i.limiters[ip] = &limiterEntry{
		limiter:   limiter,
		createdAt: time.Now(),
	}
	return limiter
}

// Shutdown stops the cleanup goroutine
func (i *IPRateLimiter) Shutdown() {
	i.cancel()
}

// RateLimitMiddleware creates a rate limiting middleware
// r is the rate (requests per second), b is the burst size
func RateLimitMiddleware(limiter *IPRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			l := limiter.GetLimiter(ip)

			if !l.Allow() {
				c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.b))
				c.Response().Header().Set("X-RateLimit-Remaining", "0")
				retryAfter := int(1.0 / float64(limiter.r))
				if retryAfter < 1 {
					retryAfter = 1
				}
				c.Response().Header().Set("Retry-After", strconv.Itoa(retryAfter))
				return echo.NewHTTPError(http.StatusTooManyRequests, "Too many requests. Please try again later.")
			}

			// Set rate limit info headers on successful requests
			remaining := int(l.Tokens())
			if remaining < 0 {
				remaining = 0
			}
			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.b))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			return next(c)
		}
	}
}
