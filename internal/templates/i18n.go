// Package templates contains i18n helper functions for Templ templates.
package templates

import (
	"context"
	"savvy/internal/i18n"
)

// T translates a message ID with optional template data
func T(ctx context.Context, messageID string, templateData ...map[string]any) string {
	return i18n.T(ctx, messageID, templateData...)
}

// Tc translates a message ID with count (for pluralization)
func Tc(ctx context.Context, messageID string, count int, templateData ...map[string]any) string {
	return i18n.Tc(ctx, messageID, count, templateData...)
}
