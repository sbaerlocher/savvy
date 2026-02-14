// Package services contains business logic.
package services

import (
	"context"
	"savvy/internal/i18n"
)

// i18nCtx creates a context with an i18n localizer for the given language.
// Use this in background services (reminders, notifications) where no HTTP
// request context with a localizer is available.
// Returns a plain context if the i18n bundle is not initialized (e.g. in tests).
func i18nCtx(lang string) context.Context {
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

// pushArticle returns the grammatical article for a resource type in the given language.
// Used in share/transfer push notification templates (e.g. "einen Gutschein", "une carte").
func pushArticle(resourceType, lang string) string {
	switch lang {
	case "de":
		// Accusative: Geschenkkarte/Karte = feminine → "eine", Gutschein = masculine → "einen"
		if resourceType == "gift_card" || resourceType == "card" {
			return "eine"
		}
		return "einen"
	case "fr":
		// Carte cadeau/carte = feminine → "une", bon = masculine → "un"
		if resourceType == "gift_card" || resourceType == "card" {
			return "une"
		}
		return "un"
	default:
		return "a"
	}
}
