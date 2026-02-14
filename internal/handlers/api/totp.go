// Package api contains JSON API handlers for the SvelteKit frontend.
package api //nolint:revive // "api" is a meaningful package name for API handlers

import (
	"log/slog"
	"net/http"
	"time"

	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TOTPHandler handles TOTP 2FA API endpoints.
type TOTPHandler struct {
	totpService services.TOTPServiceInterface
}

// NewTOTPHandler creates a new TOTP API handler.
func NewTOTPHandler(totpService services.TOTPServiceInterface) *TOTPHandler {
	return &TOTPHandler{totpService: totpService}
}

// Setup initiates TOTP setup for the authenticated user.
// POST /api/v1/auth/2fa/setup
func (h *TOTPHandler) Setup(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	// Only local auth users can enable 2FA
	if user.AuthProvider != "local" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "oauth_user",
			Message: "2FA is not available for OAuth users",
		})
	}

	setup, err := h.totpService.GenerateSetup(c.Request().Context(), user.ID, user.Email)
	if err != nil {
		if err == services.ErrTOTPAlreadyEnabled {
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "already_enabled",
				Message: "Two-factor authentication is already enabled",
			})
		}
		slog.Error("Failed to generate TOTP setup", "user_id", user.ID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to set up two-factor authentication",
		})
	}

	return c.JSON(http.StatusOK, setup)
}

// Verify verifies a TOTP code and enables 2FA.
// POST /api/v1/auth/2fa/verify
func (h *TOTPHandler) Verify(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "TOTP code is required",
		})
	}

	if err := h.totpService.VerifyAndEnable(c.Request().Context(), user.ID, req.Code); err != nil {
		switch err {
		case services.ErrTOTPAlreadyEnabled:
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "already_enabled",
				Message: "Two-factor authentication is already enabled",
			})
		case services.ErrTOTPNotSetup:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "not_setup",
				Message: "Please initiate 2FA setup first",
			})
		case services.ErrTOTPInvalidCode:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_code",
				Message: "Invalid verification code",
			})
		default:
			slog.Error("Failed to verify TOTP", "user_id", user.ID, "error", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to verify code",
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Two-factor authentication enabled successfully",
		"enabled": true,
	})
}

// Challenge verifies a TOTP code during login (step 2 of 2FA login).
// POST /api/v1/auth/2fa/challenge
func (h *TOTPHandler) Challenge(c echo.Context) error {
	// Get the pending user ID from session
	session, err := middleware.GetSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "No pending 2FA session",
		})
	}

	pendingUserID := middleware.GetSession2FAPendingUserID(session)
	if pendingUserID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "No pending 2FA session",
		})
	}

	// Enforce 5-minute expiration on 2FA pending session
	const twoFATimeout = 5 * time.Minute
	createdAt := middleware.GetSession2FAPendingCreatedAt(session)
	if createdAt == 0 || time.Since(time.Unix(createdAt, 0)) > twoFATimeout {
		delete(session.Values, middleware.SessionKey2FAPendingUserID)
		delete(session.Values, middleware.SessionKey2FAPendingCreatedAt)
		_ = middleware.SaveSession(c, session)
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "session_expired",
			Message: "2FA session has expired, please log in again",
		})
	}

	var req struct {
		Code       string `json:"code"`
		BackupCode string `json:"backup_code"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Code == "" && req.BackupCode == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "TOTP code or backup code is required",
		})
	}

	userID, err := parseUUID(pendingUserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_session",
			Message: "Invalid session data",
		})
	}

	var valid bool
	if req.Code != "" {
		valid, err = h.totpService.Verify(c.Request().Context(), userID, req.Code)
	} else {
		valid, err = h.totpService.VerifyBackupCode(c.Request().Context(), userID, req.BackupCode)
	}

	if err != nil {
		slog.Error("Failed to verify 2FA challenge", "user_id", userID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to verify code",
		})
	}

	if !valid {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_code",
			Message: "Invalid verification code",
		})
	}

	// 2FA verified - create authenticated session (regenerates to prevent session fixation)
	if _, err := middleware.CreateUserSession(c, pendingUserID); err != nil {
		slog.Error("Failed to create session after 2FA", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "session_error",
			Message: "Failed to complete authentication",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Two-factor authentication verified",
		"authenticated": true,
	})
}

// Disable disables 2FA for the authenticated user.
// POST /api/v1/auth/2fa/disable
func (h *TOTPHandler) Disable(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "TOTP code is required to disable 2FA",
		})
	}

	if err := h.totpService.Disable(c.Request().Context(), user.ID, req.Code); err != nil {
		switch err {
		case services.ErrTOTPNotEnabled:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "not_enabled",
				Message: "Two-factor authentication is not enabled",
			})
		case services.ErrTOTPInvalidCode:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_code",
				Message: "Invalid verification code",
			})
		default:
			slog.Error("Failed to disable TOTP", "user_id", user.ID, "error", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to disable two-factor authentication",
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Two-factor authentication disabled",
		"enabled": false,
	})
}

// RegenerateBackupCodes generates new backup codes.
// POST /api/v1/auth/2fa/backup-codes
func (h *TOTPHandler) RegenerateBackupCodes(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "TOTP code is required",
		})
	}

	codes, err := h.totpService.RegenerateBackupCodes(c.Request().Context(), user.ID, req.Code)
	if err != nil {
		switch err {
		case services.ErrTOTPNotEnabled:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "not_enabled",
				Message: "Two-factor authentication is not enabled",
			})
		case services.ErrTOTPInvalidCode:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_code",
				Message: "Invalid verification code",
			})
		default:
			slog.Error("Failed to regenerate backup codes", "user_id", user.ID, "error", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to regenerate backup codes",
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"backup_codes": codes,
	})
}

// Status returns the 2FA status for the authenticated user.
// GET /api/v1/auth/2fa/status
func (h *TOTPHandler) Status(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	enabled, err := h.totpService.IsEnabled(c.Request().Context(), user.ID)
	if err != nil {
		slog.Error("Failed to check 2FA status", "user_id", user.ID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to check 2FA status",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"enabled":       enabled,
		"is_local_auth": user.AuthProvider == "local",
	})
}

// parseUUID is a helper to parse UUID strings
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
