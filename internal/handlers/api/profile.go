// Package api contains JSON API handlers for the SvelteKit frontend.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"log/slog"
	"net/http"
	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/services"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// ProfileHandler handles user profile API endpoints.
type ProfileHandler struct {
	userService    services.UserServiceInterface
	accountService services.AccountServiceInterface
	sessionService services.SessionServiceInterface
}

// NewProfileHandler creates a new profile API handler.
func NewProfileHandler(userService services.UserServiceInterface, accountService services.AccountServiceInterface) *ProfileHandler {
	return &ProfileHandler{
		userService:    userService,
		accountService: accountService,
	}
}

// SetSessionService sets the session service (for revoking sessions on password change).
func (h *ProfileHandler) SetSessionService(sessionService services.SessionServiceInterface) {
	h.sessionService = sessionService
}

// GetProfile returns the authenticated user's profile
// GET /api/v1/profile
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"profile": map[string]interface{}{
			"id":                          user.ID.String(),
			"email":                       user.Email,
			"first_name":                  user.FirstName,
			"last_name":                   user.LastName,
			"language":                    user.Language,
			"auth_provider":               user.AuthProvider,
			"push_notifications_enabled":  user.PushNotificationsEnabled,
			"email_notifications_enabled": user.EmailNotificationsEnabled,
			"push_reminders_enabled":      user.PushRemindersEnabled,
			"push_sharing_enabled":        user.PushSharingEnabled,
			"email_reminders_enabled":     user.EmailRemindersEnabled,
			"email_sharing_enabled":       user.EmailSharingEnabled,
			"email_verified":              user.EmailVerified,
			"email_verified_at":           FormatTimePtr(user.EmailVerifiedAt),
			"created_at":                  FormatTime(user.CreatedAt),
			"updated_at":                  FormatTime(user.UpdatedAt),
		},
	})
}

// UpdateProfile updates the authenticated user's name and language
// PATCH /api/v1/profile
func (h *ProfileHandler) UpdateProfile(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	var req struct {
		FirstName                 *string `json:"first_name"`
		LastName                  *string `json:"last_name"`
		Language                  *string `json:"language"`
		PushNotificationsEnabled  *bool   `json:"push_notifications_enabled"`
		EmailNotificationsEnabled *bool   `json:"email_notifications_enabled"`
		PushRemindersEnabled      *bool   `json:"push_reminders_enabled"`
		PushSharingEnabled        *bool   `json:"push_sharing_enabled"`
		EmailRemindersEnabled     *bool   `json:"email_reminders_enabled"`
		EmailSharingEnabled       *bool   `json:"email_sharing_enabled"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Only local auth users can change their name (OAuth names are managed by the provider)
	if user.AuthProvider == "local" {
		if req.FirstName != nil {
			user.FirstName = *req.FirstName
		}
		if req.LastName != nil {
			user.LastName = *req.LastName
		}
	}
	if req.Language != nil {
		lang := *req.Language
		if lang == "de" || lang == "en" || lang == "fr" {
			user.Language = lang
		}
	}
	if req.PushNotificationsEnabled != nil {
		user.PushNotificationsEnabled = *req.PushNotificationsEnabled
	}
	if req.EmailNotificationsEnabled != nil {
		user.EmailNotificationsEnabled = *req.EmailNotificationsEnabled
	}
	if req.PushRemindersEnabled != nil {
		user.PushRemindersEnabled = *req.PushRemindersEnabled
	}
	if req.PushSharingEnabled != nil {
		user.PushSharingEnabled = *req.PushSharingEnabled
	}
	if req.EmailRemindersEnabled != nil {
		user.EmailRemindersEnabled = *req.EmailRemindersEnabled
	}
	if req.EmailSharingEnabled != nil {
		user.EmailSharingEnabled = *req.EmailSharingEnabled
	}

	if err := h.userService.UpdateUser(c.Request().Context(), user); err != nil {
		slog.Error("Failed to update profile", "user_id", user.ID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update profile",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"profile": map[string]interface{}{
			"id":                          user.ID.String(),
			"email":                       user.Email,
			"first_name":                  user.FirstName,
			"last_name":                   user.LastName,
			"language":                    user.Language,
			"auth_provider":               user.AuthProvider,
			"push_notifications_enabled":  user.PushNotificationsEnabled,
			"email_notifications_enabled": user.EmailNotificationsEnabled,
			"push_reminders_enabled":      user.PushRemindersEnabled,
			"push_sharing_enabled":        user.PushSharingEnabled,
			"email_reminders_enabled":     user.EmailRemindersEnabled,
			"email_sharing_enabled":       user.EmailSharingEnabled,
			"email_verified":              user.EmailVerified,
			"email_verified_at":           FormatTimePtr(user.EmailVerifiedAt),
			"created_at":                  FormatTime(user.CreatedAt),
			"updated_at":                  FormatTime(user.UpdatedAt),
		},
	})
}

// ChangePassword changes the authenticated user's password
// POST /api/v1/profile/change-password
func (h *ProfileHandler) ChangePassword(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	// Only local auth users can change password
	if user.AuthProvider != "local" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "oauth_user",
			Message: "OAuth users cannot change password",
		})
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Current password and new password are required",
		})
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_password",
			Message: "Current password is incorrect",
		})
	}

	// Validate new password
	if errCode, err := validatePassword(req.NewPassword); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   errCode,
			Message: err.Error(),
		})
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to change password",
		})
	}

	now := time.Now()
	user.PasswordHash = string(passwordHash)
	user.PasswordChangedAt = &now
	if err := h.userService.UpdateUser(c.Request().Context(), user); err != nil {
		slog.Error("Failed to update password", "user_id", user.ID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to change password",
		})
	}

	// Update session timestamp so the current session remains valid
	session, sessionErr := middleware.GetSession(c)
	if sessionErr == nil {
		session.Values[middleware.SessionKeySessionCreatedAt] = now.Unix()
		_ = middleware.SaveSession(c, session)
	}

	// Revoke all other sessions (server-side session store)
	if h.sessionService != nil {
		currentTokenHash := middleware.GetCurrentSessionTokenHash(c)
		if currentTokenHash != "" {
			if count, err := h.sessionService.RevokeOtherSessions(c.Request().Context(), user.ID, currentTokenHash); err != nil {
				slog.Error("Failed to revoke other sessions after password change", "error", err) //nolint:gosec // user.ID is internal UUID
			} else if count > 0 {
				slog.Info("Revoked other sessions after password change", "count", count) //nolint:gosec // count is internal int
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// DeleteAccount permanently deletes the authenticated user's account.
// POST /api/v1/profile/delete-account
func (h *ProfileHandler) DeleteAccount(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	var req struct {
		Password     string `json:"password"` // #nosec G117 -- password input field, not a secret
		Confirmation string `json:"confirmation"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Require "DELETE" confirmation
	if req.Confirmation != "DELETE" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "confirmation_required",
			Message: "Type DELETE to confirm account deletion",
		})
	}

	// Local auth: verify password
	if user.AuthProvider == "local" {
		if req.Password == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "password_required",
				Message: "Password is required for account deletion",
			})
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_password",
				Message: "Password is incorrect",
			})
		}
	}

	// Delete account
	if err := h.accountService.DeleteAccount(c.Request().Context(), user.ID); err != nil {
		slog.Error("Failed to delete account", "user_id", user.ID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete account",
		})
	}

	// Invalidate session
	_ = middleware.DestroySession(c)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Account deleted successfully",
	})
}
