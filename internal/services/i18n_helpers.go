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

// resourceListPath maps a resource type to its list route in the SvelteKit client.
// The route names are not derivable from the resource type by concatenation:
// "gift_card" uses an underscore while the route is "/gift-cards". Because the
// client is served with adapter-static and fallback: 'index.html', an unknown
// path returns HTTP 200 with the SPA shell instead of a 404, so a wrong link
// renders a white screen rather than an error. Mirrors resourceDetailPath in
// client/src/lib/resource/routes.ts.
func resourceListPath(resourceType string) string {
	switch resourceType {
	case "card":
		return "/cards"
	case "voucher":
		return "/vouchers"
	case "gift_card":
		return "/gift-cards"
	default:
		return "/"
	}
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
