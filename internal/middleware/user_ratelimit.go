// Package middleware provides Echo middleware for authentication, CORS, CSRF, and more.
package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"golang.org/x/time/rate"
)

// userLimiterEntry stores a rate limiter with its last-access timestamp.
type userLimiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// UserRateLimiter manages rate limiters per authenticated user.
type UserRateLimiter struct {
	limiters map[uuid.UUID]*userLimiterEntry
	mu       sync.RWMutex
	r        rate.Limit
	b        int
	ctx      context.Context
	cancel   context.CancelFunc
}

// maxUserLimiters is the maximum number of user entries to prevent unbounded memory growth.
const maxUserLimiters = 10000

// NewUserRateLimiter creates a new per-user rate limiter.
func NewUserRateLimiter(r rate.Limit, b int) *UserRateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	rl := &UserRateLimiter{
		limiters: make(map[uuid.UUID]*userLimiterEntry),
		r:        r,
		b:        b,
		ctx:      ctx,
		cancel:   cancel,
	}

	go rl.cleanup()
	return rl
}

// cleanup periodically removes idle user limiters.
func (u *UserRateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-u.ctx.Done():
			return
		case <-ticker.C:
			u.mu.Lock()
			now := time.Now()
			for id, entry := range u.limiters {
				if now.Sub(entry.lastAccess) > 1*time.Hour {
					delete(u.limiters, id)
				}
			}
			u.mu.Unlock()
		}
	}
}

// GetLimiter returns the rate limiter for the given user ID.
func (u *UserRateLimiter) GetLimiter(userID uuid.UUID) *rate.Limiter {
	u.mu.Lock()
	defer u.mu.Unlock()

	if entry, exists := u.limiters[userID]; exists {
		entry.lastAccess = time.Now()
		return entry.limiter
	}

	// Evict oldest entry if at capacity.
	if len(u.limiters) >= maxUserLimiters {
		var oldestID uuid.UUID
		var oldestTime time.Time
		first := true
		for id, entry := range u.limiters {
			if first || entry.lastAccess.Before(oldestTime) {
				oldestID = id
				oldestTime = entry.lastAccess
				first = false
			}
		}
		delete(u.limiters, oldestID)
	}

	limiter := rate.NewLimiter(u.r, u.b)
	u.limiters[userID] = &userLimiterEntry{
		limiter:    limiter,
		lastAccess: time.Now(),
	}
	return limiter
}

// Shutdown stops the cleanup goroutine.
func (u *UserRateLimiter) Shutdown() {
	u.cancel()
}

// UserRateLimitMiddleware creates a per-user rate limiting middleware.
// Must be placed AFTER authentication middleware (SetCurrentUserWithService + RequireAuth).
func UserRateLimitMiddleware(limiter *UserRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			user, ok := c.Get("current_user").(*models.User)
			if !ok || user == nil {
				// Anonymous request — skip user rate limiting (IP limiter handles this).
				return next(c)
			}

			l := limiter.GetLimiter(user.ID)

			if !l.Allow() {
				c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.b))
				c.Response().Header().Set("X-RateLimit-Remaining", "0")
				retryAfter := max(int(1.0/float64(limiter.r)), 1)
				c.Response().Header().Set("Retry-After", strconv.Itoa(retryAfter))
				return echo.NewHTTPError(http.StatusTooManyRequests, "Too many requests. Please try again later.")
			}

			remaining := max(int(l.Tokens()), 0)
			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.b))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			return next(c)
		}
	}
}
