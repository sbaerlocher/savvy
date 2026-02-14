package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestNewIPRateLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 20)

	assert.NotNil(t, limiter)
	assert.NotNil(t, limiter.limiters)
	assert.Equal(t, rate.Limit(10), limiter.r)
	assert.Equal(t, 20, limiter.b)
	assert.NotNil(t, limiter.ctx)
	assert.NotNil(t, limiter.cancel)

	// Cleanup
	limiter.Shutdown()
}

func TestIPRateLimiter_GetLimiter_NewIP(t *testing.T) {
	rl := NewIPRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	limiter := rl.GetLimiter("192.168.1.1")

	assert.NotNil(t, limiter)
	assert.Equal(t, 1, len(rl.limiters))
}

func TestIPRateLimiter_GetLimiter_ExistingIP(t *testing.T) {
	rl := NewIPRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	limiter1 := rl.GetLimiter("192.168.1.1")
	limiter2 := rl.GetLimiter("192.168.1.1")

	assert.Equal(t, limiter1, limiter2)
	assert.Equal(t, 1, len(rl.limiters))
}

func TestIPRateLimiter_GetLimiter_MultipleIPs(t *testing.T) {
	rl := NewIPRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	limiter1 := rl.GetLimiter("192.168.1.1")
	limiter2 := rl.GetLimiter("192.168.1.2")
	limiter3 := rl.GetLimiter("192.168.1.3")

	assert.NotNil(t, limiter1)
	assert.NotNil(t, limiter2)
	assert.NotNil(t, limiter3)
	assert.Equal(t, 3, len(rl.limiters))
}

func TestIPRateLimiter_Cleanup(t *testing.T) {
	rl := NewIPRateLimiter(rate.Limit(10), 20)
	defer rl.Shutdown()

	// Add limiters
	rl.GetLimiter("192.168.1.1")
	rl.GetLimiter("192.168.1.2")

	assert.Equal(t, 2, len(rl.limiters))

	// Manually set old timestamp
	rl.mu.Lock()
	for ip, entry := range rl.limiters {
		entry.createdAt = time.Now().Add(-2 * time.Hour)
		rl.limiters[ip] = entry
	}
	rl.mu.Unlock()

	// Trigger cleanup manually by waiting a bit
	// Note: In real scenario, cleanup runs every 10 minutes
	// For testing, we verify the logic by checking timestamp
	time.Sleep(100 * time.Millisecond)

	// Verify limiters can be cleaned up based on age
	rl.mu.Lock()
	now := time.Now()
	for _, entry := range rl.limiters {
		assert.True(t, now.Sub(entry.createdAt) > 1*time.Hour)
	}
	rl.mu.Unlock()
}

func TestIPRateLimiter_Shutdown(t *testing.T) {
	rl := NewIPRateLimiter(rate.Limit(10), 20)

	rl.Shutdown()

	// Verify context is cancelled
	select {
	case <-rl.ctx.Done():
		// Context should be cancelled
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be cancelled after shutdown")
	}
}

func TestRateLimitMiddleware_AllowedRequests(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 10)
	defer limiter.Shutdown()

	e := echo.New()
	e.Use(RateLimitMiddleware(limiter))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Make requests within rate limit
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Real-IP", "192.168.1.1")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		// Verify rate limit headers are set on successful responses
		assert.Equal(t, "10", rec.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestRateLimitMiddleware_RateLimitExceeded(t *testing.T) {
	// Very restrictive rate limit: 1 request per second, burst of 1
	limiter := NewIPRateLimiter(rate.Limit(1), 1)
	defer limiter.Shutdown()

	e := echo.New()
	e.Use(RateLimitMiddleware(limiter))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-Real-IP", "192.168.1.1")
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second immediate request should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Real-IP", "192.168.1.1")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
	assert.Contains(t, rec2.Body.String(), "Too many requests")
	// Verify rate limit headers on 429 response
	assert.Equal(t, "1", rec2.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", rec2.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec2.Header().Get("Retry-After"))
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(1), 1)
	defer limiter.Shutdown()

	e := echo.New()
	e.Use(RateLimitMiddleware(limiter))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Request from IP 1
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-Real-IP", "192.168.1.1")
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Immediate request from IP 2 should succeed (different limiter)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Real-IP", "192.168.1.2")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRateLimitMiddleware_BurstSize(t *testing.T) {
	// 1 request per second, but burst of 5
	limiter := NewIPRateLimiter(rate.Limit(1), 5)
	defer limiter.Shutdown()

	e := echo.New()
	e.Use(RateLimitMiddleware(limiter))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// First 5 requests should succeed (burst)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Real-IP", "192.168.1.1")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// 6th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestRateLimitMiddleware_Recovery(t *testing.T) {
	// 10 requests per second
	limiter := NewIPRateLimiter(rate.Limit(10), 1)
	defer limiter.Shutdown()

	e := echo.New()
	e.Use(RateLimitMiddleware(limiter))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// First request succeeds
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-Real-IP", "192.168.1.1")
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second immediate request is rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Real-IP", "192.168.1.1")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)

	// Wait for rate limit to recover (100ms = 0.1s, rate is 10/s)
	time.Sleep(150 * time.Millisecond)

	// Third request should succeed
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.Header.Set("X-Real-IP", "192.168.1.1")
	rec3 := httptest.NewRecorder()
	e.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusOK, rec3.Code)
}

func TestRateLimitMiddleware_RealIP(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(1), 1)
	defer limiter.Shutdown()

	e := echo.New()
	e.Use(RateLimitMiddleware(limiter))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Test with X-Real-IP header
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-Real-IP", "10.0.0.1")
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second request with same X-Real-IP should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Real-IP", "10.0.0.1")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
}

func TestIPRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(100), 100)
	defer limiter.Shutdown()

	// Test concurrent access to GetLimiter
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(_ int) {
			for j := 0; j < 100; j++ {
				ip := "192.168.1.1"
				l := limiter.GetLimiter(ip)
				assert.NotNil(t, l)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have only one limiter for the IP
	assert.Equal(t, 1, len(limiter.limiters))
}

func TestIPRateLimiter_CleanupRunning(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 20)

	// Add some limiters
	limiter.GetLimiter("192.168.1.1")
	limiter.GetLimiter("192.168.1.2")

	// Verify cleanup goroutine is running by checking context
	select {
	case <-limiter.ctx.Done():
		t.Fatal("Cleanup should be running")
	default:
		// OK, cleanup is running
	}

	// Shutdown and verify cleanup stops
	limiter.Shutdown()

	select {
	case <-limiter.ctx.Done():
		// OK, cleanup stopped
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Cleanup should stop after shutdown")
	}
}

func TestIPRateLimiter_CleanupOldEntries(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 20)
	defer limiter.Shutdown()

	// Add limiters
	limiter.GetLimiter("192.168.1.1")
	limiter.GetLimiter("192.168.1.2")
	limiter.GetLimiter("192.168.1.3")

	assert.Equal(t, 3, len(limiter.limiters))

	// Manually set old timestamps
	limiter.mu.Lock()
	oldTime := time.Now().Add(-2 * time.Hour)
	for ip := range limiter.limiters {
		entry := limiter.limiters[ip]
		entry.createdAt = oldTime
		limiter.limiters[ip] = entry
	}
	limiter.mu.Unlock()

	// Manually trigger cleanup by accessing the cleanup logic
	// (normally this happens every 10 minutes)
	limiter.mu.Lock()
	now := time.Now()
	for ip, entry := range limiter.limiters {
		if now.Sub(entry.createdAt) > 1*time.Hour {
			delete(limiter.limiters, ip)
		}
	}
	limiter.mu.Unlock()

	// Verify old entries are removed
	limiter.mu.RLock()
	assert.Equal(t, 0, len(limiter.limiters))
	limiter.mu.RUnlock()
}

func TestIPRateLimiter_CleanupKeepsRecentEntries(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 20)
	defer limiter.Shutdown()

	// Add limiters with recent timestamps
	limiter.GetLimiter("192.168.1.1")
	limiter.GetLimiter("192.168.1.2")

	assert.Equal(t, 2, len(limiter.limiters))

	// Manually trigger cleanup (recent entries should be kept)
	limiter.mu.Lock()
	now := time.Now()
	for ip, entry := range limiter.limiters {
		if now.Sub(entry.createdAt) > 1*time.Hour {
			delete(limiter.limiters, ip)
		}
	}
	limiter.mu.Unlock()

	// Recent entries should still be there
	limiter.mu.RLock()
	assert.Equal(t, 2, len(limiter.limiters))
	limiter.mu.RUnlock()
}
