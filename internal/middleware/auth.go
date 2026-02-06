// Package middleware contains Echo middleware for authentication, sessions, and observability.
package middleware

import (
	"context"
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the key for storing user in context
	UserContextKey ContextKey = "user"
)

// SetCurrentUser middleware loads the current user into context
func SetCurrentUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := GetSession(c)
		if err != nil {
			return next(c)
		}

		userIDStr, ok := session.Values["user_id"].(string)
		if !ok || userIDStr == "" {
			return next(c)
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return next(c)
		}

		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			return next(c)
		}

		c.Set("current_user", &user)

		// Also set user in request context for template helpers (barcode token generation)
		ctx := context.WithValue(c.Request().Context(), UserContextKey, &user)
		c.SetRequest(c.Request().WithContext(ctx))

		// Check if impersonating
		if impersonatedBy, ok := session.Values["impersonated_by"].(string); ok && impersonatedBy != "" {
			c.Set("is_impersonating", true)
		}

		return next(c)
	}
}

// RequireAuth middleware requires authentication
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		currentUser := c.Get("current_user")
		if currentUser == nil {
			session, _ := GetSession(c)
			session.AddFlash("Bitte melden Sie sich zuerst an", "danger")
			if err := session.Save(c.Request(), c.Response()); err != nil {
				return c.String(http.StatusInternalServerError, "Failed to save session")
			}
			return c.Redirect(http.StatusSeeOther, "/auth/login")
		}
		return next(c)
	}
}

// RequireAdmin middleware requires admin role
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		currentUser := c.Get("current_user")
		if currentUser == nil {
			session, _ := GetSession(c)
			session.AddFlash("Bitte melden Sie sich zuerst an", "danger")
			if err := session.Save(c.Request(), c.Response()); err != nil {
				return c.String(http.StatusInternalServerError, "Failed to save session")
			}
			return c.Redirect(http.StatusSeeOther, "/auth/login")
		}

		user, ok := currentUser.(*models.User)
		if !ok || !user.IsAdmin() {
			session, _ := GetSession(c)
			session.AddFlash("Sie benötigen Admin-Rechte für diese Seite", "danger")
			if err := session.Save(c.Request(), c.Response()); err != nil {
				return c.String(http.StatusInternalServerError, "Failed to save session")
			}
			return c.Redirect(http.StatusSeeOther, "/")
		}

		return next(c)
	}
}
