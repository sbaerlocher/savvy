// Package middleware contains Echo middleware for authentication, sessions, and observability.
package middleware

import (
	"net/http"
	"savvy/internal/i18n"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
)

const (
	// LanguageCookieName is the name of the cookie storing the user's language preference
	LanguageCookieName = "lang"
	// LanguageCookieMaxAge is the max age of the language cookie (1 year)
	LanguageCookieMaxAge = 365 * 24 * 60 * 60
)

// LanguageDetection middleware detects the user's preferred language and sets up the localizer
func LanguageDetection(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var lang string

		// 1. Check if language preference is stored in cookie
		if cookie, err := c.Cookie(LanguageCookieName); err == nil && cookie.Value != "" {
			lang = cookie.Value
		}

		// 2. Fallback to Accept-Language header if no cookie
		if lang == "" {
			acceptLang := c.Request().Header.Get("Accept-Language")
			lang = parseAcceptLanguage(acceptLang)
		}

		// 3. Validate language is supported, fallback to default (German)
		if !isLanguageSupported(lang) {
			lang = "de"
		}

		// 4. Create localizer for the detected language
		localizer := i18n.NewLocalizer(lang)

		// 5. Store localizer and language in context
		ctx := i18n.SetLocalizer(c.Request().Context(), localizer)
		ctx = i18n.SetLanguage(ctx, lang)
		c.SetRequest(c.Request().WithContext(ctx))

		// 6. Set language cookie if not already set
		if cookie, err := c.Cookie(LanguageCookieName); err != nil || cookie.Value != lang {
			c.SetCookie(&http.Cookie{
				Name:     LanguageCookieName,
				Value:    lang,
				Path:     "/",
				MaxAge:   LanguageCookieMaxAge,
				HttpOnly: false, // Allow JavaScript to read for language switcher
				Secure:   false, // Set to true in production with HTTPS
				SameSite: http.SameSiteLaxMode,
			})
		}

		return next(c)
	}
}

// parseAcceptLanguage parses the Accept-Language header and returns the best matching language
func parseAcceptLanguage(acceptLang string) string {
	if acceptLang == "" {
		return ""
	}

	// Parse Accept-Language header (e.g., "de-CH,de;q=0.9,en;q=0.8,fr;q=0.7")
	tags, _, err := language.ParseAcceptLanguage(acceptLang)
	if err != nil || len(tags) == 0 {
		return ""
	}

	// Match against supported languages
	matcher := language.NewMatcher(i18n.SupportedLanguages)
	tag, _, _ := matcher.Match(tags...)

	// Return base language (de, en, fr)
	base, _ := tag.Base()
	return base.String()
}

// isLanguageSupported checks if the given language code is supported
func isLanguageSupported(lang string) bool {
	for _, supportedTag := range i18n.SupportedLanguages {
		base, _ := supportedTag.Base()
		if base.String() == lang {
			return true
		}
	}
	return false
}

// SetLanguage sets the user's language preference
func SetLanguage(c echo.Context) error {
	lang := c.QueryParam("lang")

	// Validate language
	if !isLanguageSupported(lang) {
		lang = "de"
	}

	// Set cookie
	c.SetCookie(&http.Cookie{
		Name:     LanguageCookieName,
		Value:    lang,
		Path:     "/",
		MaxAge:   LanguageCookieMaxAge,
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect back to referrer or home
	referer := c.Request().Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}

	// Remove lang query param from referer
	referer = strings.Split(referer, "?")[0]

	return c.Redirect(303, referer)
}
