package setup

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"savvy/internal/config"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{
			name:     "DEBUG uppercase",
			input:    "DEBUG",
			expected: slog.LevelDebug,
		},
		{
			name:     "debug lowercase",
			input:    "debug",
			expected: slog.LevelDebug,
		},
		{
			name:     "INFO uppercase",
			input:    "INFO",
			expected: slog.LevelInfo,
		},
		{
			name:     "info lowercase",
			input:    "info",
			expected: slog.LevelInfo,
		},
		{
			name:     "WARN uppercase",
			input:    "WARN",
			expected: slog.LevelWarn,
		},
		{
			name:     "warn lowercase",
			input:    "warn",
			expected: slog.LevelWarn,
		},
		{
			name:     "WARNING uppercase",
			input:    "WARNING",
			expected: slog.LevelWarn,
		},
		{
			name:     "warning lowercase",
			input:    "warning",
			expected: slog.LevelWarn,
		},
		{
			name:     "ERROR uppercase",
			input:    "ERROR",
			expected: slog.LevelError,
		},
		{
			name:     "error lowercase",
			input:    "error",
			expected: slog.LevelError,
		},
		{
			name:     "invalid level defaults to INFO",
			input:    "INVALID",
			expected: slog.LevelInfo,
		},
		{
			name:     "empty string defaults to INFO",
			input:    "",
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInitOAuth(t *testing.T) {
	tests := []struct {
		name         string
		localLogin   bool
		disabled     bool
		discoveryOK  bool
		wantError    bool
		wantRequests int32
	}{
		{name: "OIDC-only discovery failure", wantError: true, wantRequests: 1},
		{name: "OIDC-only valid discovery", discoveryOK: true, wantRequests: 1},
		{name: "local login fallback", localLogin: true, wantRequests: 1},
		{name: "OAuth disabled", localLogin: true, disabled: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requests atomic.Int32
			issuer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if !tt.discoveryOK {
					http.Error(w, "sensitive-discovery-secret", http.StatusServiceUnavailable)
					return
				}
				url := "http://" + r.Host
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{
					"issuer":                 url,
					"authorization_endpoint": url + "/authorize",
					"token_endpoint":         url + "/token",
					"jwks_uri":               url + "/keys",
				})
			}))
			defer issuer.Close()
			cfg := &config.Config{EnableLocalLogin: tt.localLogin}
			if !tt.disabled {
				cfg.OAuthClientID = "test-client"
				cfg.OAuthClientSecret = "test-client-secret"
				cfg.OAuthIssuer = issuer.URL
			}

			provider, err := InitOAuth(cfg)
			require.Equal(t, tt.wantError, err != nil, "InitOAuth error: %v", err)
			require.Equal(t, tt.discoveryOK, provider != nil, "OAuth provider present")
			require.Equal(t, tt.wantRequests, requests.Load(), "discovery requests")
			if err != nil {
				require.NotContains(t, err.Error(), "sensitive-discovery-secret")
				require.NotContains(t, err.Error(), cfg.OAuthClientSecret)
			}
		})
	}
}

func TestInitAllDependenciesOIDCOnlyDiscoveryFailure(t *testing.T) {
	issuer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "discovery unavailable", http.StatusServiceUnavailable)
	}))
	defer issuer.Close()
	previousLogger := slog.Default()
	defer slog.SetDefault(previousLogger)

	shutdown, provider, err := InitAllDependencies(&config.Config{
		LogLevel:          "ERROR",
		OAuthClientID:     "test-client",
		OAuthClientSecret: "test-client-secret",
		OAuthIssuer:       issuer.URL,
		DatabaseURL:       "invalid-database-url",
	})
	require.NotNil(t, shutdown)
	defer Shutdown(shutdown, nil)
	require.ErrorContains(t, err, "local login disabled")
	require.Nil(t, provider)
}
