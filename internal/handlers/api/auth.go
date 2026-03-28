// Package api contains JSON API handlers for the SvelteKit frontend.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"savvy/internal/email"
	"savvy/internal/metrics"
	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/services"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

// Pre-compiled regexes for password complexity validation.
var (
	reUpper   = regexp.MustCompile(`[A-Z]`)
	reLower   = regexp.MustCompile(`[a-z]`)
	reDigit   = regexp.MustCompile(`[0-9]`)
	reSpecial = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`)
)

// dummyPasswordHash is a pre-generated bcrypt hash used for timing-attack protection.
// This is generated at startup to ensure constant-time comparison even when user doesn't exist.
var dummyPasswordHash string

func init() {
	// Generate a real bcrypt hash at startup for timing-attack protection
	// This ensures that bcrypt comparison takes the same time whether user exists or not
	// Using cost 12 for enhanced security (recommended for production)
	hash, err := bcrypt.GenerateFromPassword(
		[]byte("dummy-constant-password-for-timing-safety-protection"),
		12, // Cost 12 for enhanced security
	)
	if err != nil {
		panic("CRITICAL: failed to generate dummy bcrypt hash: " + err.Error())
	}
	dummyPasswordHash = string(hash)
}

// AuthHandler handles authentication API endpoints.
type AuthHandler struct {
	userService       services.UserServiceInterface
	emailTokenService services.EmailTokenServiceInterface
	emailService      email.ServiceInterface
	totpService       services.TOTPServiceInterface
	sessionService    services.SessionServiceInterface
	frontendURL       string
}

// NewAuthHandler creates a new auth API handler.
func NewAuthHandler(
	userService services.UserServiceInterface,
	emailTokenService services.EmailTokenServiceInterface,
	emailService email.ServiceInterface,
	frontendURL string,
) *AuthHandler {
	return &AuthHandler{
		userService:       userService,
		emailTokenService: emailTokenService,
		emailService:      emailService,
		frontendURL:       frontendURL,
	}
}

// SetTOTPService sets the TOTP service (optional, only when 2FA is enabled)
func (h *AuthHandler) SetTOTPService(totpService services.TOTPServiceInterface) {
	h.totpService = totpService
}

// SetSessionService sets the session service (for revoking sessions on password reset).
func (h *AuthHandler) SetSessionService(sessionService services.SessionServiceInterface) {
	h.sessionService = sessionService
}

// Login handles login via JSON API
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Email and password are required",
		})
	}

	// Normalize email to lowercase
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	// Reject passwords exceeding bcrypt's 72-byte limit to prevent:
	// 1. Silent truncation (two different passwords with same 72-byte prefix match)
	// 2. CPU exhaustion (bcrypt is intentionally slow on long inputs)
	if len(password) > 72 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "password_too_long",
			Message: "Password must not exceed 72 characters",
		})
	}

	user, err := h.userService.GetUserByEmail(c.Request().Context(), email)

	// Always run bcrypt comparison, even if user doesn't exist
	// This prevents timing attacks that reveal whether an email exists
	var passwordHash string
	if err != nil {
		// User not found - use the pre-generated dummy hash to maintain constant time
		passwordHash = dummyPasswordHash
	} else {
		passwordHash = user.PasswordHash

		// Check if account is locked
		if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
			return c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "account_locked",
				Message: "Account temporarily locked due to multiple failed login attempts. Please try again later.",
			})
		}
	}

	// Always perform bcrypt comparison (constant time operation)
	bcryptErr := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))

	// Only proceed if both user exists AND password matches
	if err != nil || bcryptErr != nil {
		metrics.RecordLoginAttempt("failure")

		// Increment failed login attempts for existing users
		if user != nil {
			user.FailedLoginAttempts++
			if user.FailedLoginAttempts >= 5 {
				lockUntil := time.Now().Add(15 * time.Minute)
				user.LockedUntil = &lockUntil
			}
			if updateErr := h.userService.UpdateUser(c.Request().Context(), user); updateErr != nil {
				slog.Error("Failed to update failed login attempts", "email", user.Email, "error", updateErr) // #nosec G706 -- structured logging, not injectable
			}
		}

		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid email or password",
		})
	}

	metrics.RecordLoginAttempt("success")

	// Reset failed login attempts on successful login
	if user.FailedLoginAttempts > 0 || user.LockedUntil != nil {
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
		if updateErr := h.userService.UpdateUser(c.Request().Context(), user); updateErr != nil {
			slog.Error("Failed to reset login attempts", "email", user.Email, "error", updateErr) // #nosec G706 -- structured logging, not injectable
		}
	}

	// Check if user has 2FA enabled
	if h.totpService != nil {
		totpEnabled, totpErr := h.totpService.IsEnabled(c.Request().Context(), user.ID)
		if totpErr != nil {
			slog.Error("Failed to check 2FA status", "user_id", user.ID, "error", totpErr) // #nosec G706 -- structured logging, not injectable
		}

		if totpEnabled {
			// Create a partial session with pending 2FA
			if _, sessionErr := middleware.Create2FAPendingSession(c, user.ID.String()); sessionErr != nil {
				slog.Error("Failed to create 2FA pending session", "error", sessionErr)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "session_error",
					Message: "Failed to create session",
				})
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"requires_2fa": true,
			})
		}
	}

	// Create authenticated session (regenerates to prevent session fixation)
	if _, err := middleware.CreateUserSession(c, user.ID.String()); err != nil {
		slog.Error("Failed to create session", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "session_error",
			Message: "Failed to create session",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": ToUserDTO(user),
	})
}

// Register handles user registration via JSON API
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "All fields are required",
		})
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Validate password complexity
	if errCode, err := validatePassword(req.Password); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   errCode,
			Message: err.Error(),
		})
	}

	// Check if user already exists (generic error to prevent email enumeration)
	existingUser, _ := h.userService.GetUserByEmail(c.Request().Context(), email)
	if existingUser != nil {
		// Perform bcrypt comparison to prevent timing attacks
		_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(req.Password))
		return c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "registration_failed",
			Message: "Registration could not be completed",
		})
	}

	// Hash password with cost 12 for enhanced security
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create user",
		})
	}

	// Create user
	user := &models.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         "user",
		AuthProvider: "local",
	}

	if err := h.userService.CreateUser(c.Request().Context(), user); err != nil {
		slog.Error("Failed to create user", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create user",
		})
	}

	// Send verification email asynchronously (non-blocking)
	// Use background context since the request context is cancelled after the response is sent
	if h.emailService != nil && h.emailTokenService != nil {
		go func() {
			bgCtx := context.Background()
			token, tokenErr := h.emailTokenService.CreateVerificationToken(bgCtx, user.ID)
			if tokenErr != nil {
				slog.Error("Failed to create verification token", "email", user.Email, "error", tokenErr)
				return
			}

			verifyURL := fmt.Sprintf("%s/verify-email?token=%s", h.frontendURL, token)
			displayName := user.FirstName
			if displayName == "" {
				displayName = user.Email
			}

			if sendErr := h.emailService.SendEmailVerification(bgCtx, user.Email, displayName, verifyURL, user.Language); sendErr != nil {
				slog.Error("Failed to send verification email", "email", user.Email, "error", sendErr)
			}
		}()
	}

	// Auto-login after registration: create session
	if _, err := middleware.CreateUserSession(c, user.ID.String()); err != nil {
		slog.Error("Failed to create session", "error", err)
		// Return user anyway, just without session
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"user": ToUserDTO(user),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user": ToUserDTO(user),
	})
}

// Logout handles logout via JSON API
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *echo.Context) error {
	if err := middleware.DestroySession(c); err != nil {
		slog.Error("Failed to destroy session", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "session_error",
			Message: "Failed to clear session",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out"})
}

// Me returns the current authenticated user
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *echo.Context) error {
	// RequireAuth middleware sets current_user in context
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	// Check if impersonating
	isImpersonating := false
	if val, ok := c.Get("is_impersonating").(bool); ok {
		isImpersonating = val
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":               user.ID.String(),
			"email":            user.Email,
			"first_name":       user.FirstName,
			"last_name":        user.LastName,
			"is_admin":         user.IsAdmin(),
			"is_impersonating": isImpersonating,
			"email_verified":   user.EmailVerified,
			"auth_provider":    user.AuthProvider,
			"language":         user.Language,
		},
	})
}

// validatePassword validates password complexity requirements
func validatePassword(password string) (string, error) {
	if len(password) < 12 {
		return "weak_password", &ValidationError{Message: "Password must be at least 12 characters long"}
	}
	if len(password) > 72 {
		return "password_too_long", &ValidationError{Message: "Password must not exceed 72 characters"}
	}

	hasUpper := reUpper.MatchString(password)
	hasLower := reLower.MatchString(password)
	hasDigit := reDigit.MatchString(password)
	hasSpecial := reSpecial.MatchString(password)

	complexity := 0
	if hasUpper {
		complexity++
	}
	if hasLower {
		complexity++
	}
	if hasDigit {
		complexity++
	}
	if hasSpecial {
		complexity++
	}

	if complexity < 3 {
		return "weak_password", &ValidationError{Message: "Password must contain at least 3 of: uppercase letters, lowercase letters, digits, special characters"}
	}

	return "", nil
}

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

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
