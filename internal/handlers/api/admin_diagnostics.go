// Package api contains JSON API handlers for the SvelteKit frontend.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"log/slog"
	"net/http"

	"savvy/internal/email"
	"savvy/internal/logsafe"
	"savvy/internal/models"

	"github.com/labstack/echo/v5"
)

// GetSystemHealth returns the current system health status
// GET /api/v1/admin/system-health
func (h *AdminHandler) GetSystemHealth(c *echo.Context) error {
	report, err := h.healthService.CheckReadiness(c.Request().Context())
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to check system health", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to check system health",
		})
	}

	return c.JSON(http.StatusOK, report)
}

// SendTestEmail sends a test email to the requesting admin
// POST /api/v1/admin/test-email
func (h *AdminHandler) SendTestEmail(c *echo.Context) error {
	// Get current user from context
	currentUser := c.Get("current_user")
	if currentUser == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	user, ok := currentUser.(*models.User)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to get user information",
		})
	}

	ctx := c.Request().Context()

	// Send test email using EmailService with user's preferred language
	if err := h.emailService.SendTestEmail(ctx, user.Email, user.FirstName, user.Language); err != nil {
		slog.ErrorContext(ctx, "failed to send test email", "email", user.Email, "error", err)
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "smtp_error",
			Message: "Failed to send test email. Check server logs for details.",
		})
	}

	slog.InfoContext(ctx, "test email sent", "email", user.Email, "user_id", user.ID, "language", user.Language)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Test email sent successfully! Check your inbox.",
	})
}

// SendTestPush sends a test push notification to the requesting admin
// POST /api/v1/admin/test-push
func (h *AdminHandler) SendTestPush(c *echo.Context) error {
	currentUser := c.Get("current_user")
	if currentUser == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	user, ok := currentUser.(*models.User)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to get user information",
		})
	}

	if h.pushService == nil {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "push_not_configured",
			Message: "Push notification service is not configured",
		})
	}

	ctx := c.Request().Context()

	if err := h.pushService.SendTestPush(ctx, user.ID); err != nil {
		slog.ErrorContext(ctx, "failed to send test push", "email", user.Email, "user_id", user.ID, "error", err)
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "push_error",
			Message: "Failed to send test push notification. Check server logs for details.",
		})
	}

	slog.InfoContext(ctx, "test push sent", "email", user.Email, "user_id", user.ID)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Test push notification sent! Check your browser notifications.",
	})
}

// SendPreviewEmail sends a preview of a specific email template to the requesting admin (dev only)
// POST /api/v1/admin/preview-email
func (h *AdminHandler) SendPreviewEmail(c *echo.Context) error {
	currentUser := c.Get("current_user")
	if currentUser == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	user, ok := currentUser.(*models.User)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to get user information",
		})
	}

	var req struct {
		Template string `json:"template"`
		Language string `json:"language,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	ctx := c.Request().Context()
	lang := req.Language
	if lang == "" {
		lang = user.Language
	}

	var err error

	switch req.Template {
	case "test_email":
		err = h.emailService.SendTestEmail(ctx, user.Email, user.FirstName, lang)
	case "password_reset":
		err = h.emailService.SendPasswordReset(ctx, user.Email, user.FirstName,
			"https://example.com/reset-password?token=sample-preview-token", "1 hour", lang)
	case "email_verification":
		err = h.emailService.SendEmailVerification(ctx, user.Email, user.FirstName,
			"https://example.com/verify-email?token=sample-preview-token", lang)
	case "account_deleted":
		err = h.emailService.SendAccountDeletionConfirmation(ctx, user.Email, user.FirstName, lang)
	case "expiry_reminder":
		err = h.emailService.SendExpiryReminder(ctx, user.Email, user.FirstName, email.ExpiryReminderData{
			ResourceType: "voucher",
			MerchantName: "IKEA",
			Code:         "SAVE-20-PERCENT",
			Value:        "20%",
			DaysLeft:     3,
			ExpiresAt:    "28. February 2026",
			ResourceURL:  "https://example.com/vouchers/sample-uuid",
		}, "https://example.com/unsubscribe?token=sample-preview-token&type=reminders", lang)
	case "share_notification":
		err = h.emailService.SendShareNotification(ctx, user.Email, user.FirstName, "Max Muster", "voucher", "IKEA", "20% auf alles", 50, "CHF", "https://example.com/vouchers/sample-uuid", "https://example.com/unsubscribe?token=sample-preview-token&type=notifications", lang)
	case "transfer_notification":
		err = h.emailService.SendTransferNotification(ctx, user.Email, user.FirstName, "Max Muster", "gift_card", "IKEA", "Geburtstagskarte", 100, "CHF", "https://example.com/gift-cards/sample-uuid", "https://example.com/unsubscribe?token=sample-preview-token&type=notifications", lang)
	case "validity_start":
		err = h.emailService.SendValidityStart(ctx, user.Email, user.FirstName, email.ValidityStartData{
			MerchantName: "IKEA",
			ValidFrom:    "1. March 2026",
			Code:         "WELCOME-2026",
			Value:        "CHF 50.00",
			ResourceURL:  "https://example.com/vouchers/sample-uuid",
		}, "https://example.com/unsubscribe?token=sample-preview-token&type=reminders", lang)
	default:
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_template",
			Message: "Unknown email template: " + req.Template,
		})
	}

	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to send preview email", "template", logsafe.String(req.Template), "email", logsafe.String(user.Email), "error", logsafe.Error(err))
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "smtp_error",
			Message: "Failed to send preview email. Check server logs for details.",
		})
	}

	slog.InfoContext(c.Request().Context(), "preview email sent", "template", logsafe.String(req.Template), "email", logsafe.String(user.Email), "language", logsafe.String(lang))

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Preview email sent successfully! Check your inbox.",
	})
}
