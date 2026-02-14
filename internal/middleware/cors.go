// Package middleware contains HTTP middleware for the savvy system.
package middleware

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// CORSConfig holds CORS middleware configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
}

// DefaultCORSConfig returns CORS configuration for development.
// allowOrigins specifies allowed origins (from CORS_ALLOWED_ORIGINS env var).
func DefaultCORSConfig(allowOrigins []string) CORSConfig {
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	}
	return CORSConfig{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Content-Type", "Authorization", CSRFHeaderName},
		AllowCredentials: true, // Required for session cookies
	}
}

// CORSMiddleware creates CORS middleware with the given configuration
// IMPORTANT: Only use in development! In production, use reverse proxy (Traefik) for CORS.
func CORSMiddleware(config CORSConfig) echo.MiddlewareFunc {
	// Warn about insecure wildcard + credentials combination at configuration time
	if config.AllowCredentials {
		for _, origin := range config.AllowOrigins {
			if origin == "*" {
				slog.Warn("CORS misconfiguration: wildcard origin '*' with AllowCredentials is insecure, wildcard will be ignored")
				break
			}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			origin := c.Request().Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range config.AllowOrigins {
				// Skip wildcard when credentials are enabled (CORS spec violation)
				if allowedOrigin == "*" && config.AllowCredentials {
					continue
				}
				if origin == allowedOrigin || allowedOrigin == "*" {
					allowed = true
					break
				}
			}

			if allowed {
				// Set CORS headers
				c.Response().Header().Set("Access-Control-Allow-Origin", origin)

				if config.AllowCredentials {
					c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
				}

				// Handle preflight (OPTIONS) requests
				if c.Request().Method == http.MethodOptions {
					c.Response().Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowMethods, ", "))
					c.Response().Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders, ", "))
					c.Response().Header().Set("Access-Control-Max-Age", "86400") // 24 hours
					return c.NoContent(http.StatusNoContent)
				}
			}

			return next(c)
		}
	}
}

// joinStrings is a simple string join helper
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
