// Package oauth handles OAuth2/OIDC authentication with Authentik
package oauth

import (
	"context"
	"fmt"
	"savvy/internal/config"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Provider holds OAuth2 provider configuration
type Provider struct {
	Config   *oauth2.Config
	Verifier *oidc.IDTokenVerifier
}

// NewProvider creates a new OAuth provider instance
func NewProvider(cfg *config.Config) (*Provider, error) {
	if !cfg.IsOAuthEnabled() {
		return nil, fmt.Errorf("OAuth is not configured")
	}

	ctx := context.Background()

	// Initialize OIDC provider
	provider, err := oidc.NewProvider(ctx, cfg.OAuthIssuer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	// Configure OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.OAuthClientID,
		ClientSecret: cfg.OAuthClientSecret,
		RedirectURL:  cfg.OAuthRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Create ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.OAuthClientID,
	})

	return &Provider{
		Config:   oauth2Config,
		Verifier: verifier,
	}, nil
}

// UserInfo represents user information from OIDC
type UserInfo struct {
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	FirstName         string   `json:"given_name"`
	LastName          string   `json:"family_name"`
	PreferredUsername string   `json:"preferred_username"`
	Groups            []string `json:"groups"` // Authentik groups
}
