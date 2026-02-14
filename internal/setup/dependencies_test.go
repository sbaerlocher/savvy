package setup

import (
	"log/slog"
	"testing"
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
