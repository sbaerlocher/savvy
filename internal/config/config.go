// Package config handles application configuration from environment variables.
package config

import (
	"errors"
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
	DatabaseURL        string
	ServerPort         string
	SessionSecret      string
	Environment        string
	LogLevel           string // Logging level: DEBUG, INFO, WARN, ERROR
	AutoMigrate        bool   // Enable/disable automatic migrations
	OTelEnabled        bool
	OTelEndpoint       string
	ServiceName        string
	ServiceVersion     string
	OAuthClientID      string
	OAuthClientSecret  string
	OAuthIssuer        string
	OAuthRedirectURL   string
	OAuthAdminEmails   []string // Comma-separated list of admin emails
	OAuthAdminGroup    string   // OIDC group name for admins
	EnableCards        bool     // Enable/disable cards feature
	EnableVouchers     bool     // Enable/disable vouchers feature
	EnableGiftCards    bool     // Enable/disable gift cards feature
	EnableLocalLogin   bool     // Enable/disable email/password login
	EnableRegistration bool     // Enable/disable user registration
}

// Load reads configuration from environment variables and returns a Config instance
func Load() *Config {
	env := getEnv("GO_ENV", "development")
	isProduction := env == "production"

	// OTel should be enabled by default in production
	otelEnabledDefault := isProduction

	// Default log level: INFO for production, DEBUG for development
	defaultLogLevel := "INFO"
	if !isProduction {
		defaultLogLevel = "DEBUG"
	}

	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable"),
		ServerPort:         getEnv("PORT", "3000"),
		SessionSecret:      getEnv("SESSION_SECRET", "dev-secret-change-in-production"),
		Environment:        env,
		LogLevel:           getEnv("LOG_LEVEL", defaultLogLevel),
		AutoMigrate:        getBoolEnv("AUTO_MIGRATE", true),               // Default true for dev convenience
		OTelEnabled:        getBoolEnv("OTEL_ENABLED", otelEnabledDefault), // Default true in production
		OTelEndpoint:       getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		ServiceName:        getEnv("OTEL_SERVICE_NAME", "savvy"),
		ServiceVersion:     getEnv("OTEL_SERVICE_VERSION", "1.1.0"),
		OAuthClientID:      getEnv("OAUTH_CLIENT_ID", ""),
		OAuthClientSecret:  getEnv("OAUTH_CLIENT_SECRET", ""),
		OAuthIssuer:        getEnv("OAUTH_ISSUER", ""),
		OAuthRedirectURL:   getEnv("OAUTH_REDIRECT_URL", "http://localhost:3000/auth/oauth/callback"),
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

// ValidateProduction validates that production-critical secrets are properly configured
// This prevents accidentally deploying with default development secrets
func (c *Config) ValidateProduction() error {
	if !c.IsProduction() {
		return nil // Skip validation in non-production environments
	}

	// Check SESSION_SECRET is not using default development value
	if c.SessionSecret == "dev-secret-change-in-production" {
		return errors.New("SESSION_SECRET must be changed in production (currently using default dev value)")
	}

	// Check SESSION_SECRET has minimum length
	if len(c.SessionSecret) < 32 {
		return errors.New("SESSION_SECRET must be at least 32 characters in production")
	}

	// If OAuth is enabled, validate OAuth secrets
	if c.OAuthClientID != "" || c.OAuthIssuer != "" {
		if c.OAuthClientSecret == "" {
			return errors.New("OAUTH_CLIENT_SECRET must be set in production when OAuth is enabled")
		}
		if len(c.OAuthClientSecret) < 16 {
			return errors.New("OAUTH_CLIENT_SECRET must be at least 16 characters in production")
		}
	}

	return nil
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
