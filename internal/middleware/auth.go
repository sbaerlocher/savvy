// Package middleware contains Echo middleware for authentication, sessions, and observability.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"savvy/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// UserService defines the interface needed by auth middleware
type UserService interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the key for storing user in context
	UserContextKey ContextKey = "user"
)

// SetCurrentUserWithService creates a middleware that loads the current user using UserService
func SetCurrentUserWithService(userService UserService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session, err := GetSession(c)
			if err != nil {
				return next(c)
			}

			userIDStr := GetSessionUserID(session)
			if userIDStr == "" {
				return next(c)
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return next(c)
			}

			// Use UserService instead of direct database access
			user, err := userService.GetUserByID(c.Request().Context(), userID)
			if err != nil {
				slog.Error("Failed to load user from session", // #nosec G706 -- structured logging, not format string injection
					"user_id", userID,
					"error", err,
				)
				// Continue without user — downstream RequireAuth will return 401.
				// This is fail-open by design: public routes must still work when the DB hiccups,
				// and protected routes are guarded by RequireAuth / RequireAdmin.
				return next(c)
			}

			// Check if session was created before last password change (invalidate stale sessions)
			if user.PasswordChangedAt != nil {
				sessionCreatedAt := GetSessionCreatedAt(session)
				if sessionCreatedAt == 0 || time.Unix(sessionCreatedAt, 0).Before(*user.PasswordChangedAt) {
					// Session predates password change — invalidate it
					session.Options.MaxAge = -1
					_ = SaveSession(c, session)
					slog.Info("Invalidated stale session after password change") //nolint:gosec // no tainted data
					return next(c)                                               // Continue without user; RequireAuth will 401
				}
			}

			c.Set("current_user", user)

			// Periodically update session LastActiveAt (throttled to reduce DB writes)
			if lastActiveUnix, ok := session.Values[sessionLastActiveKey].(int64); ok {
				if time.Since(time.Unix(lastActiveUnix, 0)) >= lastActiveThrottle {
					if err := SaveSession(c, session); err != nil {
						slog.Error("Failed to update session last active", "error", err)
					} else {
						session.Values[sessionLastActiveKey] = time.Now().Unix()
					}
				}
			}

			// Also set user in request context for template helpers (barcode token generation)
			ctx := context.WithValue(c.Request().Context(), UserContextKey, user)
			c.SetRequest(c.Request().WithContext(ctx))

			// Check if impersonating
			if impersonatedBy := GetSessionImpersonatedBy(session); impersonatedBy != "" {
				c.Set("is_impersonating", true)
			}

			return next(c)
		}
	}
}

// RequireAuth middleware requires authentication
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		currentUser := c.Get("current_user")
		if currentUser == nil {
			// Check if this is an API request
			isAPI := len(c.Request().URL.Path) >= 4 && c.Request().URL.Path[:4] == "/api"

			if isAPI {
				// For API routes: return JSON 401
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "Not authenticated",
				})
			}

			// For HTML routes: redirect to login page
			session, _ := GetSession(c)
			session.AddFlash("Bitte melden Sie sich zuerst an", "danger")
			if err := SaveSession(c, session); err != nil {
				return c.String(http.StatusInternalServerError, "Failed to save session")
			}
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		return next(c)
	}
}

// RequireAdmin middleware requires admin role
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check if this is an API request
		isAPI := len(c.Request().URL.Path) >= 4 && c.Request().URL.Path[:4] == "/api"

		currentUser := c.Get("current_user")
		if currentUser == nil {
			if isAPI {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "Not authenticated",
				})
			}
			session, _ := GetSession(c)
			session.AddFlash("Bitte melden Sie sich zuerst an", "danger")
			if err := SaveSession(c, session); err != nil {
				return c.String(http.StatusInternalServerError, "Failed to save session")
			}
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		user, ok := currentUser.(*models.User)
		if !ok || !user.IsAdmin() {
			if isAPI {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "forbidden",
					"message": "Admin access required",
				})
			}
			session, _ := GetSession(c)
			session.AddFlash("Sie benötigen Admin-Rechte für diese Seite", "danger")
			if err := SaveSession(c, session); err != nil {
				return c.String(http.StatusInternalServerError, "Failed to save session")
			}
			return c.Redirect(http.StatusSeeOther, "/")
		}

		return next(c)
	}
}
