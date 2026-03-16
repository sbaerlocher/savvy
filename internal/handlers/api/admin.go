// Package api contains JSON API handlers for the SvelteKit frontend.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"log/slog"
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/email"
	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/repository"
	"savvy/internal/services"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	roleUser  = "user"
	roleAdmin = "admin"
)

// AdminHandler handles admin API endpoints
type AdminHandler struct {
	adminService  services.AdminServiceInterface
	userService   services.UserServiceInterface
	healthService services.HealthCheckServiceInterface
	emailService  email.ServiceInterface
	pushService   services.PushServiceInterface
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	adminService services.AdminServiceInterface,
	userService services.UserServiceInterface,
	healthService services.HealthCheckServiceInterface,
	emailService email.ServiceInterface,
	pushService services.PushServiceInterface,
) *AdminHandler {
	return &AdminHandler{
		adminService:  adminService,
		userService:   userService,
		healthService: healthService,
		emailService:  emailService,
		pushService:   pushService,
	}
}

// ListUsers returns all users
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c echo.Context) error {
	users, err := h.adminService.GetAllUsers(c.Request().Context())
	if err != nil {
		c.Logger().Errorf("Failed to load users: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to load users",
		})
	}

	return c.JSON(http.StatusOK, AdminUserListResponse{
		Users: ToAdminUserDTOs(users),
	})
}

// GetUser returns a single user by ID
// GET /api/v1/admin/users/:id
func (h *AdminHandler) GetUser(c echo.Context) error {
	userID, err := parseResourceID(c, "user")
	if err != nil {
		return err
	}

	user, err := h.adminService.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		c.Logger().Errorf("Failed to load user: %v", err)
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": ToAdminUserDTO(user),
	})
}

// CreateUser creates a new local auth user
// POST /api/v1/admin/users
func (h *AdminHandler) CreateUser(c echo.Context) error {
	var req AdminUserCreateRequest
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
			Message: "Email, password, first name and last name are required",
		})
	}

	// Validate password complexity
	if errCode, err := validatePassword(req.Password); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   errCode,
			Message: err.Error(),
		})
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user exists
	existingUser, err := h.userService.GetUserByEmail(c.Request().Context(), email)
	if err == nil && existingUser != nil {
		return c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "email_exists",
			Message: "User with this email already exists",
		})
	}

	// Hash password with cost 12 for enhanced security
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.Logger().Errorf("Failed to hash password: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to hash password",
		})
	}

	// Set default role if not provided
	role := roleUser
	if req.Role != nil && *req.Role != "" {
		if *req.Role != roleUser && *req.Role != roleAdmin {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_role",
				Message: "Role must be 'user' or 'admin'",
			})
		}
		role = *req.Role
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         role,
		AuthProvider: "local",
	}

	if err := h.adminService.CreateLocalUser(c.Request().Context(), user); err != nil {
		c.Logger().Errorf("Failed to create user: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create user",
		})
	}

	slog.InfoContext(c.Request().Context(), "user created successfully", "email", user.Email)

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User created successfully",
		"user":    ToAdminUserDTO(user),
	})
}

// UpdateUser updates a user (email, name, role)
// PATCH /api/v1/admin/users/:id
func (h *AdminHandler) UpdateUser(c echo.Context) error {
	userID, err := parseResourceID(c, "user")
	if err != nil {
		return err
	}

	var req AdminUserUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Get current user
	user, err := h.adminService.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
	}

	// Apply updates (use existing values if not provided)
	email := user.Email
	if req.Email != nil && *req.Email != "" {
		email = strings.ToLower(strings.TrimSpace(*req.Email))
	}

	firstName := user.FirstName
	if req.FirstName != nil {
		firstName = *req.FirstName
	}

	lastName := user.LastName
	if req.LastName != nil {
		lastName = *req.LastName
	}

	role := user.Role
	if req.Role != nil && *req.Role != "" {
		// OAuth users: role is managed by the OAuth provider (via OAUTH_ADMIN_EMAILS/
		// OAUTH_ADMIN_GROUP) and re-evaluated on every login.
		if user.IsOAuthUser() {
			return c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Error:   "role_readonly",
				Message: "Role of OAuth users is managed by the OAuth provider and cannot be changed",
			})
		}
		if *req.Role != roleUser && *req.Role != roleAdmin {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_role",
				Message: "Role must be 'user' or 'admin'",
			})
		}
		role = *req.Role
	}

	// Validate required fields
	if email == "" || firstName == "" || lastName == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Email, first name and last name are required",
		})
	}

	// Check if email is being changed and if it conflicts with another user
	if email != user.Email {
		existingUser, err := h.userService.GetUserByEmail(c.Request().Context(), email)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "email_exists",
				Message: "Another user with this email already exists",
			})
		}
	}

	// Update user
	if err := h.adminService.UpdateUser(c.Request().Context(), userID, email, firstName, lastName, role); err != nil {
		c.Logger().Errorf("Failed to update user: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update user",
		})
	}

	c.Logger().Infof("User updated successfully: %s", userID.String())

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User updated successfully",
	})
}

// GetAuditLogs returns paginated audit logs with filters
// GET /api/v1/admin/audit-log
func (h *AdminHandler) GetAuditLogs(c echo.Context) error {
	var req AuditLogFiltersRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
		})
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}

	// Parse user ID filter
	var userID *uuid.UUID
	if req.UserID != nil && *req.UserID != "" {
		uid, err := uuid.Parse(*req.UserID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_user_id",
				Message: "Invalid user ID format",
			})
		}
		userID = &uid
	}

	// Build filters for service
	filters := repository.AuditLogFilters{
		UserID:       userID,
		ResourceType: req.ResourceType,
		Action:       req.Action,
		DateFrom:     req.DateFrom,
		DateTo:       req.DateTo,
		SearchQuery:  req.SearchQuery,
		Page:         req.Page,
		PerPage:      req.PerPage,
	}

	result, err := h.adminService.GetAuditLogs(c.Request().Context(), filters)
	if err != nil {
		c.Logger().Errorf("Failed to load audit logs: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to load audit logs",
		})
	}

	totalPages := int((result.Total + int64(req.PerPage) - 1) / int64(req.PerPage))

	return c.JSON(http.StatusOK, AuditLogListResponse{
		Logs:       ToAuditLogDTOs(result.Logs),
		Total:      result.Total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: totalPages,
	})
}

// StartImpersonation starts user impersonation
// POST /api/v1/admin/users/:id/impersonate
func (h *AdminHandler) StartImpersonation(c echo.Context) error {
	// 1. Get current admin from context
	currentUser := c.Get("current_user")
	if currentUser == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	admin, ok := currentUser.(*models.User)
	if !ok || !admin.IsAdmin() {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Admin role required",
		})
	}

	// 2. Parse target user ID
	targetUserID, err := parseResourceID(c, "user")
	if err != nil {
		return err
	}

	// 3. Validate impersonation
	if err := h.adminService.ValidateImpersonation(c.Request().Context(), admin.ID, targetUserID); err != nil {
		c.Logger().Errorf("Impersonation validation failed for admin %s targeting %s: %v", admin.ID, targetUserID, err)
		status := http.StatusBadRequest
		msg := "Impersonation validation failed"
		if err.Error() == "only admins can impersonate" {
			status = http.StatusForbidden
			msg = "Only admins can impersonate users"
		} else if err.Error() == "cannot impersonate yourself" {
			msg = "Cannot impersonate yourself"
		} else if err.Error() == "target user not found" {
			msg = "Target user not found"
		}
		return c.JSON(status, ErrorResponse{
			Error:   "validation_failed",
			Message: msg,
		})
	}

	// 4. Get target user
	targetUser, err := h.adminService.GetUserByID(c.Request().Context(), targetUserID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "user_not_found",
			Message: "Target user not found",
		})
	}

	// 5. Create impersonation session (regenerates to prevent session fixation)
	if _, err := middleware.CreateImpersonationSession(c, targetUser.ID.String(), admin.ID.String()); err != nil {
		c.Logger().Errorf("Failed to create impersonation session: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "session_error",
			Message: "Failed to create impersonation session",
		})
	}

	// 7. Create audit log
	// Add audit context (user ID, IP address, user agent) for audit logging
	ctx := audit.AddAuditContextToContext(c.Request().Context(), admin.ID, c.RealIP(), c.Request().UserAgent())

	// Resource data for audit log
	resourceData := map[string]interface{}{
		"target_user_id":    targetUser.ID.String(),
		"target_user_email": targetUser.Email,
		"target_user_role":  targetUser.Role,
		"admin_email":       admin.Email,
	}

	if err := h.adminService.StartImpersonation(ctx, admin.ID, targetUser.ID, resourceData); err != nil {
		c.Logger().Errorf("Failed to create audit log: %v", err)
		// Don't fail the request, just log the error
	}

	c.Logger().Infof("Admin %s started impersonating user %s", admin.Email, targetUser.Email)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Impersonation started successfully",
		"user":    ToAdminUserDTO(targetUser),
	})
}

// StopImpersonation stops user impersonation
// POST /api/v1/admin/impersonate/stop
func (h *AdminHandler) StopImpersonation(c echo.Context) error {
	// 1. Get session
	session, err := middleware.GetSession(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "session_error",
			Message: "Failed to get session",
		})
	}

	// 2. Check if impersonating
	originalUserIDStr := middleware.GetSessionOriginalUserID(session)
	if originalUserIDStr == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "not_impersonating",
			Message: "Not currently impersonating",
		})
	}

	originalUserID, err := uuid.Parse(originalUserIDStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "invalid_session",
			Message: "Invalid original user ID in session",
		})
	}

	// 3. Get current impersonated user ID (for audit log)
	currentUserIDStr := middleware.GetSessionUserID(session)
	currentUserID, err := uuid.Parse(currentUserIDStr)
	if err != nil {
		c.Logger().Warnf("Failed to parse current user ID from session: %v", err)
	}

	// 4. Restore admin session (regenerates for security)
	if _, err := middleware.StopImpersonationSession(c, originalUserID.String()); err != nil {
		c.Logger().Errorf("Failed to stop impersonation: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "session_error",
			Message: "Failed to restore admin session",
		})
	}

	// 6. Create audit log
	if currentUserID != uuid.Nil {
		// Get user information for audit log (best effort - don't fail if user lookup fails)
		var resourceData map[string]interface{}
		if targetUser, err := h.adminService.GetUserByID(c.Request().Context(), currentUserID); err == nil {
			resourceData = map[string]interface{}{
				"target_user_id":    targetUser.ID.String(),
				"target_user_email": targetUser.Email,
				"target_user_role":  targetUser.Role,
			}
		} else {
			// Fallback to minimal data if user lookup fails
			resourceData = map[string]interface{}{
				"target_user_id": currentUserID.String(),
			}
		}

		// Add audit context (user ID, IP address, user agent) for audit logging
		ctx := audit.AddAuditContextToContext(c.Request().Context(), originalUserID, c.RealIP(), c.Request().UserAgent())
		if err := h.adminService.StopImpersonation(ctx, originalUserID, currentUserID, resourceData); err != nil {
			c.Logger().Errorf("Failed to create audit log: %v", err)
			// Don't fail the request
		}
	}

	c.Logger().Infof("Admin %s stopped impersonating", originalUserID.String())

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Impersonation stopped successfully",
	})
}

// RestoreResource restores a soft-deleted resource
// POST /api/v1/admin/restore/:resource_type/:resource_id
func (h *AdminHandler) RestoreResource(c echo.Context) error {
	resourceType := c.Param("resource_type")

	// Parse resource ID
	resourceID, err := parseUUIDParam(c, "resource_id", "invalid_resource_id", "Invalid resource ID format")
	if err != nil {
		return err
	}

	// Validate resource type
	validTypes := map[string]bool{
		"cards":                  true,
		"card_shares":            true,
		"vouchers":               true,
		"voucher_shares":         true,
		"gift_cards":             true,
		"gift_card_shares":       true,
		"gift_card_transactions": true,
		"merchants":              true,
	}

	if !validTypes[resourceType] {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_resource_type",
			Message: "Invalid resource type",
		})
	}

	// Restore resource
	if err := h.adminService.RestoreResource(c.Request().Context(), resourceType, resourceID); err != nil {
		if err.Error() == "resource is not deleted" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "resource_not_deleted",
				Message: "Resource is not deleted",
			})
		}

		slog.ErrorContext(c.Request().Context(), "failed to restore resource", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to restore resource",
		})
	}

	slog.InfoContext(c.Request().Context(), "admin restored resource", "resource_type", resourceType, "resource_id", resourceID.String())

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Resource restored successfully",
	})
}

// GetSystemHealth returns the current system health status
// GET /api/v1/admin/system-health
func (h *AdminHandler) GetSystemHealth(c echo.Context) error {
	report, err := h.healthService.CheckReadiness(c.Request().Context())
	if err != nil {
		c.Logger().Errorf("Failed to check system health: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to check system health",
		})
	}

	return c.JSON(http.StatusOK, report)
}

// SendTestEmail sends a test email to the requesting admin
// POST /api/v1/admin/test-email
func (h *AdminHandler) SendTestEmail(c echo.Context) error {
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
		c.Logger().Errorf("Failed to send test email to %s: %v", user.Email, err)
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "smtp_error",
			Message: "Failed to send test email. Check server logs for details.",
		})
	}

	c.Logger().Infof("Test email sent successfully to admin %s (%s) in language %s", user.Email, user.ID.String(), user.Language)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Test email sent successfully! Check your inbox.",
	})
}

// SendTestPush sends a test push notification to the requesting admin
// POST /api/v1/admin/test-push
func (h *AdminHandler) SendTestPush(c echo.Context) error {
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
		c.Logger().Errorf("Failed to send test push to %s: %v", user.Email, err)
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "push_error",
			Message: "Failed to send test push notification. Check server logs for details.",
		})
	}

	c.Logger().Infof("Test push sent successfully to admin %s (%s)", user.Email, user.ID.String())

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Test push notification sent! Check your browser notifications.",
	})
}

// SendPreviewEmail sends a preview of a specific email template to the requesting admin (dev only)
// POST /api/v1/admin/preview-email
func (h *AdminHandler) SendPreviewEmail(c echo.Context) error {
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
		err = h.emailService.SendShareNotification(ctx, user.Email, user.FirstName, "Max Muster", "voucher", "https://example.com/vouchers/sample-uuid", "https://example.com/unsubscribe?token=sample-preview-token&type=notifications", lang)
	case "transfer_notification":
		err = h.emailService.SendTransferNotification(ctx, user.Email, user.FirstName, "Max Muster", "gift_card", "https://example.com/gift-cards/sample-uuid", "https://example.com/unsubscribe?token=sample-preview-token&type=notifications", lang)
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
		slog.ErrorContext(c.Request().Context(), "failed to send preview email", "template", req.Template, "email", user.Email, "error", err)
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "smtp_error",
			Message: "Failed to send preview email. Check server logs for details.",
		})
	}

	slog.InfoContext(c.Request().Context(), "preview email sent", "template", req.Template, "email", user.Email, "language", lang)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Preview email sent successfully! Check your inbox.",
	})
}
