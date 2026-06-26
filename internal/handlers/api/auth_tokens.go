// Package api contains JSON API handlers for the SvelteKit frontend.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"savvy/internal/models"
	"savvy/internal/services"

	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

// RequestVerification sends a new verification email to the authenticated user
// POST /api/v1/auth/request-verification
func (h *AuthHandler) RequestVerification(c *echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	// Already verified
	if user.EmailVerified {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":        "Email already verified",
			"email_verified": true,
		})
	}

	// Create token and send email
	token, err := h.emailTokenService.CreateVerificationToken(c.Request().Context(), user.ID)
	if err != nil {
		slog.Error("Failed to create verification token", "user_id", user.ID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create verification token",
		})
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", h.frontendURL, token)
	displayName := user.FirstName
	if displayName == "" {
		displayName = user.Email
	}

	if err := h.emailService.SendEmailVerification(c.Request().Context(), user.Email, displayName, verifyURL, user.Language); err != nil {
		slog.Error("Failed to send verification email", "user_id", user.ID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to send verification email",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Verification email sent",
	})
}

// VerifyEmail verifies a user's email address using a token
// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *echo.Context) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Token is required",
		})
	}

	err := h.emailTokenService.VerifyEmail(c.Request().Context(), req.Token)
	if err != nil {
		switch err {
		case services.ErrTokenNotFound:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or expired verification token",
			})
		case services.ErrTokenExpired:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "token_expired",
				Message: "Verification token has expired. Please request a new one.",
			})
		case services.ErrTokenUsed:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "token_used",
				Message: "This verification token has already been used",
			})
		default:
			slog.Error("Failed to verify email", "error", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to verify email",
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":        "Email verified successfully",
		"email_verified": true,
	})
}

// UnsubscribeNotifications handles one-click email unsubscribe via token
// POST /api/v1/auth/unsubscribe-notifications
func (h *AuthHandler) UnsubscribeNotifications(c *echo.Context) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Token is required",
		})
	}

	err := h.emailTokenService.UnsubscribeNotifications(c.Request().Context(), req.Token)
	if err != nil {
		switch err {
		case services.ErrTokenNotFound:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or expired unsubscribe token",
			})
		case services.ErrTokenExpired:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "token_expired",
				Message: "Unsubscribe token has expired",
			})
		case services.ErrTokenUsed:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "token_used",
				Message: "This unsubscribe token has already been used",
			})
		default:
			slog.Error("Failed to unsubscribe notifications", "error", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to unsubscribe",
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":               "Successfully unsubscribed from notification emails",
		"email_sharing_enabled": false,
	})
}

// UnsubscribeReminders handles one-click unsubscribe from expiry reminder emails
// POST /api/v1/auth/unsubscribe-reminders
func (h *AuthHandler) UnsubscribeReminders(c *echo.Context) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Token is required",
		})
	}

	err := h.emailTokenService.UnsubscribeReminders(c.Request().Context(), req.Token)
	if err != nil {
		switch err {
		case services.ErrTokenNotFound:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or expired unsubscribe token",
			})
		case services.ErrTokenExpired:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "token_expired",
				Message: "Unsubscribe token has expired",
			})
		case services.ErrTokenUsed:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "token_used",
				Message: "This unsubscribe token has already been used",
			})
		default:
			slog.Error("Failed to unsubscribe reminders", "error", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to unsubscribe",
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":                 "Successfully unsubscribed from expiry reminder emails",
		"email_reminders_enabled": false,
	})
}

// RequestPasswordReset handles password reset requests
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) RequestPasswordReset(c *echo.Context) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Email is required",
		})
	}

	// Always return same response to prevent email enumeration (timing-attack protection)
	successResponse := map[string]string{
		"message": "If an account exists with this email, a password reset link has been sent.",
	}

	user, err := h.userService.GetUserByEmail(c.Request().Context(), email)
	if err != nil || user == nil {
		// Simulate work to prevent timing attacks
		_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte("timing-safe-dummy"))
		return c.JSON(http.StatusOK, successResponse)
	}

	// Only allow password reset for local auth users
	if user.AuthProvider != "local" {
		return c.JSON(http.StatusOK, successResponse)
	}

	token, err := h.emailTokenService.CreatePasswordResetToken(c.Request().Context(), user.ID)
	if err != nil {
		slog.Error("Failed to create password reset token", "user_id", user.ID, "error", err) // #nosec G706 -- structured logging, not injectable
		// Still return success to prevent enumeration
		return c.JSON(http.StatusOK, successResponse)
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.frontendURL, token)
	displayName := user.FirstName
	if displayName == "" {
		displayName = user.Email
	}

	if err := h.emailService.SendPasswordReset(c.Request().Context(), user.Email, displayName, resetURL, "1 hour", user.Language); err != nil {
		slog.Error("Failed to send password reset email", "user_id", user.ID, "error", err) // #nosec G706 -- structured logging, not injectable
	}

	return c.JSON(http.StatusOK, successResponse)
}

// ResetPassword handles setting a new password with a reset token
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *echo.Context) error {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"` // #nosec G117 -- struct field name, not a hardcoded secret
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Token == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Token and password are required",
		})
	}

	// Validate password complexity
	if errCode, err := validatePassword(req.Password); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   errCode,
			Message: err.Error(),
		})
	}

	// Consume the token and get the user
	user, err := h.emailTokenService.ConsumePasswordResetToken(c.Request().Context(), req.Token)
	if err != nil {
		switch err {
		case services.ErrTokenNotFound:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or expired password reset token",
			})
		case services.ErrTokenExpired:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "token_expired",
				Message: "Password reset token has expired. Please request a new one.",
			})
		case services.ErrTokenUsed:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "token_used",
				Message: "This password reset token has already been used",
			})
		default:
			slog.Error("Failed to consume password reset token", "error", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to reset password",
			})
		}
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to reset password",
		})
	}

	// Update password, reset lockout, and mark password change time
	now := time.Now()
	user.PasswordHash = string(passwordHash)
	user.PasswordChangedAt = &now
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil

	if err := h.userService.UpdateUser(c.Request().Context(), user); err != nil {
		slog.Error("Failed to update user password", "user_id", user.ID, "error", err) // #nosec G706 -- structured logging, not injectable
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to reset password",
		})
	}

	// Revoke all sessions for this user (force re-login everywhere)
	if h.sessionService != nil {
		if count, err := h.sessionService.RevokeAllSessions(c.Request().Context(), user.ID); err != nil {
			slog.Error("Failed to revoke sessions after password reset", "error", err) //nolint:gosec // user.ID is internal UUID
		} else if count > 0 {
			slog.Info("Revoked all sessions after password reset", "count", count) //nolint:gosec // count is internal int
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password has been reset successfully. You can now log in with your new password.",
	})
}
