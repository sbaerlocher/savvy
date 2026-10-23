package config

import (
	"cmp"
	"errors"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (c *Config) hasDatabaseParameters() bool {
	return c.DatabaseHost != "" || c.DatabasePort != "" || c.DatabaseName != "" ||
		c.DatabaseUser != "" || c.DatabasePassword != "" || c.DatabaseSSLMode != ""
}

// IgnoredDatabaseParameters lists the non-blank separate parameters that DATABASE_URL overrides.
func (c *Config) IgnoredDatabaseParameters() []string {
	if c.DatabaseURL == "" {
		return nil
	}
	var keys []string
	for _, p := range []struct{ key, value string }{
		{"DATABASE_HOST", c.DatabaseHost}, {"DATABASE_PORT", c.DatabasePort},
		{"DATABASE_NAME", c.DatabaseName}, {"DATABASE_USER", c.DatabaseUser},
		{"DATABASE_PASSWORD", c.DatabasePassword}, {"DATABASE_SSLMODE", c.DatabaseSSLMode},
	} {
		if strings.TrimSpace(p.value) != "" {
			keys = append(keys, p.key)
		}
	}
	return keys
}

// DatabaseConnectionString preserves DATABASE_URL precedence and escapes separate parameters.
func (c *Config) DatabaseConnectionString() (string, error) {
	connectionString := c.DatabaseURL
	if connectionString == "" && c.hasDatabaseParameters() {
		if strings.TrimSpace(c.DatabaseHost) == "" || strings.TrimSpace(c.DatabaseName) == "" ||
			strings.TrimSpace(c.DatabaseUser) == "" || c.DatabasePassword == "" {
			return "", errors.New("DATABASE_HOST, DATABASE_NAME, DATABASE_USER and DATABASE_PASSWORD must be set")
		}
		if net.ParseIP(c.DatabaseHost) == nil && strings.ContainsAny(c.DatabaseHost, "/\\:@?#[], \t\r\n") {
			return "", errors.New("DATABASE_HOST must be a hostname or IP address without a port")
		}
		port := cmp.Or(c.DatabasePort, "5432")
		if n, err := strconv.ParseUint(port, 10, 16); err != nil || n == 0 {
			return "", errors.New("DATABASE_PORT must be a valid port number (1-65535)")
		}
		sslMode := cmp.Or(c.DatabaseSSLMode, "verify-full")
		if !slices.Contains([]string{"disable", "require", "verify-ca", "verify-full"}, sslMode) {
			return "", errors.New("DATABASE_SSLMODE must be one of: disable, require, verify-ca, verify-full")
		}
		u := url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(c.DatabaseUser, c.DatabasePassword),
			Host:   net.JoinHostPort(c.DatabaseHost, port),
			// pgx trims leading slashes from URL paths; a query parameter preserves the name.
			RawQuery: url.Values{"dbname": {c.DatabaseName}, "sslmode": {sslMode}}.Encode(),
		}
		connectionString = u.String()
	}
	if connectionString != "" {
		if _, err := pgx.ParseConfig(connectionString); err != nil {
			// Driver errors can include the full connection string and credentials.
			return "", errors.New("invalid PostgreSQL configuration; check DATABASE_URL and database/TLS parameters")
		}
	}
	return connectionString, nil
}
