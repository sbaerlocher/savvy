// Package middleware provides HTTP middleware for the Echo server.
package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"savvy/internal/config"
	"strings"

	"github.com/labstack/echo/v5"
)

// SecurityHeaders returns Echo middleware that sets security-related HTTP headers.
// For SvelteKit SPA (adapter-static), CSP uses 'unsafe-inline' as fallback
// because nonces cause browsers to ignore 'unsafe-inline', breaking Svelte.
func SecurityHeaders(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Build CSP for SvelteKit SPA (uses unsafe-inline)
			csp := buildContentSecurityPolicy(cfg)

			// Set security headers manually since we need dynamic CSP
			c.Response().Header().Set("Content-Security-Policy", csp)
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "SAMEORIGIN")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "camera=(self), microphone=(), geolocation=(), payment=()")

			// HSTS - Only in production
			if cfg.IsProduction() {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			return next(c)
		}
	}
}

// generateCSPNonce generates a cryptographically secure random nonce for CSP.
// NOTE: Currently unused for SvelteKit SPA mode because nonces cause browsers
// to ignore 'unsafe-inline', which breaks Svelte transitions and inline handlers.
// Kept for future use when SvelteKit moves to Web Animations API.
func generateCSPNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate CSP nonce: %w", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// buildContentSecurityPolicy constructs a CSP header value based on application config.
// For SvelteKit SPA (adapter-static), we use 'unsafe-inline' for both scripts and styles.
//
// Why not use nonces?
// - Browsers ignore 'unsafe-inline' when nonces/hashes are present
// - SvelteKit SPA requires 'unsafe-inline' for inline event handlers (onclick, etc.)
// - Svelte transitions create inline <style> elements that need 'unsafe-inline'
// - CSP config in svelte.config.js only works for SSR, not adapter-static
//
// This will be improved when SvelteKit moves to Web Animations API.
// See: https://github.com/sveltejs/kit/issues/11747
func buildContentSecurityPolicy(cfg *config.Config) string {
	// Base policy allowing resources from same origin
	policies := []string{
		"default-src 'self'",
		// TEMPORARY: Allow inline scripts/styles for SvelteKit SPA (SVL-004)
		// FUTURE: Implement nonce-based CSP when SvelteKit supports it
		// - Option 1: Switch to SvelteKit SSR mode (adapter-node) with CSP config
		// - Option 2: Wait for Web Animations API (removes inline styles)
		// See: https://github.com/sveltejs/kit/issues/11747
		// Added 'wasm-unsafe-eval' for barcode-detector WASM polyfill
		"script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval'",
		"style-src 'self' 'unsafe-inline'",
		// Allow images from same origin and data URIs (for barcodes)
		"img-src 'self' data:",
		// Allow fonts from same origin
		"font-src 'self'",
		// Prevent framing
		"frame-ancestors 'none'",
		// Upgrade insecure requests in production
	}

	// Add connect-src with OAuth domain if OAuth is enabled
	connectSrc := "'self'"
	if cfg.IsOAuthEnabled() {
		if oauthDomain := extractDomain(cfg.OAuthIssuer); oauthDomain != "" {
			connectSrc += " " + oauthDomain
		}
	}
	policies = append(policies, "connect-src "+connectSrc)

	// In production, upgrade all HTTP requests to HTTPS
	if cfg.IsProduction() {
		policies = append(policies, "upgrade-insecure-requests")
	}

	// Optional CSP violation reporting
	if cfg.CSPReportURI != "" {
		policies = append(policies, "report-uri "+cfg.CSPReportURI)
	}

	return strings.Join(policies, "; ")
}

// extractDomain extracts the domain (scheme + host) from a URL string.
// Returns empty string if parsing fails.
func extractDomain(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Return scheme + host (e.g., "https://auth.example.com")
	if u.Scheme != "" && u.Host != "" {
		return u.Scheme + "://" + u.Host
	}

	return ""
}
