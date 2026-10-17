package config

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestDatabaseConnectionString(t *testing.T) {
	for _, key := range strings.Fields("DATABASE_URL DATABASE_HOST DATABASE_PORT DATABASE_NAME DATABASE_USER DATABASE_PASSWORD DATABASE_SSLMODE") {
		t.Setenv(key, "")
	}
	if got, err := Load().DatabaseConnectionString(); err != nil || got != "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" {
		t.Fatal("development default changed")
	}
	password := " credential-marker@:/?#%&=+'\\word\n$雪 "
	for key, value := range map[string]string{"DATABASE_HOST": "postgres", "DATABASE_NAME": "savvy", "DATABASE_USER": "user", "DATABASE_PASSWORD": password} {
		t.Setenv(key, value)
	}
	for _, tc := range []struct {
		name, key, value string
		wantErr          bool
	}{
		{"separate parameters", "", "", false},
		{"custom port", "DATABASE_PORT", "6543", false},
		{"URL precedence", "DATABASE_URL", "postgres://user:pass@localhost/db?sslmode=disable", false},
		{"legacy DSN", "DATABASE_URL", "host=localhost user=user password=pass dbname=db sslmode=disable", false},
		{"missing host", "DATABASE_HOST", "", true},
		{"missing database", "DATABASE_NAME", "", true},
		{"missing user", "DATABASE_USER", "", true},
		{"missing password", "DATABASE_PASSWORD", "", true},
		{"invalid host", "DATABASE_HOST", "host:5432", true},
		{"invalid port", "DATABASE_PORT", "invalid", true},
		{"port out of range", "DATABASE_PORT", "65536", true},
		{"TLS fallback", "DATABASE_SSLMODE", "prefer", true},
		{"invalid TLS mode", "DATABASE_SSLMODE", "invalid", true},
		{"invalid URL", "DATABASE_URL", "postgres://user:" + password + "@host:invalid/db", true},
		{"invalid CA", "PGSSLROOTCERT", t.TempDir() + "/missing.crt", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.key != "" {
				t.Setenv(tc.key, tc.value)
			}
			if tc.key == "DATABASE_URL" {
				t.Setenv("DATABASE_HOST", "")
				t.Setenv("DATABASE_PORT", "invalid")
			}
			cfg := Load()
			got, err := cfg.DatabaseConnectionString()
			if tc.wantErr {
				if err == nil || got != "" || cfg.Validate() == nil {
					t.Fatal("invalid configuration must fail without returning a connection string")
				}
				if strings.Contains(err.Error(), "credential-marker") {
					t.Fatal("configuration error exposed credentials")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tc.key == "DATABASE_URL" {
				if got != tc.value {
					t.Fatal("DATABASE_URL did not take precedence")
				}
				return
			}
			parsed, err := pgx.ParseConfig(got)
			if err != nil {
				t.Fatal("generated connection string could not be parsed")
			}
			port := uint16(5432)
			if tc.key == "DATABASE_PORT" {
				port = 6543
			}
			if parsed.Host != "postgres" || parsed.Port != port || parsed.Database != "savvy" || parsed.User != "user" || parsed.Password != password {
				t.Fatal("database parameters did not round-trip through the driver")
			}
			if parsed.TLSConfig == nil || parsed.TLSConfig.InsecureSkipVerify || parsed.TLSConfig.ServerName != "postgres" {
				t.Fatal("default TLS must verify the server identity")
			}
			for _, fallback := range parsed.Fallbacks {
				if fallback.TLSConfig == nil {
					t.Fatal("TLS must not fall back to plaintext")
				}
			}
		})
	}
}
