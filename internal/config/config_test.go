package config

import (
	"os"
	"testing"
)

func TestValidate(t *testing.T) {
	// Valid base config for tests
	validConfig := func() *Config {
		return &Config{
			ServerPort:             "3000",
			MetricsPort:            "9090",
			SessionSecret:          "test-secret-that-is-long-enough-for-validation",
			SessionMaxAge:          604800,
			ShutdownTimeoutSeconds: 10,
			LogLevel:               "INFO",
			EnableLocalLogin:       true,
		}
	}

	tests := []struct {
		name        string
		modify      func(c *Config)
		wantErr     bool
		errContains string
	}{
		{
			name:    "Valid config",
			modify:  func(_ *Config) {},
			wantErr: false,
		},
		{
			name:        "Empty SERVER_PORT",
			modify:      func(c *Config) { c.ServerPort = "" },
			wantErr:     true,
			errContains: "SERVER_PORT must not be empty",
		},
		{
			name:        "Invalid SERVER_PORT",
			modify:      func(c *Config) { c.ServerPort = "abc" },
			wantErr:     true,
			errContains: "SERVER_PORT must be a valid port",
		},
		{
			name:        "Port 0",
			modify:      func(c *Config) { c.ServerPort = "0" },
			wantErr:     true,
			errContains: "SERVER_PORT must be a valid port",
		},
		{
			name:        "Port too high",
			modify:      func(c *Config) { c.ServerPort = "70000" },
			wantErr:     true,
			errContains: "SERVER_PORT must be a valid port",
		},
		{
			name:        "Invalid METRICS_PORT",
			modify:      func(c *Config) { c.MetricsPort = "not-a-port" },
			wantErr:     true,
			errContains: "METRICS_PORT must be a valid port",
		},
		{
			name:        "SERVER_PORT equals METRICS_PORT",
			modify:      func(c *Config) { c.ServerPort = "3000"; c.MetricsPort = "3000" },
			wantErr:     true,
			errContains: "SERVER_PORT and METRICS_PORT must be different",
		},
		{
			name:        "Negative SESSION_MAX_AGE",
			modify:      func(c *Config) { c.SessionMaxAge = -1 },
			wantErr:     true,
			errContains: "SESSION_MAX_AGE must be positive",
		},
		{
			name:        "SESSION_MAX_AGE exceeds 30 days",
			modify:      func(c *Config) { c.SessionMaxAge = 2592001 },
			wantErr:     true,
			errContains: "SESSION_MAX_AGE must not exceed",
		},
		{
			name:        "Negative SHUTDOWN_TIMEOUT",
			modify:      func(c *Config) { c.ShutdownTimeoutSeconds = 0 },
			wantErr:     true,
			errContains: "SHUTDOWN_TIMEOUT_SECONDS must be positive",
		},
		{
			name:        "Invalid LOG_LEVEL",
			modify:      func(c *Config) { c.LogLevel = "TRACE" },
			wantErr:     true,
			errContains: "LOG_LEVEL must be one of",
		},
		{
			name:        "Short SESSION_SECRET",
			modify:      func(c *Config) { c.SessionSecret = "too-short" },
			wantErr:     true,
			errContains: "SESSION_SECRET must be at least 32 characters",
		},
		{
			name: "Partial OAuth config missing client ID",
			modify: func(c *Config) {
				c.OAuthClientSecret = "some-secret"
				c.OAuthIssuer = "https://auth.example.com"
			},
			wantErr:     true,
			errContains: "OAUTH_CLIENT_ID must be set",
		},
		{
			name: "Partial OAuth config missing issuer",
			modify: func(c *Config) {
				c.OAuthClientID = "client-id"
				c.OAuthClientSecret = "some-secret"
			},
			wantErr:     true,
			errContains: "OAUTH_ISSUER must be set",
		},
		{
			name: "Partial OAuth config missing secret",
			modify: func(c *Config) {
				c.OAuthClientID = "client-id"
				c.OAuthIssuer = "https://auth.example.com"
			},
			wantErr:     true,
			errContains: "OAUTH_CLIENT_SECRET must be set",
		},
		{
			name:    "Valid CSP report URI",
			modify:  func(c *Config) { c.CSPReportURI = "https://csp.example.com/report" },
			wantErr: false,
		},
		{
			name:        "CSP report URI with semicolon injection",
			modify:      func(c *Config) { c.CSPReportURI = "https://evil.com; script-src 'unsafe-eval'" },
			wantErr:     true,
			errContains: "CSP_REPORT_URI must be a valid HTTP(S) URL",
		},
		{
			name:        "CSP report URI with invalid scheme",
			modify:      func(c *Config) { c.CSPReportURI = "ftp://example.com/report" },
			wantErr:     true,
			errContains: "CSP_REPORT_URI must be a valid HTTP(S) URL",
		},
		{
			name: "No auth method enabled",
			modify: func(c *Config) {
				c.EnableLocalLogin = false
				// OAuth not configured = not enabled
			},
			wantErr:     true,
			errContains: "at least one authentication method",
		},
		{
			name: "Only OAuth enabled (no local login)",
			modify: func(c *Config) {
				c.EnableLocalLogin = false
				c.OAuthClientID = "client-id"
				c.OAuthClientSecret = "some-secret"
				c.OAuthIssuer = "https://auth.example.com"
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.modify(cfg)

			err := cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateProduction(t *testing.T) {
	tests := []struct { // #nosec G101 -- test credentials, not real secrets
		name          string
		environment   string
		sessionSecret string
		oauthClientID string
		oauthSecret   string
		oauthIssuer   string
		wantErr       bool
		errContains   string
	}{
		{
			name:          "Development environment - no validation",
			environment:   "development",
			sessionSecret: "dev-secret-change-in-production-not-for-use",
			wantErr:       false,
		},
		{
			name:          "Production with default SESSION_SECRET",
			environment:   "production",
			sessionSecret: "dev-secret-change-in-production-not-for-use",
			wantErr:       true,
			errContains:   "SESSION_SECRET must be changed",
		},
		{
			name:          "Production with valid SESSION_SECRET",
			environment:   "production",
			sessionSecret: "this-is-a-very-secure-secret-key-with-more-than-32-chars",
			wantErr:       false,
		},
		{
			name:          "Production with OAuth and short secret",
			environment:   "production",
			sessionSecret: "this-is-a-very-secure-secret-key-with-more-than-32-chars",
			oauthClientID: "client-id",
			oauthIssuer:   "https://auth.example.com",
			oauthSecret:   "short",
			wantErr:       true,
			errContains:   "at least 16 characters",
		},
		{
			name:          "Production with OAuth and valid secret",
			environment:   "production",
			sessionSecret: "this-is-a-very-secure-secret-key-with-more-than-32-chars",
			oauthClientID: "client-id",
			oauthIssuer:   "https://auth.example.com",
			oauthSecret:   "valid-oauth-secret-16-chars", // #nosec G101 -- test credential, not real
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config
			cfg := &Config{
				Environment:       tt.environment,
				SessionSecret:     tt.sessionSecret,
				SessionMaxAge:     604800, // 7 days - required for validation
				OAuthClientID:     tt.oauthClientID,
				OAuthClientSecret: tt.oauthSecret,
				OAuthIssuer:       tt.oauthIssuer,
			}

			// Run validation
			err := cfg.ValidateProduction()

			// Check result
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateProduction() expected error but got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateProduction() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateProduction() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateProduction_Integration(t *testing.T) {
	// Test with actual environment variables
	t.Run("Production mode via GO_ENV", func(t *testing.T) {
		// Set production environment
		if err := os.Setenv("GO_ENV", "production"); err != nil {
			t.Fatalf("Failed to set GO_ENV: %v", err)
		}
		if err := os.Setenv("SESSION_SECRET", "test-secret-that-is-long-enough-for-prod-validation"); err != nil {
			t.Fatalf("Failed to set SESSION_SECRET: %v", err)
		}
		defer func() {
			_ = os.Unsetenv("GO_ENV")
			_ = os.Unsetenv("SESSION_SECRET")
		}()

		cfg := Load()
		err := cfg.ValidateProduction()
		if err != nil {
			t.Errorf("ValidateProduction() with valid secrets returned error: %v", err)
		}
	})

	t.Run("Production mode with default secret fails", func(t *testing.T) {
		if err := os.Setenv("GO_ENV", "production"); err != nil {
			t.Fatalf("Failed to set GO_ENV: %v", err)
		}
		if err := os.Setenv("SESSION_SECRET", "dev-secret-change-in-production-not-for-use"); err != nil {
			t.Fatalf("Failed to set SESSION_SECRET: %v", err)
		}
		defer func() {
			_ = os.Unsetenv("GO_ENV")
			_ = os.Unsetenv("SESSION_SECRET")
		}()

		cfg := Load()
		err := cfg.ValidateProduction()
		if err == nil {
			t.Error("ValidateProduction() expected error with default secret in production")
		}
	})
}

func TestLoadWithLogLevel(t *testing.T) {
	tests := []struct {
		name             string
		envLogLevel      string
		envGoEnv         string
		expectedLogLevel string
	}{
		{
			name:             "LOG_LEVEL=DEBUG explicitly set",
			envLogLevel:      "DEBUG",
			envGoEnv:         "production",
			expectedLogLevel: "DEBUG",
		},
		{
			name:             "LOG_LEVEL=INFO explicitly set",
			envLogLevel:      "INFO",
			envGoEnv:         "development",
			expectedLogLevel: "INFO",
		},
		{
			name:             "LOG_LEVEL=WARN explicitly set",
			envLogLevel:      "WARN",
			envGoEnv:         "production",
			expectedLogLevel: "WARN",
		},
		{
			name:             "LOG_LEVEL=ERROR explicitly set",
			envLogLevel:      "ERROR",
			envGoEnv:         "development",
			expectedLogLevel: "ERROR",
		},
		{
			name:             "Default LOG_LEVEL in production (INFO)",
			envLogLevel:      "",
			envGoEnv:         "production",
			expectedLogLevel: "INFO",
		},
		{
			name:             "Default LOG_LEVEL in development (DEBUG)",
			envLogLevel:      "",
			envGoEnv:         "development",
			expectedLogLevel: "DEBUG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			_ = os.Unsetenv("LOG_LEVEL")
			_ = os.Unsetenv("GO_ENV")

			// Set test environment
			if tt.envLogLevel != "" {
				if err := os.Setenv("LOG_LEVEL", tt.envLogLevel); err != nil {
					t.Fatalf("Failed to set LOG_LEVEL: %v", err)
				}
			}
			if tt.envGoEnv != "" {
				if err := os.Setenv("GO_ENV", tt.envGoEnv); err != nil {
					t.Fatalf("Failed to set GO_ENV: %v", err)
				}
			}

			// Load config
			cfg := Load()

			// Verify LOG_LEVEL
			if cfg.LogLevel != tt.expectedLogLevel {
				t.Errorf("Expected LogLevel=%q, got %q", tt.expectedLogLevel, cfg.LogLevel)
			}

			// Cleanup
			_ = os.Unsetenv("LOG_LEVEL")
			_ = os.Unsetenv("GO_ENV")
		})
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Save current environment
	originalEnv := os.Getenv("GO_ENV")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("GO_ENV", originalEnv)
		} else {
			_ = os.Unsetenv("GO_ENV")
		}
	}()

	// Clear GO_ENV to ensure defaults
	_ = os.Unsetenv("GO_ENV")
	_ = os.Unsetenv("LOG_LEVEL")

	cfg := Load()

	// Verify LOG_LEVEL defaults to DEBUG in development
	if cfg.Environment != "development" {
		t.Errorf("Expected Environment=development, got %s", cfg.Environment)
	}
	if cfg.LogLevel != "DEBUG" {
		t.Errorf("Expected LogLevel=DEBUG in development, got %s", cfg.LogLevel)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
