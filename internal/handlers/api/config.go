// Package api provides JSON API handlers for the application.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"net/http"
	"savvy/internal/config"

	"github.com/labstack/echo/v5"
)

// ConfigHandler handles configuration API endpoints.
type ConfigHandler struct {
	config *config.Config
}

// NewConfigHandler creates a new config API handler.
func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{config: cfg}
}

// ConfigResponse represents the public app configuration
type ConfigResponse struct {
	OAuth         OAuthConfig `json:"oauth"`
	Features      Features    `json:"features"`
	LocalLogin    bool        `json:"local_login_enabled"`
	Registration  bool        `json:"registration_enabled"`
	SMTPEnabled   bool        `json:"smtp_enabled"`
	PushEnabled   bool        `json:"push_enabled"`
	TwoFAEnabled  bool        `json:"two_factor_enabled"`
	IsDevelopment bool        `json:"is_development"`
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	Enabled  bool   `json:"enabled"`
	LoginURL string `json:"login_url,omitempty"`
}

// Features represents feature toggles
type Features struct {
	Cards     bool `json:"cards"`
	Vouchers  bool `json:"vouchers"`
	GiftCards bool `json:"gift_cards"`
}

// GetConfig returns public app configuration
// GET /api/v1/config
func (h *ConfigHandler) GetConfig(c echo.Context) error {
	response := ConfigResponse{
		OAuth: OAuthConfig{
			Enabled: h.config.IsOAuthEnabled(),
		},
		Features: Features{
			Cards:     h.config.EnableCards,
			Vouchers:  h.config.EnableVouchers,
			GiftCards: h.config.EnableGiftCards,
		},
		LocalLogin:    h.config.EnableLocalLogin,
		Registration:  h.config.EnableRegistration,
		SMTPEnabled:   h.config.IsSMTPEnabled(),
		PushEnabled:   h.config.IsPushEnabled(),
		TwoFAEnabled:  h.config.Is2FAEnabled(),
		IsDevelopment: !h.config.IsProduction(),
	}

	// Add OAuth login URL if OAuth is enabled
	if response.OAuth.Enabled {
		response.OAuth.LoginURL = "/auth/oauth/login"
	}

	return c.JSON(http.StatusOK, response)
}
