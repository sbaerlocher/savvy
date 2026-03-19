// Package config handles application configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// ConfigContextKey is the key for storing config in context
	ConfigContextKey ContextKey = "config"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	DatabaseURL            string
	ServerPort             string
	MetricsPort            string // Separate port for Prometheus metrics (default: 9090)
	SessionSecret          string // #nosec G117 -- struct field name, not a hardcoded secret
	SessionMaxAge          int    // Session max age in seconds (default: 604800 = 7 days)
	ShutdownTimeoutSeconds int    // Graceful shutdown timeout in seconds (default: 10)
	Environment            string
	LogLevel               string // Logging level: DEBUG, INFO, WARN, ERROR
	AutoMigrate            bool   // Enable/disable automatic migrations
	OTelEnabled            bool
	OTelEndpoint           string
	ServiceName            string
	ServiceVersion         string
	OAuthClientID          string
	OAuthClientSecret      string
	OAuthIssuer            string
	OAuthRedirectURL       string
	OAuthAdminEmails       []string       // Comma-separated list of admin emails
	OAuthAdminGroup        string         // OIDC group name for admins
	FrontendURL            string         // OAuth success redirect URL (OAUTH_SUCCESS_URL or deprecated FRONTEND_URL)
	EnableCards            bool           // Enable/disable cards feature
	EnableVouchers         bool           // Enable/disable vouchers feature
	EnableGiftCards        bool           // Enable/disable gift cards feature
	EnableLocalLogin       bool           // Enable/disable email/password login
	EnableRegistration     bool           // Enable/disable user registration
	CSPReportURI           string         // Optional CSP report-uri endpoint for violation reporting
	SMTPHost               string         // SMTP server host
	SMTPPort               int            // SMTP server port
	SMTPUsername           string         // SMTP authentication username
	SMTPPassword           string         // SMTP authentication password
	SMTPFromEmail          string         // Sender email address
	SMTPFromName           string         // Sender display name
	SMTPUseTLS             bool           // Use TLS for SMTP connection
	VAPIDPublicKey         string         // VAPID public key for Web Push
	VAPIDPrivateKey        string         // VAPID private key for Web Push
	VAPIDSubject           string         // VAPID subject (mailto: or URL)
	EnableExpiryReminders  bool           // Enable/disable expiry reminders
	ReminderDaysBefore     []int          // Days before expiry to send reminders (e.g., 7,3,1)
	ReminderCheckTime      string         // Daily time for reminder check in HH:MM format (e.g., "08:00")
	Enable2FA              bool           // Enable/disable two-factor authentication
	TOTPIssuer             string         // TOTP issuer name shown in authenticator apps
	TOTPEncryptionKey      string         // AES-256 encryption key for TOTP secrets
	Timezone               string         // IANA timezone for date calculations (e.g., "Europe/Zurich")
	Location               *time.Location // Parsed timezone location
	CORSAllowedOrigins     []string       // Allowed CORS origins (comma-separated, development only)
}

// Load reads configuration from environment variables and returns a Config instance
func Load() *Config {
	// Support both new and old env var names for backward compatibility
	env := getEnvWithFallback("ENVIRONMENT", "GO_ENV", "development")
	isProduction := env == "production"

	// OTel should be enabled by default in production
	otelEnabledDefault := isProduction

	// Default log level: INFO for production, DEBUG for development
	defaultLogLevel := "INFO"
	if !isProduction {
		defaultLogLevel = "DEBUG"
	}

	timezone := getEnv("TIMEZONE", "UTC")
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	return &Config{
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable"),
		ServerPort:             getEnvWithFallback("SERVER_PORT", "PORT", "3000"),
		MetricsPort:            getEnv("METRICS_PORT", "9090"), // Separate port for Prometheus metrics
		SessionSecret:          getEnv("SESSION_SECRET", "dev-secret-change-in-production-not-for-use"),
		SessionMaxAge:          getIntEnv("SESSION_MAX_AGE", 604800),      // 7 days default
		ShutdownTimeoutSeconds: getIntEnv("SHUTDOWN_TIMEOUT_SECONDS", 10), // 10 seconds default
		Environment:            env,
		LogLevel:               getEnv("LOG_LEVEL", defaultLogLevel),
		AutoMigrate:            getBoolEnv("AUTO_MIGRATE", true),               // Default true for dev convenience
		OTelEnabled:            getBoolEnv("OTEL_ENABLED", otelEnabledDefault), // Default true in production
		OTelEndpoint:           getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		ServiceName:            getEnv("OTEL_SERVICE_NAME", "savvy"),
		ServiceVersion:         getEnv("OTEL_SERVICE_VERSION", "2.0.0"),
		OAuthClientID:          getEnv("OAUTH_CLIENT_ID", ""),
		OAuthClientSecret:      getEnv("OAUTH_CLIENT_SECRET", ""),
		OAuthIssuer:            getEnv("OAUTH_ISSUER", ""),
		OAuthRedirectURL:       getEnv("OAUTH_REDIRECT_URL", "http://localhost:3000/auth/oauth/callback"),
		OAuthAdminEmails:       getEnvSlice("OAUTH_ADMIN_EMAILS", []string{}),
		OAuthAdminGroup:        getEnv("OAUTH_ADMIN_GROUP", ""),
		FrontendURL:            getEnvWithFallback("OAUTH_SUCCESS_URL", "FRONTEND_URL", "http://localhost:5173"),
		EnableCards:            getBoolEnv("ENABLE_CARDS", true),        // Default true
		EnableVouchers:         getBoolEnv("ENABLE_VOUCHERS", true),     // Default true
		EnableGiftCards:        getBoolEnv("ENABLE_GIFT_CARDS", true),   // Default true
		EnableLocalLogin:       getBoolEnv("ENABLE_LOCAL_LOGIN", true),  // Default true
		EnableRegistration:     getBoolEnv("ENABLE_REGISTRATION", true), // Default true
		CSPReportURI:           getEnv("CSP_REPORT_URI", ""),
		SMTPHost:               getEnv("SMTP_HOST", ""),
		SMTPPort:               getIntEnv("SMTP_PORT", 587),
		SMTPUsername:           getEnv("SMTP_USERNAME", ""),
		SMTPPassword:           getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:          getEnv("SMTP_FROM_EMAIL", ""),
		SMTPFromName:           getEnv("SMTP_FROM_NAME", "Savvy"),
		SMTPUseTLS:             getBoolEnv("SMTP_USE_TLS", true),
		VAPIDPublicKey:         getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey:        getEnv("VAPID_PRIVATE_KEY", ""),
		VAPIDSubject:           getEnv("VAPID_SUBJECT", ""),
		EnableExpiryReminders:  getBoolEnv("ENABLE_EXPIRY_REMINDERS", true),
		ReminderDaysBefore:     getIntSliceEnv("REMINDER_DAYS_BEFORE", []int{7, 3, 1}),
		ReminderCheckTime:      getEnv("REMINDER_CHECK_TIME", "08:00"),
		Enable2FA:              getBoolEnv("ENABLE_2FA", false),
		TOTPIssuer:             getEnv("TOTP_ISSUER", "Savvy"),
		TOTPEncryptionKey:      getEnv("TOTP_ENCRYPTION_KEY", ""),
		Timezone:               timezone,
		Location:               loc,
		CORSAllowedOrigins:     getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
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

// IsSMTPEnabled returns true if SMTP is configured
func (c *Config) IsSMTPEnabled() bool {
	return c.SMTPHost != "" && c.SMTPFromEmail != ""
}

// IsPushEnabled returns true if VAPID keys are configured for Web Push
func (c *Config) IsPushEnabled() bool {
	return c.VAPIDPublicKey != "" && c.VAPIDPrivateKey != "" && c.VAPIDSubject != ""
}

// Is2FAEnabled returns true if 2FA is enabled and encryption key is configured
func (c *Config) Is2FAEnabled() bool {
	return c.Enable2FA && c.TOTPEncryptionKey != ""
}

// Validate checks configuration constraints that apply in all environments.
// This catches configuration errors early, regardless of GO_ENV.
func (c *Config) Validate() error {
	if err := c.validatePorts(); err != nil {
		return err
	}

	if c.SessionMaxAge <= 0 {
		return errors.New("SESSION_MAX_AGE must be positive")
	}
	if c.SessionMaxAge > 2592000 {
		return errors.New("SESSION_MAX_AGE must not exceed 2592000 (30 days)")
	}
	if c.ShutdownTimeoutSeconds <= 0 {
		return errors.New("SHUTDOWN_TIMEOUT_SECONDS must be positive")
	}

	validLogLevels := map[string]bool{"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true}
	if !validLogLevels[strings.ToUpper(c.LogLevel)] {
		return errors.New("LOG_LEVEL must be one of: DEBUG, INFO, WARN, ERROR")
	}

	if err := c.validateOAuth(); err != nil {
		return err
	}

	if len(c.SessionSecret) < 32 {
		return errors.New("SESSION_SECRET must be at least 32 characters")
	}
	// Warn loudly when the known-weak default secret is used outside local development.
	// Production enforcement is handled separately in ValidateProduction(), but staging
	// environments must also not run with the default because they often use real credentials.
	if c.SessionSecret == "dev-secret-change-in-production-not-for-use" && c.Environment != "development" {
		slog.Warn("SESSION_SECRET is set to the default development value — sessions can be forged; set a strong secret via SESSION_SECRET",
			"environment", c.Environment)
	}

	if err := c.validateURLs(); err != nil {
		return err
	}

	if err := c.validateSMTP(); err != nil {
		return err
	}

	if !c.EnableLocalLogin && !c.IsOAuthEnabled() {
		return errors.New("at least one authentication method must be enabled (ENABLE_LOCAL_LOGIN or OAuth)")
	}

	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return fmt.Errorf("TIMEZONE must be a valid IANA timezone (e.g., Europe/Zurich): %w", err)
	}

	if c.ReminderCheckTime != "" {
		if _, err := time.Parse("15:04", c.ReminderCheckTime); err != nil {
			return fmt.Errorf("REMINDER_CHECK_TIME must be in HH:MM format (e.g., 08:00): %w", err)
		}
	}

	if c.Is2FAEnabled() && len(c.TOTPEncryptionKey) != 32 {
		return fmt.Errorf("TOTP_ENCRYPTION_KEY must be exactly 32 bytes when 2FA is enabled, got %d", len(c.TOTPEncryptionKey))
	}

	return nil
}

// validatePorts checks that SERVER_PORT and METRICS_PORT are valid and don't collide.
func (c *Config) validatePorts() error {
	if c.ServerPort == "" {
		return errors.New("SERVER_PORT must not be empty")
	}
	if port, err := strconv.Atoi(c.ServerPort); err != nil || port < 1 || port > 65535 {
		return errors.New("SERVER_PORT must be a valid port number (1-65535)")
	}
	if c.MetricsPort != "" {
		if port, err := strconv.Atoi(c.MetricsPort); err != nil || port < 1 || port > 65535 {
			return errors.New("METRICS_PORT must be a valid port number (1-65535)")
		}
	}
	if c.ServerPort == c.MetricsPort {
		return errors.New("SERVER_PORT and METRICS_PORT must be different")
	}
	return nil
}

// validateOAuth checks that OAuth fields are consistently configured.
func (c *Config) validateOAuth() error {
	isPartiallyConfigured := c.OAuthClientID != "" || c.OAuthClientSecret != "" || c.OAuthIssuer != ""
	if !isPartiallyConfigured {
		return nil
	}
	if c.OAuthClientID == "" {
		return errors.New("OAUTH_CLIENT_ID must be set when OAuth is partially configured")
	}
	if c.OAuthIssuer == "" {
		return errors.New("OAUTH_ISSUER must be set when OAuth is partially configured")
	}
	if c.OAuthClientSecret == "" {
		return errors.New("OAUTH_CLIENT_SECRET must be set when OAuth is enabled")
	}
	return nil
}

// validateURLs checks that CSP_REPORT_URI, FrontendURL, and OTel endpoint are valid.
func (c *Config) validateURLs() error {
	if c.CSPReportURI != "" {
		u, err := url.Parse(c.CSPReportURI)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || strings.Contains(c.CSPReportURI, ";") {
			return errors.New("CSP_REPORT_URI must be a valid HTTP(S) URL without semicolons")
		}
	}
	if c.FrontendURL != "" {
		u, err := url.Parse(c.FrontendURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errors.New("OAUTH_SUCCESS_URL/FRONTEND_URL must be a valid HTTP(S) URL")
		}
	}
	if c.OTelEnabled && c.OTelEndpoint != "" {
		endpoint := c.OTelEndpoint
		endpoint = strings.TrimPrefix(endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		if strings.ContainsAny(endpoint, "?#@") || strings.Contains(endpoint, "..") {
			return errors.New("OTEL_EXPORTER_OTLP_ENDPOINT must be a valid host:port or URL without query parameters")
		}
		// Block SSRF: reject RFC-1918 and link-local IP literals as OTel endpoint hosts.
		// Hostname-based targets (e.g., "otel-collector:4318") are permitted as they are
		// intentionally internal services, whereas raw IP literals should never point to
		// cloud metadata endpoints (169.254.x.x) or unexpected internal ranges.
		host := endpoint
		if h, _, err := net.SplitHostPort(endpoint); err == nil {
			host = h
		}
		if ip := net.ParseIP(host); ip != nil {
			if isInternalIP(ip) {
				return errors.New("OTEL_EXPORTER_OTLP_ENDPOINT must not be an internal or link-local IP address")
			}
		}
	}
	return nil
}

// internalCIDRs lists RFC-1918 private ranges and link-local addresses.
var internalCIDRs = func() []*net.IPNet {
	cidrs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16", // link-local (AWS metadata: 169.254.169.254)
		"127.0.0.0/8",    // loopback
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique-local
		"fe80::/10",      // IPv6 link-local
	}
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, n, _ := net.ParseCIDR(cidr)
		if n != nil {
			nets = append(nets, n)
		}
	}
	return nets
}()

// isInternalIP returns true if ip falls within any private/link-local/loopback range.
func isInternalIP(ip net.IP) bool {
	for _, cidr := range internalCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// validateSMTP checks that SMTP fields are consistently configured.
func (c *Config) validateSMTP() error {
	isPartiallyConfigured := c.SMTPHost != "" || c.SMTPFromEmail != ""
	if !isPartiallyConfigured {
		return nil
	}
	if c.SMTPHost == "" {
		return errors.New("SMTP_HOST must be set when SMTP is partially configured")
	}
	if c.SMTPFromEmail == "" {
		return errors.New("SMTP_FROM_EMAIL must be set when SMTP is partially configured")
	}
	if c.SMTPPort < 1 || c.SMTPPort > 65535 {
		return errors.New("SMTP_PORT must be a valid port number (1-65535)")
	}
	return nil
}

// ValidateProduction validates that production-critical secrets are properly configured.
// This prevents accidentally deploying with default development secrets.
// Call Validate() first for general checks.
func (c *Config) ValidateProduction() error {
	if !c.IsProduction() {
		return nil // Skip validation in non-production environments
	}

	// Check SESSION_SECRET is not using default development value
	if c.SessionSecret == "dev-secret-change-in-production-not-for-use" {
		return errors.New("SESSION_SECRET must be changed in production (currently using default dev value)")
	}

	// If OAuth is configured, validate secret strength
	if c.IsOAuthEnabled() {
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

// getEnvWithFallback tries the primary key first, then falls back to the legacy key
// This enables backward compatibility when renaming environment variables
func getEnvWithFallback(primaryKey, fallbackKey, defaultValue string) string {
	if value := os.Getenv(primaryKey); value != "" {
		return value
	}
	if value := os.Getenv(fallbackKey); value != "" {
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

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return intVal
	}
	return defaultValue
}

func getIntSliceEnv(key string, defaultValue []int) []int {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]int, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				intVal, err := strconv.Atoi(trimmed)
				if err != nil {
					return defaultValue
				}
				result = append(result, intVal)
			}
		}
		if len(result) > 0 {
			return result
		}
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
