// Package i18n provides internationalization support using go-i18n v2.
package i18n

import (
	"context"
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Note: Translation files are loaded from ../../locales/ directory
// We don't use embed here to allow runtime changes during development

// Bundle is the global i18n bundle
var Bundle *i18n.Bundle

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const localizerKey contextKey = "localizer"

// SupportedLanguages lists all supported language tags
var SupportedLanguages = []language.Tag{
	language.German,  // de - Default
	language.English, // en
	language.French,  // fr
}

// Init initializes the i18n bundle and loads translation files
func Init() error {
	Bundle = i18n.NewBundle(language.German) // Default language
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load translation files from disk
	for _, lang := range SupportedLanguages {
		filename := "locales/" + lang.String() + ".json"
		if _, err := Bundle.LoadMessageFile(filename); err != nil {
			return err
		}
	}

	return nil
}

// NewLocalizer creates a new localizer for the given language preferences
func NewLocalizer(langs ...string) *i18n.Localizer {
	return i18n.NewLocalizer(Bundle, langs...)
}

// SetLocalizer stores a localizer in the context
func SetLocalizer(ctx context.Context, localizer *i18n.Localizer) context.Context {
	return context.WithValue(ctx, localizerKey, localizer)
}

// GetLocalizer retrieves the localizer from the context
func GetLocalizer(ctx context.Context) *i18n.Localizer {
	if localizer, ok := ctx.Value(localizerKey).(*i18n.Localizer); ok {
		return localizer
	}
	// Fallback to default language
	return NewLocalizer(language.German.String())
}

// T translates a message ID with optional template data
func T(ctx context.Context, messageID string, templateData ...map[string]any) string {
	localizer := GetLocalizer(ctx)

	cfg := &i18n.LocalizeConfig{
		MessageID: messageID,
	}

	if len(templateData) > 0 {
		cfg.TemplateData = templateData[0]
	}

	translation, err := localizer.Localize(cfg)
	if err != nil {
		// Return message ID if translation not found
		return messageID
	}

	return translation
}

// Tc translates a message ID with count (for pluralization)
func Tc(ctx context.Context, messageID string, count int, templateData ...map[string]any) string {
	localizer := GetLocalizer(ctx)

	cfg := &i18n.LocalizeConfig{
		MessageID:   messageID,
		PluralCount: count,
	}

	if len(templateData) > 0 {
		cfg.TemplateData = templateData[0]
	} else {
		cfg.TemplateData = map[string]any{
			"Count": count,
		}
	}

	translation, err := localizer.Localize(cfg)
	if err != nil {
		return messageID
	}

	return translation
}
