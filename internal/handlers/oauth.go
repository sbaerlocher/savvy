// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"savvy/internal/config"
	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/oauth"
	"savvy/internal/services"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

// errRedirected is returned by helper functions after sending an HTTP redirect.
// Callers must check for this error to avoid continuing with nil values.
var errRedirected = errors.New("already redirected")

// OAuthHandler handles OAuth authentication operations.
// All OAuth state is held in struct fields (no package-level globals).
type OAuthHandler struct {
	userService services.UserServiceInterface
	provider    *oauth.Provider
	cfg         *config.Config
}

// NewOAuthHandler creates a new OAuth handler with injected dependencies.
func NewOAuthHandler(userService services.UserServiceInterface, provider *oauth.Provider, cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		userService: userService,
		provider:    provider,
		cfg:         cfg,
	}
}

const (
	roleUser  = "user"
	roleAdmin = "admin"
)

// Login redirects to OAuth provider login page.
// GET /auth/oauth/login
func (h *OAuthHandler) Login(c echo.Context) error {
	if h.provider == nil {
		return c.Redirect(http.StatusSeeOther, "/login?error=oauth_not_configured")
	}

	state, err := generateRandomString(32)
	if err != nil {
		slog.Error("Failed to generate state", "error", err)
		return c.Redirect(http.StatusSeeOther, "/login?error=state_generation_failed")
	}

	sess, err := middleware.GetSession(c)
	if err != nil {
		slog.Error("Failed to get session for OAuth login", "error", err)
		return c.Redirect(http.StatusSeeOther, "/login?error=session_error")
	}
	// Clear previous session data to prevent stale OAuth state from previous attempts
	// while preserving internal PGStore metadata needed for session persistence
	middleware.ClearSessionUserValues(sess)
	sess.Values[middleware.SessionKeyOAuthState] = state
	if err := middleware.SaveSession(c, sess); err != nil {
		slog.Error("Failed to save OAuth session state", "error", err)
		return c.Redirect(http.StatusSeeOther, "/login?error=session_error")
	}

	url := h.provider.Config.AuthCodeURL(state)
	return c.Redirect(http.StatusSeeOther, url)
}

// Callback handles the OAuth callback from the provider.
// GET /auth/oauth/callback
func (h *OAuthHandler) Callback(c echo.Context) error {
	if h.provider == nil {
		return c.Redirect(http.StatusSeeOther, "/login?error=oauth_not_configured")
	}

	// Validate OAuth state
	code, err := h.validateOAuthState(c)
	if err != nil {
		return err // Already redirected in helper
	}

	// Exchange code for token and get user info
	userInfo, err := h.exchangeCodeForToken(c, code)
	if err != nil {
		return err // Already redirected in helper
	}

	// Extract and validate user info
	email, firstName, lastName, err := h.extractUserInfo(c, userInfo)
	if err != nil {
		return err // Already redirected in helper
	}

	// Create or update user
	user, err := h.createOrUpdateUser(c, email, firstName, lastName, userInfo.Groups)
	if err != nil {
		return err // Already redirected in helper
	}

	// Save user session
	if err := h.saveUserSession(c, user); err != nil {
		return err // Already redirected in helper
	}

	slog.Info("OAuth login successful", "email", maskEmail(email))

	// Redirect to frontend
	return h.redirectToFrontend(c)
}

// validateOAuthState validates the OAuth state parameter and returns the authorization code
func (h *OAuthHandler) validateOAuthState(c echo.Context) (string, error) {
	sess, _ := middleware.GetSession(c)
	savedState := middleware.GetSessionOAuthState(sess)
	if savedState == "" {
		slog.Warn("OAuth authentication failed: state not found in session")
		_ = c.Redirect(http.StatusSeeOther, "/login?error=authentication_failed")
		return "", errRedirected
	}

	state := c.QueryParam("state")
	if state != savedState {
		slog.Warn("OAuth authentication failed: state mismatch")
		_ = c.Redirect(http.StatusSeeOther, "/login?error=authentication_failed")
		return "", errRedirected
	}

	delete(sess.Values, middleware.SessionKeyOAuthState)
	// Persist state deletion immediately to prevent replay attacks
	if err := middleware.SaveSession(c, sess); err != nil {
		slog.Warn("Failed to save session after state deletion", "error", err)
	}

	code := c.QueryParam("code")
	if code == "" {
		slog.Warn("OAuth callback missing code")
		_ = c.Redirect(http.StatusSeeOther, "/login?error=missing_code")
		return "", errRedirected
	}

	return code, nil
}

// exchangeCodeForToken exchanges the authorization code for tokens and returns user info
func (h *OAuthHandler) exchangeCodeForToken(c echo.Context, code string) (*oauth.UserInfo, error) {
	// Use request context for proper timeout and trace propagation
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	oauth2Token, err := h.provider.Config.Exchange(ctx, code)
	if err != nil {
		slog.Error("Failed to exchange code", "error", err)
		_ = c.Redirect(http.StatusSeeOther, "/login?error=token_exchange_failed")
		return nil, errRedirected
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		slog.Error("OAuth authentication failed: no id_token in response")
		_ = c.Redirect(http.StatusSeeOther, "/login?error=authentication_failed")
		return nil, errRedirected
	}

	idToken, err := h.provider.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		slog.Error("OAuth authentication failed: ID token verification failed", "error", err)
		_ = c.Redirect(http.StatusSeeOther, "/login?error=authentication_failed")
		return nil, errRedirected
	}

	var userInfo oauth.UserInfo
	if err := idToken.Claims(&userInfo); err != nil {
		slog.Error("OAuth authentication failed: claims parsing failed", "error", err)
		_ = c.Redirect(http.StatusSeeOther, "/login?error=authentication_failed")
		return nil, errRedirected
	}

	slog.Info("OAuth claims received", "email", maskEmail(userInfo.Email), "groups_count", len(userInfo.Groups))

	return &userInfo, nil
}

// maskEmail masks an email address for logging (e.g., "te***@example.com")
func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return "***"
	}
	local := parts[0]
	if len(local) <= 2 {
		return local + "***@" + parts[1]
	}
	return local[:2] + "***@" + parts[1]
}

// extractUserInfo extracts and validates user information from OAuth claims
func (h *OAuthHandler) extractUserInfo(c echo.Context, userInfo *oauth.UserInfo) (email, firstName, lastName string, err error) {
	email = strings.ToLower(strings.TrimSpace(userInfo.Email))
	if email == "" {
		slog.Error("Email not found in OAuth claims")
		_ = c.Redirect(http.StatusSeeOther, "/login?error=no_email")
		return "", "", "", errRedirected
	}

	firstName = userInfo.FirstName
	lastName = userInfo.LastName
	if lastName == "" && firstName != "" && strings.Contains(firstName, " ") {
		nameParts := strings.SplitN(firstName, " ", 2)
		firstName = nameParts[0]
		lastName = nameParts[1]
		slog.Debug("Split name from OAuth claims", "first_name", firstName, "last_name", lastName)
	}

	return email, firstName, lastName, nil
}

// shouldBeAdmin checks if a user should have admin privileges based on email or group membership.
func (h *OAuthHandler) shouldBeAdmin(email string, groups []string) bool {
	if h.cfg == nil {
		return false
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	for _, adminEmail := range h.cfg.OAuthAdminEmails {
		if strings.ToLower(strings.TrimSpace(adminEmail)) == normalizedEmail {
			return true
		}
	}

	if h.cfg.OAuthAdminGroup != "" {
		normalizedConfigGroup := strings.ToLower(strings.TrimSpace(h.cfg.OAuthAdminGroup))
		slog.Debug("OAuth admin group check",
			"config_group", normalizedConfigGroup,
			"user_groups_count", len(groups))

		for i, group := range groups {
			normalizedGroup := strings.ToLower(strings.TrimSpace(group))
			slog.Debug("OAuth admin group matching",
				"group_index", i,
				"group", normalizedGroup,
				"config_group", normalizedConfigGroup)

			// Exact case-insensitive match only (no automatic normalization)
			if normalizedGroup == normalizedConfigGroup {
				slog.Info("OAuth admin access granted via group match",
					"email", email,
					"group", normalizedGroup)
				return true
			}
		}
		slog.Debug("OAuth admin group check completed - no match found",
			"email", email,
			"groups_checked", len(groups))
	}

	return false
}

// createOrUpdateUser creates a new user or updates an existing one
func (h *OAuthHandler) createOrUpdateUser(c echo.Context, email, firstName, lastName string, groups []string) (*models.User, error) {
	user, err := h.userService.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		// Create new user
		return h.createNewOAuthUser(c, email, firstName, lastName, groups)
	}

	// Update existing user
	return h.updateExistingOAuthUser(c, user, firstName, lastName, groups)
}

// createNewOAuthUser creates a new user from OAuth login
func (h *OAuthHandler) createNewOAuthUser(c echo.Context, email, firstName, lastName string, groups []string) (*models.User, error) {
	slog.Info("Creating new user from OAuth", "email", maskEmail(email))

	randomPassword, err := generateRandomString(32)
	if err != nil {
		slog.Error("Failed to generate random password", "error", err)
		_ = c.Redirect(http.StatusSeeOther, "/login?error=user_creation_failed")
		return nil, errRedirected
	}

	// Hash password with cost 12 for enhanced security
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword), 12)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		_ = c.Redirect(http.StatusSeeOther, "/login?error=user_creation_failed")
		return nil, errRedirected
	}

	isAdmin := h.shouldBeAdmin(email, groups)
	role := roleUser
	if isAdmin {
		role = roleAdmin
	}

	newUser := models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		AuthProvider: "oauth",
	}

	if err := h.userService.CreateUser(c.Request().Context(), &newUser); err != nil {
		slog.Error("Failed to create user", "error", err)
		_ = c.Redirect(http.StatusSeeOther, "/login?error=user_creation_failed")
		return nil, errRedirected
	}

	if isAdmin {
		slog.Info("Created admin user from OAuth", "email", maskEmail(email), "groups_count", len(groups))
	}

	return &newUser, nil
}

// updateExistingOAuthUser updates an existing user with OAuth data
func (h *OAuthHandler) updateExistingOAuthUser(c echo.Context, user *models.User, firstName, lastName string, groups []string) (*models.User, error) {
	slog.Info("OAuth login for existing user", "email", maskEmail(user.Email))
	slog.Debug("Checking admin status", "groups_count", len(groups))

	isAdmin := h.shouldBeAdmin(user.Email, groups)
	slog.Debug("Admin check result", "is_admin", isAdmin)

	expectedRole := roleUser
	if isAdmin {
		expectedRole = roleAdmin
		slog.Debug("Setting expected role", "role", "admin")
	} else {
		slog.Debug("Setting expected role", "role", "user", "current_role", user.Role)
	}

	updated := false
	if firstName != "" && user.FirstName != firstName {
		user.FirstName = firstName
		updated = true
	}
	if lastName != "" && user.LastName != lastName {
		user.LastName = lastName
		updated = true
	}
	if user.Role != expectedRole {
		user.Role = expectedRole
		updated = true
		slog.Info("Updated admin status", "email", maskEmail(user.Email), "is_admin", isAdmin, "groups_count", len(groups))
	}
	if user.AuthProvider != "oauth" {
		user.AuthProvider = "oauth"
		updated = true
	}

	if updated {
		if err := h.userService.UpdateUser(c.Request().Context(), user); err != nil {
			slog.Error("Failed to update user profile", "error", err)
		}
	}

	return user, nil
}

// saveUserSession regenerates the session and saves user information
func (h *OAuthHandler) saveUserSession(c echo.Context, user *models.User) error {
	// Create authenticated session (regenerates to prevent session fixation)
	newSess, err := middleware.CreateUserSession(c, user.ID.String())
	if err != nil {
		slog.Error("Failed to create user session", "error", err)
		_ = c.Redirect(http.StatusSeeOther, "/login?error=session_error")
		return errRedirected
	}

	// Mark this session as OAuth-originated
	newSess.Values[middleware.SessionKeyOAuthLogin] = true
	if err := middleware.SaveSession(c, newSess); err != nil {
		slog.Error("Failed to save OAuth session flag", "error", err)
		// Session already created by CreateUserSession, this is best-effort
	}

	return nil
}

// redirectToFrontend redirects to the frontend URL after successful OAuth login
func (h *OAuthHandler) redirectToFrontend(c echo.Context) error {
	redirectURL := "/"
	if h.cfg != nil && h.cfg.FrontendURL != "" {
		redirectURL = h.cfg.FrontendURL + "/"
	}
	return c.Redirect(http.StatusSeeOther, redirectURL)
}

// generateRandomString generates a cryptographically secure random string of the specified length.
// Uses hex encoding (2 chars per byte) for simplicity and full entropy preservation.
func generateRandomString(length int) (string, error) {
	// Hex encoding produces 2 characters per byte
	numBytes := (length + 1) / 2

	bytes := make([]byte, numBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to hex and truncate to exact length
	encoded := hex.EncodeToString(bytes)
	if len(encoded) > length {
		encoded = encoded[:length]
	}

	return encoded, nil
}
