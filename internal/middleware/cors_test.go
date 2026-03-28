package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig(nil)

	assert.Equal(t, []string{"http://localhost:5173", "http://localhost:3000"}, config.AllowOrigins)
	assert.Contains(t, config.AllowMethods, http.MethodGet)
	assert.Contains(t, config.AllowMethods, http.MethodPost)
	assert.Contains(t, config.AllowMethods, http.MethodPut)
	assert.Contains(t, config.AllowMethods, http.MethodPatch)
	assert.Contains(t, config.AllowMethods, http.MethodDelete)
	assert.Contains(t, config.AllowMethods, http.MethodOptions)
	assert.Contains(t, config.AllowHeaders, "Content-Type")
	assert.Contains(t, config.AllowHeaders, "Authorization")
	assert.Contains(t, config.AllowHeaders, CSRFHeaderName)
	assert.True(t, config.AllowCredentials)
}

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}

	e := echo.New()
	e.Use(CORSMiddleware(config))
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://localhost:5173", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{http.MethodGet},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}

	e := echo.New()
	e.Use(CORSMiddleware(config))
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Request should succeed but CORS headers should not be set
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_WildcardOrigin(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: false,
	}

	e := echo.New()
	e.Use(CORSMiddleware(config))
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://any-domain.com")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://any-domain.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowHeaders:     []string{"Content-Type", "Authorization", CSRFHeaderName},
		AllowCredentials: true,
	}

	e := echo.New()
	e.Use(CORSMiddleware(config))
	e.OPTIONS("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "http://localhost:5173", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), http.MethodGet)
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), http.MethodPost)
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), http.MethodDelete)
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), CSRFHeaderName)
	assert.Equal(t, "86400", rec.Header().Get("Access-Control-Max-Age"))
}

func TestCORSMiddleware_PreflightDisallowedOrigin(t *testing.T) {
	config := CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{http.MethodGet},
		AllowHeaders: []string{"Content-Type"},
	}

	e := echo.New()
	e.Use(CORSMiddleware(config))
	e.OPTIONS("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Preflight should still succeed but without CORS headers
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_NoOriginHeader(t *testing.T) {
	config := DefaultCORSConfig(nil)

	e := echo.New()
	e.Use(CORSMiddleware(config))
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Origin header
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_CredentialsDisabled(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{http.MethodGet},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: false,
	}

	e := echo.New()
	e.Use(CORSMiddleware(config))
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://localhost:5173", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		sep      string
		expected string
	}{
		{
			name:     "Empty slice",
			input:    []string{},
			sep:      ", ",
			expected: "",
		},
		{
			name:     "Single element",
			input:    []string{"one"},
			sep:      ", ",
			expected: "one",
		},
		{
			name:     "Multiple elements",
			input:    []string{"one", "two", "three"},
			sep:      ", ",
			expected: "one, two, three",
		},
		{
			name:     "Different separator",
			input:    []string{"a", "b", "c"},
			sep:      "|",
			expected: "a|b|c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.input, tt.sep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCORSMiddleware_WildcardWithCredentials(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}

	e := echo.New()
	e.Use(CORSMiddleware(config))
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://any-domain.com")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Wildcard should be ignored when credentials are enabled (CORS spec violation)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_MultipleOrigins(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "https://example.com"},
		AllowMethods:     []string{http.MethodGet},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}

	e := echo.New()
	e.Use(CORSMiddleware(config))
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Test each allowed origin
	allowedOrigins := []string{"http://localhost:5173", "http://localhost:3000", "https://example.com"}
	for _, origin := range allowedOrigins {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, origin, rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
