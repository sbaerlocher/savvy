package config

import (
	"os"
	"testing"
)

func TestValidateProduction(t *testing.T) {
	tests := []struct {
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
			sessionSecret: "dev-secret-change-in-production",
			wantErr:       false,
		},
		{
			name:          "Production with default SESSION_SECRET",
			environment:   "production",
			sessionSecret: "dev-secret-change-in-production",
			wantErr:       true,
			errContains:   "SESSION_SECRET must be changed",
		},
		{
			name:          "Production with short SESSION_SECRET",
			environment:   "production",
			sessionSecret: "short",
			wantErr:       true,
			errContains:   "at least 32 characters",
		},
		{
			name:          "Production with valid SESSION_SECRET",
			environment:   "production",
			sessionSecret: "this-is-a-very-secure-secret-key-with-more-than-32-chars",
			wantErr:       false,
		},
		{
			name:          "Production with OAuth but no secret",
			environment:   "production",
			sessionSecret: "this-is-a-very-secure-secret-key-with-more-than-32-chars",
			oauthClientID: "client-id",
			oauthIssuer:   "https://auth.example.com",
			oauthSecret:   "",
			wantErr:       true,
			errContains:   "OAUTH_CLIENT_SECRET must be set",
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
			oauthSecret:   "valid-oauth-secret-16-chars",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config
			cfg := &Config{
				Environment:       tt.environment,
				SessionSecret:     tt.sessionSecret,
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
		os.Setenv("GO_ENV", "production")
		os.Setenv("SESSION_SECRET", "test-secret-that-is-long-enough-for-prod-validation")
		defer func() {
			os.Unsetenv("GO_ENV")
			os.Unsetenv("SESSION_SECRET")
		}()

		cfg := Load()
		err := cfg.ValidateProduction()
		if err != nil {
			t.Errorf("ValidateProduction() with valid secrets returned error: %v", err)
		}
	})

	t.Run("Production mode with default secret fails", func(t *testing.T) {
		os.Setenv("GO_ENV", "production")
		os.Setenv("SESSION_SECRET", "dev-secret-change-in-production")
		defer func() {
			os.Unsetenv("GO_ENV")
			os.Unsetenv("SESSION_SECRET")
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
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("GO_ENV")

			// Set test environment
			if tt.envLogLevel != "" {
				os.Setenv("LOG_LEVEL", tt.envLogLevel)
			}
			if tt.envGoEnv != "" {
				os.Setenv("GO_ENV", tt.envGoEnv)
			}

			// Load config
			cfg := Load()

			// Verify LOG_LEVEL
			if cfg.LogLevel != tt.expectedLogLevel {
				t.Errorf("Expected LogLevel=%q, got %q", tt.expectedLogLevel, cfg.LogLevel)
			}

			// Cleanup
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("GO_ENV")
		})
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Save current environment
	originalEnv := os.Getenv("GO_ENV")
	defer func() {
		if originalEnv != "" {
			os.Setenv("GO_ENV", originalEnv)
		} else {
			os.Unsetenv("GO_ENV")
		}
	}()

	// Clear GO_ENV to ensure defaults
	os.Unsetenv("GO_ENV")
	os.Unsetenv("LOG_LEVEL")

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
