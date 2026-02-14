// Package email provides SMTP email sending functionality.
package email

import (
	"context"
	"savvy/internal/i18n"
)

// emailCtx creates a context with an i18n localizer for the given language.
// Used by email string functions to translate email content.
// Returns a plain context if the i18n bundle is not initialized (e.g. in tests).
func emailCtx(lang string) context.Context {
	if lang == "" {
		lang = "de"
	}
	ctx := context.Background()
	if i18n.Bundle == nil {
		return ctx
	}
	localizer := i18n.NewLocalizer(lang)
	return i18n.SetLocalizer(ctx, localizer)
}

// et is a shorthand for i18n.T within the email package.
func et(ctx context.Context, id string, data ...map[string]any) string {
	return i18n.T(ctx, id, data...)
}

// etc is a shorthand for i18n.Tc within the email package (pluralization).
func etc(ctx context.Context, id string, count int, data ...map[string]any) string {
	return i18n.Tc(ctx, id, count, data...)
}
