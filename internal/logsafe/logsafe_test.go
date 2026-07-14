package logsafe

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello world", "hello world"},
		{"newline forging", "user\ninjected FATAL log line", "userinjected FATAL log line"},
		{"carriage return", "a\r\nb", "ab"},
		{"tab stripped", "a\tb", "ab"},
		{"del stripped", "a\x7fb", "ab"},
		{"other control chars", "a\x00\x1bb", "ab"},
		{"unicode preserved", "Zürich café 日本語", "Zürich café 日本語"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.in); got != tt.want {
				t.Errorf("String(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
