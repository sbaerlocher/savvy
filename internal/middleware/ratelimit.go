// Package middleware contains Echo middleware for authentication, sessions, and observability.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"golang.org/x/time/rate"
)

// IPRateLimiter manages rate limiters per IP address
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	r        rate.Limit // requests per second
	b        int        // burst size
}

// NewIPRateLimiter creates a new rate limiter with the specified rate and burst
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

// GetLimiter returns the rate limiter for the given IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.limiters[ip] = limiter

		// Clean up old limiters after 1 hour
		go func() {
			time.Sleep(1 * time.Hour)
			i.mu.Lock()
			delete(i.limiters, ip)
			i.mu.Unlock()
		}()
	}

	return limiter
}

// RateLimitMiddleware creates a rate limiting middleware
// r is the rate (requests per second), b is the burst size
func RateLimitMiddleware(limiter *IPRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			l := limiter.GetLimiter(ip)

			if !l.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "Too many requests. Please try again later.")
			}

			return next(c)
		}
	}
}
