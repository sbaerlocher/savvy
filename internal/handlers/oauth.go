// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/oauth"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	roleUser  = "user"
	roleAdmin = "admin"
)

var oauthProvider *oauth.Provider
var oauthConfig *config.Config
var oauthEnabled bool

// InitOAuth initializes the OAuth provider
func InitOAuth(provider *oauth.Provider, cfg *config.Config) {
	oauthProvider = provider
	oauthConfig = cfg
	oauthEnabled = provider != nil
}

// IsOAuthEnabled returns true if OAuth is configured and enabled
func IsOAuthEnabled() bool {
	return oauthEnabled
}

// IsLocalLoginEnabled returns true if local login (email/password) is enabled
func IsLocalLoginEnabled() bool {
	if oauthConfig == nil {
		return true // Default to enabled if no config
	}
	return oauthConfig.EnableLocalLogin
}

// IsRegistrationEnabled returns true if user registration is enabled
func IsRegistrationEnabled() bool {
	if oauthConfig == nil {
		return true // Default to enabled if no config
	}
	return oauthConfig.EnableRegistration
}

// shouldBeAdmin checks if a user should have admin privileges based on email or group membership
func shouldBeAdmin(email string, groups []string) bool {
	if oauthConfig == nil {
		return false
	}

	// Check if email is in admin emails list
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	for _, adminEmail := range oauthConfig.OAuthAdminEmails {
		if strings.ToLower(strings.TrimSpace(adminEmail)) == normalizedEmail {
			return true
		}
	}

	// Check if user is in admin group
	if oauthConfig.OAuthAdminGroup != "" {
		// Normalize both the configured group and the user's groups for comparison
		normalizedConfigGroup := strings.ToLower(strings.TrimSpace(oauthConfig.OAuthAdminGroup))

		for _, group := range groups {
			normalizedGroup := strings.ToLower(strings.TrimSpace(group))

			// Try exact match first
			if normalizedGroup == normalizedConfigGroup {
				return true
			}

			// Try with underscores replaced by spaces
			if strings.ReplaceAll(normalizedGroup, "_", " ") == normalizedConfigGroup {
				return true
			}

			// Try with spaces replaced by underscores
			if strings.ReplaceAll(normalizedGroup, " ", "_") == normalizedConfigGroup {
				return true
			}
		}
	}

	return false
}

// OAuthLogin redirects to Authentik login
func OAuthLogin(c echo.Context) error {
	if oauthProvider == nil {
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=oauth_not_configured")
	}

	// Generate state token for CSRF protection
	state, err := generateRandomString(32)
	if err != nil {
		c.Logger().Errorf("Failed to generate state: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=state_generation_failed")
	}

	// Store state in session
	sess, _ := middleware.GetSession(c)
	sess.Values["oauth_state"] = state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save session")
	}

	// Redirect to Authentik
	url := oauthProvider.Config.AuthCodeURL(state)
	return c.Redirect(http.StatusSeeOther, url)
}

// OAuthCallback handles the OAuth callback from Authentik
func OAuthCallback(c echo.Context) error {
	if oauthProvider == nil {
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=oauth_not_configured")
	}

	// Verify state
	sess, _ := middleware.GetSession(c)
	savedState, ok := sess.Values["oauth_state"].(string)
	if !ok || savedState == "" {
		c.Logger().Warn("OAuth state not found in session")
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=invalid_state")
	}

	state := c.QueryParam("state")
	if state != savedState {
		c.Logger().Warnf("OAuth state mismatch: got %s, expected %s", state, savedState)
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=state_mismatch")
	}

	// Clear state from session
	delete(sess.Values, "oauth_state")

	// Exchange code for token
	code := c.QueryParam("code")
	if code == "" {
		c.Logger().Warn("OAuth callback missing code")
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=missing_code")
	}

	ctx := context.Background()
	oauth2Token, err := oauthProvider.Config.Exchange(ctx, code)
	if err != nil {
		c.Logger().Errorf("Failed to exchange code: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=token_exchange_failed")
	}

	// Extract ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		c.Logger().Error("No id_token in OAuth response")
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=no_id_token")
	}

	// Verify ID token
	idToken, err := oauthProvider.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		c.Logger().Errorf("Failed to verify ID token: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=token_verification_failed")
	}

	// Extract user info
	var userInfo oauth.UserInfo
	if err := idToken.Claims(&userInfo); err != nil {
		c.Logger().Errorf("Failed to parse claims: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=claims_parsing_failed")
	}

	// Debug logging
	c.Logger().Printf("OAuth claims - Email: %s, FirstName: '%s', LastName: '%s', PreferredUsername: %s, Groups: %v",
		userInfo.Email, userInfo.FirstName, userInfo.LastName, userInfo.PreferredUsername, userInfo.Groups)

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(userInfo.Email))
	if email == "" {
		c.Logger().Error("Email not found in OAuth claims")
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=no_email")
	}

	// Handle name splitting if LastName is empty but FirstName contains full name
	firstName := userInfo.FirstName
	lastName := userInfo.LastName
	if lastName == "" && firstName != "" && strings.Contains(firstName, " ") {
		// Split full name into first and last name
		nameParts := strings.SplitN(firstName, " ", 2)
		firstName = nameParts[0]
		lastName = nameParts[1]
		c.Logger().Printf("Split name: FirstName='%s', LastName='%s'", firstName, lastName)
	}

	// Find or create user
	var user models.User
	err = database.DB.Where("LOWER(email) = ?", email).First(&user).Error
	if err != nil {
		// User doesn't exist, create new user
		c.Logger().Printf("Creating new user from OAuth: %s", email)

		// Generate random password (won't be used for OAuth login)
		randomPassword, err := generateRandomString(32)
		if err != nil {
			c.Logger().Errorf("Failed to generate random password: %v", err)
			return c.Redirect(http.StatusSeeOther, "/auth/login?error=user_creation_failed")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
		if err != nil {
			c.Logger().Errorf("Failed to hash password: %v", err)
			return c.Redirect(http.StatusSeeOther, "/auth/login?error=user_creation_failed")
		}

		// Check if user should be admin
		isAdmin := shouldBeAdmin(email, userInfo.Groups)
		role := roleUser
		if isAdmin {
			role = roleAdmin
		}

		user = models.User{
			Email:        email,
			PasswordHash: string(hashedPassword),
			FirstName:    firstName,
			LastName:     lastName,
			Role:         role,
			AuthProvider: "oauth",
		}

		if err := database.DB.Create(&user).Error; err != nil {
			c.Logger().Errorf("Failed to create user: %v", err)
			return c.Redirect(http.StatusSeeOther, "/auth/login?error=user_creation_failed")
		}

		if isAdmin {
			c.Logger().Printf("Created admin user from OAuth: %s (groups: %v)", email, userInfo.Groups)
		}
	} else {
		// User exists, optionally update profile
		c.Logger().Printf("OAuth login for existing user: %s", email)

		// Check if user should be admin (permissions may have changed)
		isAdmin := shouldBeAdmin(email, userInfo.Groups)
		expectedRole := roleUser
		if isAdmin {
			expectedRole = roleAdmin
		}

		// Update name if changed
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
			c.Logger().Printf("Updated admin status for user %s: %v (groups: %v)", email, isAdmin, userInfo.Groups)
		}
		// Ensure AuthProvider is set to oauth for existing users
		if user.AuthProvider != "oauth" {
			user.AuthProvider = "oauth"
			updated = true
		}

		if updated {
			database.DB.Save(&user)
		}
	}

	// Regenerate session to prevent session fixation attacks
	// This creates a NEW session with a FRESH session ID
	// Note: We already used 'sess' for OAuth state verification, now regenerate it
	newSess, err := middleware.RegenerateSession(c)
	if err != nil {
		c.Logger().Errorf("Failed to regenerate session: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=session_error")
	}

	newSess.Values["user_id"] = user.ID.String()
	newSess.Values["oauth_login"] = true
	if err := newSess.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Errorf("Failed to save session: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=session_save_failed")
	}

	c.Logger().Printf("OAuth login successful for user: %s", email)
	return c.Redirect(http.StatusSeeOther, "/")
}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
