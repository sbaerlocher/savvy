package testutil

import (
	"strings"
	"testing"
)

func TestInitSharedDBRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	once.Do(initSharedDB)
	if initErr == nil || !strings.Contains(initErr.Error(), "DATABASE_URL not set") {
		t.Fatalf("initErr = %v, want it to mention DATABASE_URL not set", initErr)
	}
	if sharedDB != nil {
		t.Fatal("sharedDB opened without DATABASE_URL")
	}
}
