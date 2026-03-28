// Package middleware contains HTTP middleware for the savvy system.
package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

// CSRFCookieName is the name of the CSRF cookie
const CSRFCookieName = "csrf_token"

// CSRFHeaderName is the name of the CSRF header
const CSRFHeaderName = "X-CSRF-Token"

// csrfSessionKey is the session key for the server-side CSRF token
const csrfSessionKey = "csrf_token"

// CSRFApiMiddleware implements session-bound CSRF protection for SPAs.
// The authoritative token is stored in the Gorilla session (server-side).
// A non-HttpOnly cookie mirrors the token so the SPA can read it.
// On mutations, the X-CSRF-Token header is validated against the session value
// using constant-time comparison.
// Token rotates automatically when the session is regenerated (login/logout).
func CSRFApiMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get or generate CSRF token from session (server-side source of truth)
		token, err := getOrGenerateSessionCSRFToken(c)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "csrf_generation_failed",
				"message": "Failed to generate security token",
			})
		}

		// Set CSRF cookie (non-HttpOnly so JavaScript can read it)
		cookie := &http.Cookie{
			Name:     CSRFCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: false, // Important: JS must be able to read this
			Secure:   c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https",
			SameSite: http.SameSiteLaxMode, // Lax for OAuth compatibility
			MaxAge:   86400,                // 24 hours
		}
		c.SetCookie(cookie)

		// For mutations (POST, PUT, PATCH, DELETE), validate CSRF token
		method := c.Request().Method
		if method == http.MethodPost || method == http.MethodPut ||
			method == http.MethodPatch || method == http.MethodDelete {

			// Get token from header
			headerToken := c.Request().Header.Get(CSRFHeaderName)

			// Constant-time comparison against session token (not cookie)
			if headerToken == "" || subtle.ConstantTimeCompare([]byte(headerToken), []byte(token)) != 1 {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "csrf_token_mismatch",
					"message": "CSRF token validation failed",
				})
			}
		}

		return next(c)
	}
}

// getOrGenerateSessionCSRFToken retrieves the CSRF token from the session or generates a new one.
// The session is the server-side source of truth — not the cookie.
func getOrGenerateSessionCSRFToken(c echo.Context) (string, error) {
	sess, err := GetSession(c)
	if err != nil {
		// gorilla/sessions returns a valid empty session even on decode errors
		// (e.g. stale cookies after secret rotation or DB reset).
		// Only fail if the session itself is nil.
		if sess == nil {
			return "", fmt.Errorf("failed to get session: %w", err)
		}
		slog.Debug("Session decode error (stale cookie), using new session", "error", err)
	}

	// Check for existing token in session
	if token, ok := sess.Values[csrfSessionKey].(string); ok && token != "" {
		return token, nil
	}

	// Generate new token and persist in session
	token, err := generateCSRFToken()
	if err != nil {
		return "", err
	}

	sess.Values[csrfSessionKey] = token
	if err := SaveSession(c, sess); err != nil {
		return "", fmt.Errorf("failed to save CSRF token to session: %w", err)
	}

	return token, nil
}

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		slog.Error("Failed to generate CSRF token", "error", err)
		return "", fmt.Errorf("CSRF token generation failed: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
