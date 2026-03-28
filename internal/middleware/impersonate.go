// Package middleware provides HTTP middleware for authentication, authorization, and request processing.
package middleware

import (
	"net/http"
	"savvy/internal/models"

	"github.com/labstack/echo/v5"
)

// RequireImpersonationOrAdmin checks if the user is an admin OR currently impersonating another user.
// This middleware is used for routes that should be accessible to:
// - Admins (always)
// - Regular users (only when impersonating)
//
// Use case: Merchant CRUD routes should be admin-only normally, but accessible during impersonation
// for testing or support purposes.
//
// Example usage:
//
//	merchantsCRUD := protected.Group("/merchants", middleware.RequireImpersonationOrAdmin)
//	merchantsCRUD.GET("/new", handlers.MerchantsNew)
//	merchantsCRUD.POST("", handlers.MerchantsCreate)
//
// How it works:
//   - If user is admin (user.IsAdmin == true), allow access
//   - If user is NOT admin, check session for "original_user_id" AND "original_user_is_admin" keys
//   - Both must be set and the admin flag must be true → user is legitimately impersonating → allow access
//   - Otherwise, deny access with 403 Forbidden
func RequireImpersonationOrAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get current user from context (set by RequireAuth middleware)
		user, ok := c.Get("current_user").(*models.User)
		if !ok || user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
		}

		// Allow if user is admin
		if user.IsAdmin() {
			return next(c)
		}

		// Check if user is impersonating (session key: "original_user_id")
		sess, err := GetSession(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get session")
		}

		// Both original_user_id and original_user_is_admin must be present and the flag
		// must be true. This prevents a session-store bug from granting access to a
		// non-admin who has only the ID key set without the verified admin flag.
		if GetSessionOriginalUserID(sess) != "" && GetSessionOriginalUserIsAdmin(sess) {
			return next(c)
		}

		// Neither admin nor impersonating, deny access
		return echo.NewHTTPError(http.StatusForbidden, "Access denied. Admin or impersonation required.")
	}
}
