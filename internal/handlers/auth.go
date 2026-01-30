// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"net/http"
	"net/mail"
	"savvy/internal/database"
	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/validation"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthLoginGet shows the login page
func AuthLoginGet(c echo.Context) error {
	csrfToken := c.Get("csrf").(string)
	ctx := c.Request().Context()

	// If local login is disabled and OAuth is enabled, redirect to OAuth
	if !IsLocalLoginEnabled() && IsOAuthEnabled() {
		return c.Redirect(http.StatusSeeOther, "/auth/oauth/login")
	}

	// If both local login and OAuth are disabled, return 404
	if !IsLocalLoginEnabled() && !IsOAuthEnabled() {
		return echo.NewHTTPError(http.StatusNotFound, "Login is disabled")
	}

	return templates.Login(ctx, csrfToken, IsOAuthEnabled(), IsLocalLoginEnabled()).Render(ctx, c.Response().Writer)
}

// AuthLoginPost handles login with constant-time response to prevent account enumeration
func AuthLoginPost(c echo.Context) error {
	// Check if local login is enabled
	if !IsLocalLoginEnabled() {
		return echo.NewHTTPError(http.StatusNotFound, "Local login is disabled")
	}

	// Validate input
	req := validation.LoginRequest{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	// Normalize email to lowercase
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	var user models.User
	err := database.DB.Where("LOWER(email) = ?", email).First(&user).Error

	// Always run bcrypt comparison, even if user doesn't exist
	// This prevents timing attacks that reveal whether an email exists
	var passwordHash string
	if err != nil {
		// User not found - use a dummy hash to maintain constant time
		// This is a bcrypt hash of "dummy-password-for-timing-safety"
		// #nosec G101 - This is not a real credential, it's a timing-attack mitigation dummy hash
		passwordHash = "$2a$10$rKjJZ3L.3C8qX9F5H5kqj.4nZ7Y5L5J5J5J5J5J5J5J5J5J5J5J5J"
	} else {
		passwordHash = user.PasswordHash
	}

	// Always perform bcrypt comparison (constant time operation)
	bcryptErr := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))

	// Only proceed if both user exists AND password matches
	if err == nil && bcryptErr == nil {
		// Regenerate session to prevent session fixation attacks
		// This creates a NEW session with a FRESH session ID
		newSession, err := middleware.RegenerateSession(c)
		if err != nil {
			c.Logger().Errorf("Failed to regenerate session: %v", err)
			return c.Redirect(http.StatusSeeOther, "/auth/login?error=session_error")
		}

		newSession.Values["user_id"] = user.ID.String()
		if err := newSession.Save(c.Request(), c.Response()); err != nil {
			c.Logger().Errorf("Failed to save session: %v", err)
			return c.Redirect(http.StatusSeeOther, "/auth/login?error=session_error")
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}

	// Always return the same error regardless of whether user exists or password is wrong
	return c.Redirect(http.StatusSeeOther, "/auth/login")
}

// AuthRegisterGet shows the register page
func AuthRegisterGet(c echo.Context) error {
	// Check if registration is enabled
	if !IsRegistrationEnabled() {
		// Redirect to login page instead of showing error
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	csrfToken := c.Get("csrf").(string)
	return templates.Register(c.Request().Context(), csrfToken).Render(c.Request().Context(), c.Response().Writer)
}

// AuthRegisterPost handles registration
func AuthRegisterPost(c echo.Context) error {
	// Check if registration is enabled
	if !IsRegistrationEnabled() {
		// Redirect to login page instead of showing error
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	// Validate input
	req := validation.RegisterRequest{
		Email:     c.FormValue("email"),
		Password:  c.FormValue("password"),
		FirstName: c.FormValue("first_name"),
		LastName:  c.FormValue("last_name"),
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return c.Redirect(http.StatusSeeOther, "/auth/register?error=validation")
	}

	// Normalize email to lowercase
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Validate email format
	_, err := mail.ParseAddress(email)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/auth/register?error=invalid_email")
	}

	// Check if email already exists (case-insensitive)
	var existingUser models.User
	if database.DB.Where("LOWER(email) = ?", email).First(&existingUser).Error == nil {
		return c.Redirect(http.StatusSeeOther, "/auth/register?error=email_exists")
	}

	// Validate password strength
	if err := validation.ValidatePassword(req.Password); err != nil {
		return c.Redirect(http.StatusSeeOther, "/auth/register?error=weak_password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/auth/register")
	}

	user := models.User{
		Email:        email, // Store normalized lowercase email
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         "user",
		AuthProvider: "local",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/auth/register")
	}

	// Regenerate session to prevent session fixation attacks
	// This creates a NEW session with a FRESH session ID
	newSession, err := middleware.RegenerateSession(c)
	if err != nil {
		c.Logger().Errorf("Failed to regenerate session: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/register?error=session_error")
	}

	newSession.Values["user_id"] = user.ID.String()
	if err := newSession.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Errorf("Failed to save session: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/register?error=session_error")
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// AuthLogout logs out the user
func AuthLogout(c echo.Context) error {
	session, _ := middleware.GetSession(c)
	session.Options.MaxAge = -1
	if err := session.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Errorf("Failed to save session during logout: %v", err)
	}
	return c.Redirect(http.StatusSeeOther, "/auth/login")
}

// AdminImpersonate allows admin to impersonate another user
func AdminImpersonate(c echo.Context) error {
	currentUser := c.Get("current_user").(*models.User)
	if !currentUser.IsAdmin() {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	targetUserID := c.Param("id")
	userUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/admin/users")
	}

	var targetUser models.User
	if err := database.DB.First(&targetUser, userUUID).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/admin/users")
	}

	// No session regeneration needed for impersonation
	// (session fixation is not a risk here since admin is already authenticated)
	session, _ := middleware.GetSession(c)

	c.Logger().Printf("Admin %s (%s) impersonating user %s",
		currentUser.DisplayName(), currentUser.ID, targetUser.DisplayName())

	// Store original user_id and switch to target user
	session.Values["impersonated_by"] = currentUser.ID.String()
	session.Values["user_id"] = targetUser.ID.String()
	// Note: oauth_login flag is automatically preserved (no session regeneration)

	if err := session.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Errorf("Failed to save session during impersonation: %v", err)
		return c.Redirect(http.StatusSeeOther, "/admin/users")
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// AdminStopImpersonate stops impersonating
func AdminStopImpersonate(c echo.Context) error {
	session, _ := middleware.GetSession(c)

	impersonatedBy, ok := session.Values["impersonated_by"].(string)
	if !ok || impersonatedBy == "" {
		c.Logger().Warn("Stop impersonate called but no impersonation active")
		return c.Redirect(http.StatusSeeOther, "/")
	}

	c.Logger().Printf("Stopping impersonation, returning to admin user: %s", impersonatedBy)

	session.Values["user_id"] = impersonatedBy
	delete(session.Values, "impersonated_by")

	// Log session state for debugging
	c.Logger().Printf("Session after stop impersonate - user_id: %s, oauth_login: %v",
		session.Values["user_id"], session.Values["oauth_login"])

	if err := session.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Errorf("Failed to save session when stopping impersonation: %v", err)
	}

	return c.Redirect(http.StatusSeeOther, "/admin/users")
}
