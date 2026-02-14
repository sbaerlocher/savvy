package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestParseAcceptLanguage_ValidHeader(t *testing.T) {
	tests := []struct {
		name         string
		acceptLang   string
		expectedLang string
	}{
		{
			name:         "German with quality",
			acceptLang:   "de-CH,de;q=0.9,en;q=0.8",
			expectedLang: "de",
		},
		{
			name:         "English preference",
			acceptLang:   "en-US,en;q=0.9",
			expectedLang: "en",
		},
		{
			name:         "French primary",
			acceptLang:   "fr-FR,fr;q=0.9,en;q=0.8",
			expectedLang: "fr",
		},
		{
			name:         "Multiple languages",
			acceptLang:   "en-GB,en;q=0.9,de;q=0.8,fr;q=0.7",
			expectedLang: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAcceptLanguage(tt.acceptLang)
			assert.Equal(t, tt.expectedLang, result)
		})
	}
}

func TestParseAcceptLanguage_EmptyHeader(t *testing.T) {
	result := parseAcceptLanguage("")
	assert.Empty(t, result)
}

func TestParseAcceptLanguage_InvalidHeader(t *testing.T) {
	// Should not panic with invalid header
	assert.NotPanics(t, func() {
		result := parseAcceptLanguage("invalid-header-value")
		_ = result // May be empty or fallback
	})
}

func TestIsLanguageSupported(t *testing.T) {
	tests := []struct {
		lang     string
		expected bool
	}{
		{"de", true},
		{"en", true},
		{"fr", true},
		{"es", false},
		{"zh", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := isLanguageSupported(tt.lang)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetLanguage_ValidLanguage(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/set-language?lang=en", nil)
	req.Header.Set("Referer", "/cards")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SetLanguage(c)
	assert.NoError(t, err)

	// Should redirect to the relative path
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))

	// Check cookie is set
	cookies := rec.Result().Cookies()
	var langCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == LanguageCookieName {
			langCookie = cookie
			break
		}
	}

	assert.NotNil(t, langCookie)
	assert.Equal(t, "en", langCookie.Value)
}

func TestSetLanguage_InvalidLanguage(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/set-language?lang=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SetLanguage(c)
	assert.NoError(t, err)

	// Should set default language (German)
	cookies := rec.Result().Cookies()
	var langCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == LanguageCookieName {
			langCookie = cookie
			break
		}
	}

	assert.NotNil(t, langCookie)
	assert.Equal(t, "de", langCookie.Value)
}

func TestSetLanguage_NoReferer(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/set-language?lang=fr", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SetLanguage(c)
	assert.NoError(t, err)

	// Should redirect to home
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/", rec.Header().Get("Location"))
}

func TestSetLanguage_RefererWithQueryParams(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/set-language?lang=en", nil)
	req.Header.Set("Referer", "/cards?page=2&sort=name")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SetLanguage(c)
	assert.NoError(t, err)

	// Should redirect without query params
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/cards", rec.Header().Get("Location"))
}

func TestSetLanguage_OpenRedirectPrevention(t *testing.T) {
	tests := []struct {
		name    string
		referer string
	}{
		{"absolute URL", "https://evil.com/phishing"},
		{"http URL", "http://evil.com/steal"},
		{"protocol-relative URL", "//evil.com/bypass"},
		{"protocol-relative with path", "//evil.com/foo/bar"},
		{"javascript URI", "javascript:alert(1)"},
		{"data URI", "data:text/html,<h1>evil</h1>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestAuth()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/set-language?lang=en", nil)
			req.Header.Set("Referer", tt.referer)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := SetLanguage(c)
			assert.NoError(t, err)

			// All external/malicious referers must redirect to home
			assert.Equal(t, http.StatusSeeOther, rec.Code)
			assert.Equal(t, "/", rec.Header().Get("Location"))
		})
	}
}

func TestSetLanguage_AllSupportedLanguages(t *testing.T) {
	languages := []string{"de", "en", "fr"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			setupTestAuth()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/set-language?lang="+lang, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := SetLanguage(c)
			assert.NoError(t, err)

			cookies := rec.Result().Cookies()
			var langCookie *http.Cookie
			for _, cookie := range cookies {
				if cookie.Name == LanguageCookieName {
					langCookie = cookie
					break
				}
			}

			assert.NotNil(t, langCookie)
			assert.Equal(t, lang, langCookie.Value)
		})
	}
}

func TestLanguageCookieConstants(t *testing.T) {
	assert.Equal(t, "lang", LanguageCookieName)
	assert.Equal(t, 365*24*60*60, LanguageCookieMaxAge)
}

func TestLanguageDetection_Integration(t *testing.T) {
	// Note: We cannot fully test LanguageDetection middleware because it requires i18n.Bundle to be initialized
	// This would require embedding translation files in the test package
	// Instead, we test the helper functions which provide good coverage
	// The middleware itself is covered by integration tests in the main application

	// Test that parseAcceptLanguage works correctly
	lang := parseAcceptLanguage("de-CH,de;q=0.9")
	assert.Equal(t, "de", lang)

	// Test that isLanguageSupported works correctly
	assert.True(t, isLanguageSupported("de"))
	assert.True(t, isLanguageSupported("en"))
	assert.True(t, isLanguageSupported("fr"))
	assert.False(t, isLanguageSupported("es"))
}

func TestSetLanguage_CookieProperties(t *testing.T) {
	setupTestAuth()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/set-language?lang=en", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SetLanguage(c)
	assert.NoError(t, err)

	cookies := rec.Result().Cookies()
	var langCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == LanguageCookieName {
			langCookie = cookie
			break
		}
	}

	assert.NotNil(t, langCookie)
	assert.Equal(t, "/", langCookie.Path)
	assert.Equal(t, LanguageCookieMaxAge, langCookie.MaxAge)
	assert.False(t, langCookie.HttpOnly) // Allow JavaScript to read for language switcher
	assert.Equal(t, http.SameSiteLaxMode, langCookie.SameSite)
}
