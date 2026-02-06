// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"net/http"
	"savvy/internal/models"
	"savvy/internal/services"
	"savvy/internal/templates"
	"savvy/internal/validation"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

// AdminHandler handles admin operations
type AdminHandler struct {
	adminService services.AdminServiceInterface
	userService  services.UserServiceInterface
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService services.AdminServiceInterface, userService services.UserServiceInterface) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		userService:  userService,
	}
}

// UsersIndex lists all users for admin
func (h *AdminHandler) UsersIndex(c echo.Context) error {
	currentUser := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	users, err := h.adminService.GetAllUsers(c.Request().Context())
	if err != nil {
		return err
	}

	return templates.AdminUsersIndex(c.Request().Context(), csrfToken, users, currentUser, isImpersonating).Render(c.Request().Context(), c.Response().Writer)
}

// AuditLogIndex displays audit log entries with filtering
func (h *AdminHandler) AuditLogIndex(c echo.Context) error {
	currentUser := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	page := 1
	perPage := 50

	filterUser := c.QueryParam("user")
	filterResourceType := c.QueryParam("resource_type")
	filterAction := c.QueryParam("action")
	filterDateFrom := c.QueryParam("date_from")
	filterDateTo := c.QueryParam("date_to")
	searchQuery := c.QueryParam("search")

	users, _ := h.adminService.GetAllUsers(c.Request().Context())

	var filterUserID *uuid.UUID
	if filterUser != "" {
		if userID, err := uuid.Parse(filterUser); err == nil {
			filterUserID = &userID
		}
	}

	filters := services.AuditLogFilters{
		UserID:       filterUserID,
		ResourceType: filterResourceType,
		Action:       filterAction,
		DateFrom:     filterDateFrom,
		DateTo:       filterDateTo,
		SearchQuery:  searchQuery,
		Page:         page,
		PerPage:      perPage,
	}

	result, err := h.adminService.GetAuditLogs(c.Request().Context(), filters)
	if err != nil {
		return err
	}

	return templates.AdminAuditLogIndex(
		c.Request().Context(),
		csrfToken,
		result.Logs,
		users,
		currentUser,
		isImpersonating,
		page,
		perPage,
		int(result.Total),
		filterUser,
		filterResourceType,
		filterAction,
		filterDateFrom,
		filterDateTo,
		searchQuery,
	).Render(c.Request().Context(), c.Response().Writer)
}

// UpdateUserRole updates a user's role (only for local auth users)
func (h *AdminHandler) UpdateUserRole(c echo.Context) error {
	currentUser := c.Get("current_user").(*models.User)
	targetUserID := c.Param("id")
	newRole := c.FormValue("role")

	userUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return c.String(400, "Invalid user ID")
	}

	if newRole != "user" && newRole != "admin" {
		return c.String(400, "Invalid role. Must be 'user' or 'admin'")
	}

	if currentUser.ID == userUUID {
		return c.String(403, "You cannot change your own role")
	}

	targetUser, err := h.userService.GetUserByID(c.Request().Context(), userUUID)
	if err != nil {
		return c.String(404, "User not found")
	}

	if targetUser.IsOAuthUser() {
		return c.String(403, "Cannot change role for OAuth users. Roles are managed by OAuth provider.")
	}

	oldRole := targetUser.Role

	if err := h.adminService.UpdateUserRole(c.Request().Context(), userUUID, newRole); err != nil {
		c.Logger().Errorf("Failed to update user role: %v", err)
		return c.String(500, "Failed to update user role")
	}

	resourceData := `{"role_change": {"old": "` + oldRole + `", "new": "` + newRole + `"}, "user_email": "` + targetUser.Email + `"}`
	auditLog := models.AuditLog{
		UserID:       &currentUser.ID,
		Action:       "update",
		ResourceType: "users",
		ResourceID:   targetUser.ID,
		ResourceData: resourceData,
		IPAddress:    c.RealIP(),
		UserAgent:    c.Request().UserAgent(),
	}
	if err := h.adminService.CreateAuditLog(c.Request().Context(), &auditLog); err != nil {
		c.Logger().Errorf("Failed to log role change: %v", err)
	}

	return c.Redirect(303, "/admin/users")
}

// RestoreResource restores a soft-deleted resource
func (h *AdminHandler) RestoreResource(c echo.Context) error {
	resourceType := c.FormValue("resource_type")
	resourceID := c.FormValue("resource_id")

	if resourceType == "" || resourceID == "" {
		return c.Redirect(303, "/admin/audit-log?error=missing_params")
	}

	id, err := uuid.Parse(resourceID)
	if err != nil {
		return c.Redirect(303, "/admin/audit-log?error=invalid_id")
	}

	if err := h.adminService.RestoreResource(c.Request().Context(), resourceType, id); err != nil {
		c.Logger().Errorf("Failed to restore resource: %v", err)
		return c.Redirect(303, "/admin/audit-log?error=restore_failed&type="+resourceType)
	}

	currentUser := c.Get("current_user").(*models.User)
	auditLog := models.AuditLog{
		UserID:       &currentUser.ID,
		Action:       "restore",
		ResourceType: resourceType,
		ResourceID:   id,
		IPAddress:    c.RealIP(),
		UserAgent:    c.Request().UserAgent(),
	}
	if err := h.adminService.CreateAuditLog(c.Request().Context(), &auditLog); err != nil {
		c.Logger().Errorf("Failed to log restore action: %v", err)
	}

	return c.Redirect(303, "/admin/audit-log?success=restored&type="+resourceType)
}

// CreateUserGet shows the user creation form (only if local login is enabled)
func (h *AdminHandler) CreateUserGet(c echo.Context) error {
	// Check if local login is enabled
	if !IsLocalLoginEnabled() {
		return echo.NewHTTPError(http.StatusNotFound, "User creation is disabled when local login is disabled")
	}

	currentUser := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	return templates.AdminCreateUser(c.Request().Context(), csrfToken, currentUser, isImpersonating, nil).Render(c.Request().Context(), c.Response().Writer)
}

// CreateUserPost handles user creation (only if local login is enabled)
func (h *AdminHandler) CreateUserPost(c echo.Context) error {
	// Check if local login is enabled
	if !IsLocalLoginEnabled() {
		return echo.NewHTTPError(http.StatusNotFound, "User creation is disabled when local login is disabled")
	}

	currentUser := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	req := validation.RegisterRequest{
		Email:     c.FormValue("email"),
		Password:  c.FormValue("password"),
		FirstName: c.FormValue("first_name"),
		LastName:  c.FormValue("last_name"),
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return templates.AdminCreateUser(c.Request().Context(), csrfToken, currentUser, isImpersonating, map[string]string{
			"error": "Bitte fülle alle Felder korrekt aus",
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	if _, err := h.userService.GetUserByEmail(c.Request().Context(), email); err == nil {
		return templates.AdminCreateUser(c.Request().Context(), csrfToken, currentUser, isImpersonating, map[string]string{
			"error": "Ein Benutzer mit dieser E-Mail-Adresse existiert bereits",
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	role := "user"
	if c.FormValue("is_admin") == "on" {
		role = "admin"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return templates.AdminCreateUser(c.Request().Context(), csrfToken, currentUser, isImpersonating, map[string]string{
			"error": "Fehler beim Hashen des Passworts",
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	user := models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         role,
		AuthProvider: "local",
	}

	if err := h.adminService.CreateLocalUser(c.Request().Context(), &user); err != nil {
		return templates.AdminCreateUser(c.Request().Context(), csrfToken, currentUser, isImpersonating, map[string]string{
			"error": "Fehler beim Erstellen des Benutzers",
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	auditLog := models.AuditLog{
		UserID:       &currentUser.ID,
		Action:       "create_user",
		ResourceType: "users",
		ResourceID:   user.ID,
		IPAddress:    c.RealIP(),
		UserAgent:    c.Request().UserAgent(),
	}
	h.adminService.CreateAuditLog(c.Request().Context(), &auditLog)

	return c.Redirect(http.StatusSeeOther, "/admin/users?success=user_created")
}
