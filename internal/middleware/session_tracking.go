// Package middleware provides Echo middleware for session tracking.
// Note: Session metrics (active_sessions) removed - not fully implemented.
package middleware

import (
	"sync"

	"github.com/labstack/echo/v4"
)

var (
	// activeSessions tracks currently active user sessions
	activeSessions = make(map[string]bool)
	sessionsMutex  sync.RWMutex
)

// SessionTracking tracks active user sessions for metrics
func SessionTracking(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session using the existing GetSession function (returns "session")
		sess, err := GetSession(c)
		if err != nil {
			return next(c)
		}

		// Check if user is logged in
		if sessionID := GetSessionUserID(sess); sessionID != "" {
			// Track this session (metrics removed - see metrics.go)
			sessionsMutex.Lock()
			if !activeSessions[sessionID] {
				activeSessions[sessionID] = true
			}
			sessionsMutex.Unlock()
		}

		return next(c)
	}
}

// CleanupInactiveSessions removes sessions that haven't been seen in a while
// This should be called periodically (e.g., every 5 minutes)
// Note: Metrics removed - session tracking kept for future use
func CleanupInactiveSessions() {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	// Clear all and let active users re-add themselves
	count := len(activeSessions)
	activeSessions = make(map[string]bool)

	// Sessions will be re-counted as users make requests
	_ = count
}
