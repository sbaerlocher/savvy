// Package middleware provides Echo middleware for session tracking and metrics.
package middleware

import (
	"savvy/internal/metrics"
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
		userID := sess.Values["user_id"]
		if userID != nil {
			sessionID, ok := userID.(string)
			if ok && sessionID != "" {
				// Track this session
				sessionsMutex.Lock()
				if !activeSessions[sessionID] {
					activeSessions[sessionID] = true
					metrics.IncrementActiveSessions()
				}
				sessionsMutex.Unlock()
			}
		}

		return next(c)
	}
}

// CleanupInactiveSessions removes sessions that haven't been seen in a while
// This should be called periodically (e.g., every 5 minutes)
func CleanupInactiveSessions() {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	// Clear all and let active users re-add themselves
	count := len(activeSessions)
	activeSessions = make(map[string]bool)

	// Reset the gauge to 0
	metrics.SetActiveSessions(0)

	// Sessions will be re-counted as users make requests
	_ = count
}
