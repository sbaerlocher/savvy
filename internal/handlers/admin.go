// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/validation"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// AdminUsersIndex lists all users for admin
func AdminUsersIndex(c echo.Context) error {
	currentUser := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	var users []models.User
	if err := database.DB.Order("created_at DESC").Find(&users).Error; err != nil {
		return err
	}

	return templates.AdminUsersIndex(c.Request().Context(), csrfToken, users, currentUser, isImpersonating).Render(c.Request().Context(), c.Response().Writer)
}

// AdminAuditLogIndex displays audit log entries with filtering
func AdminAuditLogIndex(c echo.Context) error {
	currentUser := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	// Pagination
	page := 1
	perPage := 50

	// Filter parameters
	filterUser := c.QueryParam("user")
	filterResourceType := c.QueryParam("resource_type")
	filterAction := c.QueryParam("action")
	filterDateFrom := c.QueryParam("date_from")
	filterDateTo := c.QueryParam("date_to")
	searchQuery := c.QueryParam("search")

	// Get all users for filter dropdown
	var users []models.User
	database.DB.Order("first_name, last_name").Find(&users)

	// Build query with filters
	query := database.DB.Model(&models.AuditLog{}).
		Preload("User").
		Order("created_at DESC")

	// Apply filters
	if filterUser != "" {
		if userID, err := uuid.Parse(filterUser); err == nil {
			query = query.Where("user_id = ?", userID)
		}
	}

	if filterResourceType != "" {
		query = query.Where("resource_type = ?", filterResourceType)
	}

	if filterAction != "" {
		query = query.Where("action = ?", filterAction)
	}

	if filterDateFrom != "" {
		query = query.Where("created_at >= ?", filterDateFrom)
	}

	if filterDateTo != "" {
		query = query.Where("created_at <= ?", filterDateTo+" 23:59:59")
	}

	if searchQuery != "" {
		// Search in resource_data JSON field
		query = query.Where("resource_data::text ILIKE ?", "%"+searchQuery+"%")
	}

	// Count total with filters
	var total int64
	query.Count(&total)

	// Get paginated results
	var auditLogs []models.AuditLog
	offset := (page - 1) * perPage
	if err := query.Limit(perPage).Offset(offset).Find(&auditLogs).Error; err != nil {
		return err
	}

	return templates.AdminAuditLogIndex(
		c.Request().Context(),
		csrfToken,
		auditLogs,
		users,
		currentUser,
		isImpersonating,
		page,
		perPage,
		int(total),
		filterUser,
		filterResourceType,
		filterAction,
		filterDateFrom,
		filterDateTo,
		searchQuery,
	).Render(c.Request().Context(), c.Response().Writer)
}

// AdminUpdateUserRole updates a user's role (only for local auth users)
func AdminUpdateUserRole(c echo.Context) error {
	currentUser := c.Get("current_user").(*models.User)
	targetUserID := c.Param("id")
	newRole := c.FormValue("role")

	// Parse target user ID
	userUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return c.String(400, "Invalid user ID")
	}

	// Validate new role
	if newRole != "user" && newRole != "admin" {
		return c.String(400, "Invalid role. Must be 'user' or 'admin'")
	}

	// Prevent users from changing their own role
	if currentUser.ID == userUUID {
		return c.String(403, "You cannot change your own role")
	}

	// Get target user
	var targetUser models.User
	if err := database.DB.First(&targetUser, userUUID).Error; err != nil {
		return c.String(404, "User not found")
	}

	// Prevent changing OAuth users' roles
	if targetUser.IsOAuthUser() {
		return c.String(403, "Cannot change role for OAuth users. Roles are managed by OAuth provider.")
	}

	// Store old role for audit log
	oldRole := targetUser.Role

	// Update role
	targetUser.Role = newRole
	if err := database.DB.Save(&targetUser).Error; err != nil {
		c.Logger().Errorf("Failed to update user role: %v", err)
		return c.String(500, "Failed to update user role")
	}

	// Create audit log entry
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
	if err := database.DB.Create(&auditLog).Error; err != nil {
		c.Logger().Errorf("Failed to log role change: %v", err)
	}

	return c.Redirect(303, "/admin/users")
}

// AdminRestoreResource restores a soft-deleted resource
func AdminRestoreResource(c echo.Context) error {
	resourceType := c.FormValue("resource_type")
	resourceID := c.FormValue("resource_id")

	if resourceType == "" || resourceID == "" {
		return c.Redirect(303, "/admin/audit-log?error=missing_params")
	}

	// Parse UUID
	id, err := uuid.Parse(resourceID)
	if err != nil {
		return c.Redirect(303, "/admin/audit-log?error=invalid_id")
	}

	// Restore based on resource type
	var tableName string

	switch resourceType {
	case "cards":
		var card models.Card
		if err := database.DB.Unscoped().Where("id = ?", id).First(&card).Error; err != nil {
			return c.Redirect(303, "/admin/audit-log?error=not_found&type=card")
		}
		if !card.DeletedAt.Valid {
			return c.Redirect(303, "/admin/audit-log?error=not_deleted&type=card")
		}
		database.DB.Unscoped().Model(&card).Update("deleted_at", nil)
		tableName = "cards"

	case "card_shares":
		var share models.CardShare
		if err := database.DB.Unscoped().Where("id = ?", id).First(&share).Error; err != nil {
			return c.Redirect(303, "/admin/audit-log?error=not_found&type=card_share")
		}
		if !share.DeletedAt.Valid {
			return c.Redirect(303, "/admin/audit-log?error=not_deleted&type=card_share")
		}
		database.DB.Unscoped().Model(&share).Update("deleted_at", nil)
		tableName = "card_shares"

	case "vouchers":
		var voucher models.Voucher
		if err := database.DB.Unscoped().Where("id = ?", id).First(&voucher).Error; err != nil {
			return c.Redirect(303, "/admin/audit-log?error=not_found&type=voucher")
		}
		if !voucher.DeletedAt.Valid {
			return c.Redirect(303, "/admin/audit-log?error=not_deleted&type=voucher")
		}
		database.DB.Unscoped().Model(&voucher).Update("deleted_at", nil)
		tableName = "vouchers"

	case "voucher_shares":
		var share models.VoucherShare
		if err := database.DB.Unscoped().Where("id = ?", id).First(&share).Error; err != nil {
			return c.Redirect(303, "/admin/audit-log?error=not_found&type=voucher_share")
		}
		if !share.DeletedAt.Valid {
			return c.Redirect(303, "/admin/audit-log?error=not_deleted&type=voucher_share")
		}
		database.DB.Unscoped().Model(&share).Update("deleted_at", nil)
		tableName = "voucher_shares"

	case "gift_cards":
		var giftCard models.GiftCard
		if err := database.DB.Unscoped().Where("id = ?", id).First(&giftCard).Error; err != nil {
			return c.Redirect(303, "/admin/audit-log?error=not_found&type=gift_card")
		}
		if !giftCard.DeletedAt.Valid {
			return c.Redirect(303, "/admin/audit-log?error=not_deleted&type=gift_card")
		}
		database.DB.Unscoped().Model(&giftCard).Update("deleted_at", nil)
		tableName = "gift_cards"

	case "gift_card_shares":
		var share models.GiftCardShare
		if err := database.DB.Unscoped().Where("id = ?", id).First(&share).Error; err != nil {
			return c.Redirect(303, "/admin/audit-log?error=not_found&type=gift_card_share")
		}
		if !share.DeletedAt.Valid {
			return c.Redirect(303, "/admin/audit-log?error=not_deleted&type=gift_card_share")
		}
		database.DB.Unscoped().Model(&share).Update("deleted_at", nil)
		tableName = "gift_card_shares"

	case "gift_card_transactions":
		var transaction models.GiftCardTransaction
		if err := database.DB.Unscoped().Where("id = ?", id).First(&transaction).Error; err != nil {
			return c.Redirect(303, "/admin/audit-log?error=not_found&type=transaction")
		}
		if !transaction.DeletedAt.Valid {
			return c.Redirect(303, "/admin/audit-log?error=not_deleted&type=transaction")
		}
		database.DB.Unscoped().Model(&transaction).Update("deleted_at", nil)
		tableName = "gift_card_transactions"

	case "merchants":
		var merchant models.Merchant
		if err := database.DB.Unscoped().Where("id = ?", id).First(&merchant).Error; err != nil {
			return c.Redirect(303, "/admin/audit-log?error=not_found&type=merchant")
		}
		if !merchant.DeletedAt.Valid {
			return c.Redirect(303, "/admin/audit-log?error=not_deleted&type=merchant")
		}
		database.DB.Unscoped().Model(&merchant).Update("deleted_at", nil)
		tableName = "merchants"

	default:
		return c.Redirect(303, "/admin/audit-log?error=unsupported_type")
	}

	// Create new audit log entry for restore action
	auditLog := models.AuditLog{
		UserID:       &c.Get("current_user").(*models.User).ID,
		Action:       "restore",
		ResourceType: tableName,
		ResourceID:   id,
		IPAddress:    c.RealIP(),
		UserAgent:    c.Request().UserAgent(),
	}
	if err := database.DB.Create(&auditLog).Error; err != nil {
		c.Logger().Errorf("Failed to log restore action: %v", err)
	}

	return c.Redirect(303, "/admin/audit-log?success=restored&type="+resourceType)
}

// AdminCreateUserGet shows the user creation form (only if local login is enabled)
func AdminCreateUserGet(c echo.Context) error {
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

// AdminCreateUserPost handles user creation (only if local login is enabled)
func AdminCreateUserPost(c echo.Context) error {
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

	// Validate input
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

	// Normalize email to lowercase
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("LOWER(email) = ?", email).First(&existingUser).Error; err == nil {
		return templates.AdminCreateUser(c.Request().Context(), csrfToken, currentUser, isImpersonating, map[string]string{
			"error": "Ein Benutzer mit dieser E-Mail-Adresse existiert bereits",
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	// Check if admin checkbox is set
	role := "user"
	if c.FormValue("is_admin") == "on" {
		role = "admin"
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return templates.AdminCreateUser(c.Request().Context(), csrfToken, currentUser, isImpersonating, map[string]string{
			"error": "Fehler beim Hashen des Passworts",
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	// Create user
	user := models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         role,
		AuthProvider: "local",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return templates.AdminCreateUser(c.Request().Context(), csrfToken, currentUser, isImpersonating, map[string]string{
			"error": "Fehler beim Erstellen des Benutzers",
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	// Create audit log entry
	auditLog := models.AuditLog{
		UserID:       &currentUser.ID,
		Action:       "create_user",
		ResourceType: "users",
		ResourceID:   user.ID,
		IPAddress:    c.RealIP(),
		UserAgent:    c.Request().UserAgent(),
	}
	database.DB.Create(&auditLog)

	return c.Redirect(http.StatusSeeOther, "/admin/users?success=user_created")
}
