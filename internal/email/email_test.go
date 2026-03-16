package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeHeader(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal@example.com", "normal@example.com"},
		{"evil\r\nBcc: target@evil.com", "evilBcc: target@evil.com"},
		{"evil\nX-Header: injected", "evilX-Header: injected"},
		{"evil\rX-Header: injected", "evilX-Header: injected"},
		{"multi\r\n\r\ninjection", "multiinjection"},
		{"", ""},
		{"no-special-chars", "no-special-chars"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, sanitizeHeader(tt.input))
	}
}
