// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"savvy/internal/config"
	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/oauth"
	"savvy/internal/services"
	"strings"

	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

// OAuthHandler handles OAuth authentication operations.
type OAuthHandler struct {
	userService services.UserServiceInterface
}

// NewOAuthHandler creates a new OAuth handler.
func NewOAuthHandler(userService services.UserServiceInterface) *OAuthHandler {
	return &OAuthHandler{userService: userService}
}

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

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	for _, adminEmail := range oauthConfig.OAuthAdminEmails {
		if strings.ToLower(strings.TrimSpace(adminEmail)) == normalizedEmail {
			return true
		}
	}

	if oauthConfig.OAuthAdminGroup != "" {
		normalizedConfigGroup := strings.ToLower(strings.TrimSpace(oauthConfig.OAuthAdminGroup))

		for _, group := range groups {
			normalizedGroup := strings.ToLower(strings.TrimSpace(group))

			if normalizedGroup == normalizedConfigGroup {
				return true
			}

			if strings.ReplaceAll(normalizedGroup, "_", " ") == normalizedConfigGroup {
				return true
			}

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

	state, err := generateRandomString(32)
	if err != nil {
		c.Logger().Errorf("Failed to generate state: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=state_generation_failed")
	}

	sess, _ := middleware.GetSession(c)
	sess.Values["oauth_state"] = state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save session")
	}

	url := oauthProvider.Config.AuthCodeURL(state)
	return c.Redirect(http.StatusSeeOther, url)
}

// Callback handles the OAuth callback from Authentik
func (h *OAuthHandler) Callback(c echo.Context) error {
	if oauthProvider == nil {
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=oauth_not_configured")
	}

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

	delete(sess.Values, "oauth_state")

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

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		c.Logger().Error("No id_token in OAuth response")
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=no_id_token")
	}

	idToken, err := oauthProvider.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		c.Logger().Errorf("Failed to verify ID token: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=token_verification_failed")
	}

	var userInfo oauth.UserInfo
	if err := idToken.Claims(&userInfo); err != nil {
		c.Logger().Errorf("Failed to parse claims: %v", err)
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=claims_parsing_failed")
	}

	c.Logger().Printf("OAuth claims - Email: %s, FirstName: '%s', LastName: '%s', PreferredUsername: %s, Groups: %v",
		userInfo.Email, userInfo.FirstName, userInfo.LastName, userInfo.PreferredUsername, userInfo.Groups)

	email := strings.ToLower(strings.TrimSpace(userInfo.Email))
	if email == "" {
		c.Logger().Error("Email not found in OAuth claims")
		return c.Redirect(http.StatusSeeOther, "/auth/login?error=no_email")
	}

	firstName := userInfo.FirstName
	lastName := userInfo.LastName
	if lastName == "" && firstName != "" && strings.Contains(firstName, " ") {
		nameParts := strings.SplitN(firstName, " ", 2)
		firstName = nameParts[0]
		lastName = nameParts[1]
		c.Logger().Printf("Split name: FirstName='%s', LastName='%s'", firstName, lastName)
	}

	user, err := h.userService.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		c.Logger().Printf("Creating new user from OAuth: %s", email)

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

		isAdmin := shouldBeAdmin(email, userInfo.Groups)
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
			c.Logger().Errorf("Failed to create user: %v", err)
			return c.Redirect(http.StatusSeeOther, "/auth/login?error=user_creation_failed")
		}
		user = &newUser

		if isAdmin {
			c.Logger().Printf("Created admin user from OAuth: %s (groups: %v)", email, userInfo.Groups)
		}
	} else {
		c.Logger().Printf("OAuth login for existing user: %s", email)

		isAdmin := shouldBeAdmin(email, userInfo.Groups)
		expectedRole := roleUser
		if isAdmin {
			expectedRole = roleAdmin
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
			c.Logger().Printf("Updated admin status for user %s: %v (groups: %v)", email, isAdmin, userInfo.Groups)
		}
		if user.AuthProvider != "oauth" {
			user.AuthProvider = "oauth"
			updated = true
		}

		if updated {
			h.userService.UpdateUser(c.Request().Context(), user)
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
