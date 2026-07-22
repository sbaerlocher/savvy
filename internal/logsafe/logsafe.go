// Package logsafe provides sanitization helpers for values that originate from
// user input and are written to structured logs. Removing control characters
// (in particular CR and LF) prevents log-injection / log-forging, where an
// attacker embeds newlines to fabricate additional log entries.
package logsafe

import (
	"strings"

	"github.com/google/uuid"
)

// controlReplacer strips the characters that let user input break out of a
// single log line. CR and LF are the forging vectors; the remaining C0 control
// characters and DEL are stripped as defense-in-depth so terminal-viewed logs
// cannot be manipulated with escape sequences.
var controlReplacer = func() *strings.Replacer {
	pairs := make([]string, 0, 66)
	for c := 0; c < 0x20; c++ {
		pairs = append(pairs, string(rune(c)), "")
	}
	pairs = append(pairs, "\x7f", "") // DEL
	return strings.NewReplacer(pairs...)
}()

// String returns s with all control characters (including CR and LF) removed,
// making it safe to embed in a structured log entry. Use it to wrap any
// user-controlled string passed as a slog attribute value.
func String(s string) string {
	// CR and LF are stripped via literal strings.ReplaceAll calls (redundant
	// with controlReplacer) because static analyzers such as CodeQL only
	// recognize replacements of literal newline characters as log-injection
	// sanitizers.
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return controlReplacer.Replace(s)
}

// Error returns err's message sanitized like String. It returns "" for nil so
// it can wrap any error passed as a slog attribute value.
func Error(err error) string {
	if err == nil {
		return ""
	}
	return String(err.Error())
}

// UUID returns id's canonical string form routed through String. UUIDs cannot
// contain control characters, but ones parsed from request input are flagged
// as tainted by static analysis when logged directly.
func UUID(id uuid.UUID) string {
	return String(id.String())
}
