// Package config handles application configuration from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// ConfigContextKey is the key for storing config in context
	ConfigContextKey ContextKey = "config"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	DatabaseURL       string
	ServerPort        string
	SessionSecret     string
	Environment       string
	AutoMigrate       bool // Enable/disable automatic migrations
	OTelEnabled       bool
	OTelEndpoint      string
	ServiceName       string
	ServiceVersion    string
	OAuthClientID     string
	OAuthClientSecret string
	OAuthIssuer       string
	OAuthRedirectURL  string
	OAuthAdminEmails    []string // Comma-separated list of admin emails
	OAuthAdminGroup     string   // OIDC group name for admins
	EnableCards         bool     // Enable/disable cards feature
	EnableVouchers      bool     // Enable/disable vouchers feature
	EnableGiftCards     bool     // Enable/disable gift cards feature
	EnableLocalLogin    bool     // Enable/disable email/password login
	EnableRegistration  bool     // Enable/disable user registration
}

// Load reads configuration from environment variables and returns a Config instance
func Load() *Config {
	env := getEnv("GO_ENV", "development")
	isProduction := env == "production"

	// OTel should be enabled by default in production
	otelEnabledDefault := isProduction

	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable"),
		ServerPort:        getEnv("PORT", "3000"),
		SessionSecret:     getEnv("SESSION_SECRET", "dev-secret-change-in-production"),
		Environment:       env,
		AutoMigrate:       getBoolEnv("AUTO_MIGRATE", true), // Default true for dev convenience
		OTelEnabled:       getBoolEnv("OTEL_ENABLED", otelEnabledDefault), // Default true in production
		OTelEndpoint:      getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		ServiceName:       getEnv("OTEL_SERVICE_NAME", "savvy"),
		ServiceVersion:    getEnv("OTEL_SERVICE_VERSION", "1.1.0"),
		OAuthClientID:     getEnv("OAUTH_CLIENT_ID", ""),
		OAuthClientSecret: getEnv("OAUTH_CLIENT_SECRET", ""),
		OAuthIssuer:       getEnv("OAUTH_ISSUER", ""),
		OAuthRedirectURL:  getEnv("OAUTH_REDIRECT_URL", "http://localhost:3000/auth/oauth/callback"),
		OAuthAdminEmails:   getEnvSlice("OAUTH_ADMIN_EMAILS", []string{}),
		OAuthAdminGroup:    getEnv("OAUTH_ADMIN_GROUP", ""),
		EnableCards:        getBoolEnv("ENABLE_CARDS", true),        // Default true
		EnableVouchers:     getBoolEnv("ENABLE_VOUCHERS", true),     // Default true
		EnableGiftCards:    getBoolEnv("ENABLE_GIFT_CARDS", true),   // Default true
		EnableLocalLogin:   getBoolEnv("ENABLE_LOCAL_LOGIN", true),  // Default true
		EnableRegistration: getBoolEnv("ENABLE_REGISTRATION", true), // Default true
	}
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsOAuthEnabled returns true if OAuth is configured
func (c *Config) IsOAuthEnabled() bool {
	return c.OAuthClientID != "" && c.OAuthClientSecret != "" && c.OAuthIssuer != ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return boolVal
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim spaces
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}
