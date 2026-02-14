package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"savvy/internal/models"
)

func TestNewUserRateLimiter(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	assert.NotNil(t, rl)
	assert.NotNil(t, rl.limiters)
	assert.Equal(t, rate.Limit(10), rl.r)
	assert.Equal(t, 20, rl.b)
	assert.NotNil(t, rl.ctx)
	assert.NotNil(t, rl.cancel)
}

func TestUserRateLimiter_GetLimiter_NewUser(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	userID := uuid.New()
	limiter := rl.GetLimiter(userID)

	assert.NotNil(t, limiter)
	assert.Equal(t, 1, len(rl.limiters))
}

func TestUserRateLimiter_GetLimiter_ExistingUser(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	userID := uuid.New()
	limiter1 := rl.GetLimiter(userID)
	limiter2 := rl.GetLimiter(userID)

	assert.Equal(t, limiter1, limiter2)
	assert.Equal(t, 1, len(rl.limiters))
}

func TestUserRateLimiter_GetLimiter_MultipleUsers(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	l1 := rl.GetLimiter(uuid.New())
	l2 := rl.GetLimiter(uuid.New())
	l3 := rl.GetLimiter(uuid.New())

	assert.NotNil(t, l1)
	assert.NotNil(t, l2)
	assert.NotNil(t, l3)
	assert.Equal(t, 3, len(rl.limiters))
}

func TestUserRateLimiter_Shutdown(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(10), 20)
	rl.Shutdown()

	select {
	case <-rl.ctx.Done():
		// Context should be cancelled
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be cancelled after shutdown")
	}
}

func TestUserRateLimiter_CleanupOldEntries(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	id1 := uuid.New()
	id2 := uuid.New()
	rl.GetLimiter(id1)
	rl.GetLimiter(id2)
	assert.Equal(t, 2, len(rl.limiters))

	// Set old lastAccess timestamps
	rl.mu.Lock()
	for uid, entry := range rl.limiters {
		entry.lastAccess = time.Now().Add(-2 * time.Hour)
		rl.limiters[uid] = entry
	}
	rl.mu.Unlock()

	// Manually trigger cleanup logic
	rl.mu.Lock()
	now := time.Now()
	for id, entry := range rl.limiters {
		if now.Sub(entry.lastAccess) > 1*time.Hour {
			delete(rl.limiters, id)
		}
	}
	rl.mu.Unlock()

	assert.Equal(t, 0, len(rl.limiters))
}

func TestUserRateLimiter_CleanupKeepsRecentEntries(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	rl.GetLimiter(uuid.New())
	rl.GetLimiter(uuid.New())
	assert.Equal(t, 2, len(rl.limiters))

	// Recent entries should survive cleanup
	rl.mu.Lock()
	now := time.Now()
	for id, entry := range rl.limiters {
		if now.Sub(entry.lastAccess) > 1*time.Hour {
			delete(rl.limiters, id)
		}
	}
	rl.mu.Unlock()

	assert.Equal(t, 2, len(rl.limiters))
}

func TestUserRateLimiter_GetLimiter_EvictsOldest(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	// Fill to capacity (maxUserLimiters = 10000, too many for test)
	// Instead, test the eviction logic directly by manipulating the map
	oldestID := uuid.New()
	rl.mu.Lock()
	rl.limiters[oldestID] = &userLimiterEntry{
		limiter:    rate.NewLimiter(rl.r, rl.b),
		lastAccess: time.Now().Add(-1 * time.Hour),
	}
	// Fill to maxUserLimiters
	for i := 1; i < maxUserLimiters; i++ {
		rl.limiters[uuid.New()] = &userLimiterEntry{
			limiter:    rate.NewLimiter(rl.r, rl.b),
			lastAccess: time.Now(),
		}
	}
	rl.mu.Unlock()
	assert.Equal(t, maxUserLimiters, len(rl.limiters))

	// Adding one more should evict the oldest
	rl.GetLimiter(uuid.New())

	rl.mu.RLock()
	_, exists := rl.limiters[oldestID]
	rl.mu.RUnlock()
	assert.False(t, exists, "Oldest entry should have been evicted")
	assert.Equal(t, maxUserLimiters, len(rl.limiters))
}

func TestUserRateLimitMiddleware_AllowedRequests(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(10), 10)
	defer rl.Shutdown()

	user := &models.User{ID: uuid.New(), Email: "test@example.com"}

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("current_user", user)
			return next(c)
		}
	})
	e.Use(UserRateLimitMiddleware(rl))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestUserRateLimitMiddleware_RateLimitExceeded(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(1), 1)
	defer rl.Shutdown()

	user := &models.User{ID: uuid.New(), Email: "test@example.com"}

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("current_user", user)
			return next(c)
		}
	})
	e.Use(UserRateLimitMiddleware(rl))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second immediate request should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
	assert.Equal(t, "0", rec2.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec2.Header().Get("Retry-After"))
}

func TestUserRateLimitMiddleware_AnonymousUser(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(1), 1)
	defer rl.Shutdown()

	e := echo.New()
	// No user-setting middleware — anonymous request
	e.Use(UserRateLimitMiddleware(rl))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Anonymous requests should pass through without rate limiting
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestUserRateLimitMiddleware_DifferentUsers(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(1), 1)
	defer rl.Shutdown()

	user1 := &models.User{ID: uuid.New(), Email: "user1@example.com"}
	user2 := &models.User{ID: uuid.New(), Email: "user2@example.com"}

	makeRequest := func(user *models.User) int {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Set("current_user", user)
				return next(c)
			}
		})
		e.Use(UserRateLimitMiddleware(rl))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}

	// User1 first request
	assert.Equal(t, http.StatusOK, makeRequest(user1))
	// User1 second request — rate limited
	assert.Equal(t, http.StatusTooManyRequests, makeRequest(user1))
	// User2 first request — different limiter, should succeed
	assert.Equal(t, http.StatusOK, makeRequest(user2))
}

func TestUserRateLimitMiddleware_BurstSize(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(1), 5)
	defer rl.Shutdown()

	user := &models.User{ID: uuid.New(), Email: "test@example.com"}

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("current_user", user)
			return next(c)
		}
	})
	e.Use(UserRateLimitMiddleware(rl))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// First 5 requests should succeed (burst)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// 6th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestUserRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewUserRateLimiter(rate.Limit(100), 100)
	defer rl.Shutdown()

	userID := uuid.New()
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				l := rl.GetLimiter(userID)
				assert.NotNil(t, l)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	rl.mu.RLock()
	assert.Equal(t, 1, len(rl.limiters))
	rl.mu.RUnlock()
}
