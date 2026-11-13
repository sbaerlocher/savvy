package testutil

import (
	"testing"
)

func TestDatabaseURLRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := DatabaseURL(); err == nil || err.Error() != "DATABASE_URL not set" {
		t.Fatalf("DatabaseURL() error = %v, want DATABASE_URL not set", err)
	}
}
