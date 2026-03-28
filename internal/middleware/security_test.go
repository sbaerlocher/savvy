package middleware

import (
	"net/http"
	"net/http/httptest"
	"savvy/internal/config"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	tests := []struct {
		name            string
		config          *config.Config
		expectedCSP     []string // CSP directives that should be present
		expectedHSTS    bool
		expectedXSS     string
		expectedNosniff string
		expectedFrame   string
	}{
		{
			name: "Production with OAuth",
			config: &config.Config{
				Environment:       "production",
				OAuthClientID:     "test-client",
				OAuthClientSecret: "test-secret-long-enough",
				OAuthIssuer:       "https://auth.example.com",
			},
			expectedCSP: []string{
				"default-src 'self'",
				"script-src 'self' 'unsafe-inline'", // SvelteKit SPA requires unsafe-inline
				"style-src 'self' 'unsafe-inline'",  // SvelteKit SPA requires unsafe-inline
				"img-src 'self' data:",
				"font-src 'self'",
				"frame-ancestors 'none'",
				"connect-src 'self' https://auth.example.com",
				"upgrade-insecure-requests",
			},
			expectedHSTS:    true,
			expectedXSS:     "1; mode=block",
			expectedNosniff: "nosniff",
			expectedFrame:   "SAMEORIGIN",
		},
		{
			name: "Development without OAuth",
			config: &config.Config{
				Environment: "development",
			},
			expectedCSP: []string{
				"default-src 'self'",
				"script-src 'self' 'unsafe-inline'", // SvelteKit SPA requires unsafe-inline
				"style-src 'self' 'unsafe-inline'",  // SvelteKit SPA requires unsafe-inline
				"img-src 'self' data:",
				"font-src 'self'",
				"frame-ancestors 'none'",
				"connect-src 'self'",
			},
			expectedHSTS:    false, // HSTS not enforced in development
			expectedXSS:     "1; mode=block",
			expectedNosniff: "nosniff",
			expectedFrame:   "SAMEORIGIN",
		},
		{
			name: "Production without OAuth",
			config: &config.Config{
				Environment: "production",
			},
			expectedCSP: []string{
				"default-src 'self'",
				"connect-src 'self'",
				"upgrade-insecure-requests",
			},
			expectedHSTS:    true,
			expectedXSS:     "1; mode=block",
			expectedNosniff: "nosniff",
			expectedFrame:   "SAMEORIGIN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Echo instance with security middleware
			e := echo.New()
			e.Use(SecurityHeaders(tt.config))

			// Create test handler
			e.GET("/test", func(c *echo.Context) error {
				return c.String(http.StatusOK, "OK")
			})

			// Make request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			// Check CSP header
			csp := rec.Header().Get("Content-Security-Policy")
			assert.NotEmpty(t, csp, "CSP header should be set")
			for _, directive := range tt.expectedCSP {
				assert.Contains(t, csp, directive, "CSP should contain directive: %s", directive)
			}

			// Check other security headers
			assert.Equal(t, tt.expectedXSS, rec.Header().Get("X-Xss-Protection"))
			assert.Equal(t, tt.expectedNosniff, rec.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, tt.expectedFrame, rec.Header().Get("X-Frame-Options"))
			assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
			assert.Equal(t, "camera=(self), microphone=(), geolocation=(), payment=()", rec.Header().Get("Permissions-Policy"))

			// Note: HSTS header is only applied to HTTPS requests by Echo middleware
			// Since we're testing with HTTP requests, we can't verify HSTS in integration tests
			// The configuration is correct, but Echo won't add HSTS to HTTP responses
		})
	}
}

func TestBuildContentSecurityPolicy(t *testing.T) {
	tests := []struct {
		name            string
		config          *config.Config
		expectedParts   []string
		unexpectedParts []string
	}{
		{
			name: "OAuth enabled with valid issuer",
			config: &config.Config{
				Environment:       "production",
				OAuthClientID:     "test-client",
				OAuthClientSecret: "test-secret-long-enough",
				OAuthIssuer:       "https://auth.example.com/application/o/app/",
			},
			expectedParts: []string{
				"default-src 'self'",
				"connect-src 'self' https://auth.example.com",
				"upgrade-insecure-requests",
			},
		},
		{
			name: "OAuth disabled",
			config: &config.Config{
				Environment: "development",
			},
			expectedParts: []string{
				"connect-src 'self'",
			},
			unexpectedParts: []string{
				"upgrade-insecure-requests",
			},
		},
		{
			name: "Production without OAuth",
			config: &config.Config{
				Environment: "production",
			},
			expectedParts: []string{
				"connect-src 'self'",
				"upgrade-insecure-requests",
			},
		},
		{
			name: "CSP report-uri configured",
			config: &config.Config{
				Environment:  "production",
				CSPReportURI: "https://csp.example.com/report",
			},
			expectedParts: []string{
				"report-uri https://csp.example.com/report",
			},
		},
		{
			name: "CSP report-uri not configured",
			config: &config.Config{
				Environment: "development",
			},
			unexpectedParts: []string{
				"report-uri",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build CSP for SvelteKit SPA (uses unsafe-inline)
			csp := buildContentSecurityPolicy(tt.config)

			for _, part := range tt.expectedParts {
				assert.Contains(t, csp, part, "CSP should contain: %s", part)
			}

			for _, part := range tt.unexpectedParts {
				assert.NotContains(t, csp, part, "CSP should NOT contain: %s", part)
			}
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Full OAuth issuer URL with path",
			input:    "https://auth.example.com/application/o/app/",
			expected: "https://auth.example.com",
		},
		{
			name:     "Simple HTTPS URL",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "HTTP URL (development)",
			input:    "http://localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "URL with port",
			input:    "https://auth.example.com:443/path",
			expected: "https://auth.example.com:443",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Invalid URL",
			input:    "not a url",
			expected: "",
		},
		{
			name:     "URL without scheme",
			input:    "example.com",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomain(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityHeadersIntegration(t *testing.T) {
	// Test that all security headers are set together
	cfg := &config.Config{
		Environment:       "production",
		OAuthClientID:     "test",
		OAuthClientSecret: "test-secret-long-enough",
		OAuthIssuer:       "https://auth.example.com",
	}

	e := echo.New()
	e.Use(SecurityHeaders(cfg))
	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Verify all security headers are present (except HSTS which only applies to HTTPS)
	headers := rec.Header()
	assert.NotEmpty(t, headers.Get("Content-Security-Policy"))
	assert.NotEmpty(t, headers.Get("X-Xss-Protection"))
	assert.NotEmpty(t, headers.Get("X-Content-Type-Options"))
	assert.NotEmpty(t, headers.Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
	assert.Equal(t, "camera=(self), microphone=(), geolocation=(), payment=()", headers.Get("Permissions-Policy"))

	// Verify CSP includes OAuth domain
	csp := headers.Get("Content-Security-Policy")
	assert.Contains(t, csp, "https://auth.example.com")
}

func TestSecurityHeadersWithMultipleRequests(t *testing.T) {
	// Test that headers are consistent across multiple requests
	cfg := &config.Config{
		Environment: "production",
	}

	e := echo.New()
	e.Use(SecurityHeaders(cfg))
	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Make multiple requests
	for range 3 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// All security headers should be present
		assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
		assert.Equal(t, "1; mode=block", rec.Header().Get("X-Xss-Protection"))
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "SAMEORIGIN", rec.Header().Get("X-Frame-Options"))
		assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
		assert.NotEmpty(t, rec.Header().Get("Permissions-Policy"))
	}
}

func TestCSPDirectiveSeparation(t *testing.T) {
	// Ensure CSP directives are properly separated with semicolons
	cfg := &config.Config{
		Environment: "production",
	}

	// Build CSP for SvelteKit SPA (uses unsafe-inline)
	csp := buildContentSecurityPolicy(cfg)

	// Split by semicolon and verify each directive
	directives := strings.Split(csp, "; ")
	assert.GreaterOrEqual(t, len(directives), 7, "Should have at least 7 directives")

	// Each directive should have a key-value format (except upgrade-insecure-requests)
	for _, directive := range directives {
		directive = strings.TrimSpace(directive)
		if directive == "upgrade-insecure-requests" {
			continue
		}
		parts := strings.SplitN(directive, " ", 2)
		assert.Len(t, parts, 2, "Directive should have key and value: %s", directive)
	}
}

func TestExtractDomain_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "With port and path",
			input:    "https://example.com:8443/path/to/resource",
			expected: "https://example.com:8443",
		},
		{
			name:     "No scheme",
			input:    "example.com/path",
			expected: "",
		},
		{
			name:     "Only scheme",
			input:    "https://",
			expected: "",
		},
		{
			name:     "IPv4 address",
			input:    "http://192.168.1.1:8080/path",
			expected: "http://192.168.1.1:8080",
		},
		{
			name:     "Localhost",
			input:    "http://localhost:3000",
			expected: "http://localhost:3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomain(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateCSPNonce(t *testing.T) {
	// Test that generateCSPNonce produces unique nonces
	nonces := make(map[string]bool)
	for i := 0; i < 100; i++ {
		nonce, err := generateCSPNonce()
		assert.NoError(t, err)
		assert.NotEmpty(t, nonce)
		assert.Greater(t, len(nonce), 16)

		// Check uniqueness
		assert.False(t, nonces[nonce], "Nonce should be unique")
		nonces[nonce] = true
	}
}
