package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSessionTracking_NoSession(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	e.Use(SessionTracking)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSessionTracking_WithUserID(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Create session with user_id
	c := e.NewContext(req, rec)
	session, _ := GetSession(c)
	testUserID := uuid.New().String()
	session.Values["user_id"] = testUserID
	_ = session.Save(req, rec)

	// Add session cookie to request
	cookies := rec.Result().Cookies()
	req.Header.Set("Cookie", cookies[0].Name+"="+cookies[0].Value)

	// Create new recorder for actual test
	rec2 := httptest.NewRecorder()
	e.Use(SessionTracking)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.ServeHTTP(rec2, req)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestSessionTracking_EmptyUserID(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Create session with empty user_id
	c := e.NewContext(req, rec)
	session, _ := GetSession(c)
	session.Values["user_id"] = ""
	_ = session.Save(req, rec)

	// Add session cookie to request
	cookies := rec.Result().Cookies()
	req.Header.Set("Cookie", cookies[0].Name+"="+cookies[0].Value)

	// Create new recorder for actual test
	rec2 := httptest.NewRecorder()
	e.Use(SessionTracking)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.ServeHTTP(rec2, req)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestSessionTracking_InvalidUserIDType(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Create session with invalid user_id type
	c := e.NewContext(req, rec)
	session, _ := GetSession(c)
	session.Values["user_id"] = 12345 // int instead of string
	_ = session.Save(req, rec)

	// Add session cookie to request
	cookies := rec.Result().Cookies()
	req.Header.Set("Cookie", cookies[0].Name+"="+cookies[0].Value)

	// Create new recorder for actual test
	rec2 := httptest.NewRecorder()
	e.Use(SessionTracking)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.ServeHTTP(rec2, req)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestCleanupInactiveSessions(t *testing.T) {
	// Reset activeSessions
	sessionsMutex.Lock()
	activeSessions = make(map[string]bool)
	activeSessions["session1"] = true
	activeSessions["session2"] = true
	activeSessions["session3"] = true
	sessionsMutex.Unlock()

	// Verify sessions exist
	sessionsMutex.RLock()
	assert.Equal(t, 3, len(activeSessions))
	sessionsMutex.RUnlock()

	// Cleanup
	CleanupInactiveSessions()

	// Verify sessions are cleared
	sessionsMutex.RLock()
	assert.Equal(t, 0, len(activeSessions))
	sessionsMutex.RUnlock()
}

func TestCleanupInactiveSessions_EmptySessions(t *testing.T) {
	// Start with empty sessions
	sessionsMutex.Lock()
	activeSessions = make(map[string]bool)
	sessionsMutex.Unlock()

	// Cleanup should not panic
	assert.NotPanics(t, func() {
		CleanupInactiveSessions()
	})

	sessionsMutex.RLock()
	assert.Equal(t, 0, len(activeSessions))
	sessionsMutex.RUnlock()
}

func TestSessionTracking_ConcurrentAccess(_ *testing.T) {
	setupTestAuth()

	e := echo.New()
	e.Use(SessionTracking)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Make multiple concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic
}

func TestSessionTracking_PassesRequest(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	e.Use(SessionTracking)

	called := false
	e.GET("/test", func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}
