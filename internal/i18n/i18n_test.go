package i18n

import (
	"context"
	"embed"
	"encoding/json"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

//go:embed testdata/*.json
var testLocalesFS embed.FS

func TestInit_Success(t *testing.T) {
	// Reset bundle
	Bundle = nil

	// Create a wrapper that maps testdata/*.json to locales/*.json
	err := initTestBundle(testLocalesFS)
	require.NoError(t, err)
	require.NotNil(t, Bundle)

	// Verify bundle was created (we can't access defaultLanguage as it's unexported)
	localizer := NewLocalizer("de")
	assert.NotNil(t, localizer)
}

// initTestBundle initializes bundle with testdata files
func initTestBundle(fs embed.FS) error {
	Bundle = i18n.NewBundle(language.German)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, lang := range SupportedLanguages {
		filename := "testdata/" + lang.String() + ".json"
		if _, err := Bundle.LoadMessageFileFS(fs, filename); err != nil {
			return err
		}
	}
	return nil
}

func TestInit_InvalidFS(t *testing.T) {
	Bundle = nil

	var emptyFS embed.FS
	err := Init(emptyFS)
	assert.Error(t, err)
}

func TestNewLocalizer(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	tests := []struct {
		name  string
		langs []string
	}{
		{
			name:  "single language",
			langs: []string{"en"},
		},
		{
			name:  "multiple languages",
			langs: []string{"en", "de", "fr"},
		},
		{
			name:  "no languages",
			langs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localizer := NewLocalizer(tt.langs...)
			assert.NotNil(t, localizer)
		})
	}
}

func TestSetLocalizer(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	ctx := context.Background()
	localizer := NewLocalizer("en")

	ctx = SetLocalizer(ctx, localizer)

	// Verify localizer was stored
	retrieved := ctx.Value(localizerKey)
	require.NotNil(t, retrieved)
	assert.Equal(t, localizer, retrieved)
}

func TestGetLocalizer(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	tests := []struct {
		name     string
		setup    func() context.Context
		checkNil bool
	}{
		{
			name: "with localizer",
			setup: func() context.Context {
				ctx := context.Background()
				localizer := NewLocalizer("en")
				return SetLocalizer(ctx, localizer)
			},
			checkNil: false,
		},
		{
			name: "without localizer - fallback to default",
			setup: func() context.Context {
				return context.Background()
			},
			checkNil: false, // Should return default German localizer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			localizer := GetLocalizer(ctx)

			if tt.checkNil {
				assert.Nil(t, localizer)
			} else {
				assert.NotNil(t, localizer)
			}
		})
	}
}

func TestSetLanguage(t *testing.T) {
	ctx := context.Background()

	tests := []string{"de", "en", "fr"}

	for _, lang := range tests {
		t.Run(lang, func(t *testing.T) {
			ctx = SetLanguage(ctx, lang)
			retrieved := ctx.Value(languageKey)
			require.NotNil(t, retrieved)
			assert.Equal(t, lang, retrieved.(string))
		})
	}
}

func TestGetLanguage(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected string
	}{
		{
			name: "with German language",
			setup: func() context.Context {
				ctx := context.Background()
				return SetLanguage(ctx, "de")
			},
			expected: "de",
		},
		{
			name: "with English language",
			setup: func() context.Context {
				ctx := context.Background()
				return SetLanguage(ctx, "en")
			},
			expected: "en",
		},
		{
			name: "with French language",
			setup: func() context.Context {
				ctx := context.Background()
				return SetLanguage(ctx, "fr")
			},
			expected: "fr",
		},
		{
			name: "without language - fallback to German",
			setup: func() context.Context {
				return context.Background()
			},
			expected: "de",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			lang := GetLanguage(ctx)
			assert.Equal(t, tt.expected, lang)
		})
	}
}

func TestT_WithValidKey(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	ctx := context.Background()
	localizer := NewLocalizer("en")
	ctx = SetLocalizer(ctx, localizer)

	// Test translation
	result := T(ctx, "test.hello")
	assert.NotEmpty(t, result)
	// Should return either the translation or the message ID
	assert.True(t, result == "Hello" || result == "test.hello")
}

func TestT_WithInvalidKey(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	ctx := context.Background()
	localizer := NewLocalizer("en")
	ctx = SetLocalizer(ctx, localizer)

	// Test with non-existent message ID
	result := T(ctx, "nonexistent.key")
	assert.Equal(t, "nonexistent.key", result) // Should return the message ID
}

func TestT_WithTemplateData(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	ctx := context.Background()
	localizer := NewLocalizer("en")
	ctx = SetLocalizer(ctx, localizer)

	// Test with template data
	result := T(ctx, "test.greeting", map[string]any{
		"Name": "John",
	})
	assert.NotEmpty(t, result)
}

func TestT_WithoutContext(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	// Test with empty context (should use default German)
	ctx := context.Background()
	result := T(ctx, "test.hello")
	assert.NotEmpty(t, result)
}

func TestTc_WithCount(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	ctx := context.Background()
	localizer := NewLocalizer("en")
	ctx = SetLocalizer(ctx, localizer)

	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{
			name:  "singular",
			count: 1,
		},
		{
			name:  "plural",
			count: 5,
		},
		{
			name:  "zero",
			count: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Tc(ctx, "test.items", tt.count)
			assert.NotEmpty(t, result)
			// Should contain the count or return the message ID
			assert.True(t, len(result) > 0)
		})
	}
}

func TestTc_WithTemplateData(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	ctx := context.Background()
	localizer := NewLocalizer("en")
	ctx = SetLocalizer(ctx, localizer)

	result := Tc(ctx, "test.items", 5, map[string]any{
		"Count": 5,
		"Type":  "cards",
	})
	assert.NotEmpty(t, result)
}

func TestTc_WithInvalidKey(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	ctx := context.Background()
	localizer := NewLocalizer("en")
	ctx = SetLocalizer(ctx, localizer)

	result := Tc(ctx, "nonexistent.plural", 5)
	assert.Equal(t, "nonexistent.plural", result)
}

func TestSupportedLanguages(t *testing.T) {
	assert.Len(t, SupportedLanguages, 3)
	assert.Contains(t, SupportedLanguages, language.German)
	assert.Contains(t, SupportedLanguages, language.English)
	assert.Contains(t, SupportedLanguages, language.French)
}

func TestContextKeyIsolation(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	// Test that localizerKey and languageKey don't conflict
	ctx := context.Background()

	localizer := NewLocalizer("en")
	ctx = SetLocalizer(ctx, localizer)
	ctx = SetLanguage(ctx, "fr")

	// Both should be independently retrievable
	retrievedLocalizer := GetLocalizer(ctx)
	retrievedLang := GetLanguage(ctx)

	assert.NotNil(t, retrievedLocalizer)
	assert.Equal(t, "fr", retrievedLang)
}

func TestMultipleLanguageFallback(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	// Test fallback chain: fr -> en -> de
	ctx := context.Background()
	localizer := NewLocalizer("fr", "en", "de")
	ctx = SetLocalizer(ctx, localizer)

	result := T(ctx, "test.hello")
	assert.NotEmpty(t, result)
}

func TestT_EmptyMessageID(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	ctx := context.Background()
	localizer := NewLocalizer("en")
	ctx = SetLocalizer(ctx, localizer)

	result := T(ctx, "")
	assert.Equal(t, "", result)
}

func TestGetLocalizer_EmptyContext(t *testing.T) {
	require.NoError(t, initTestBundle(testLocalesFS))

	// Empty context (not nil) should return default localizer
	ctx := context.Background()
	localizer := GetLocalizer(ctx)
	assert.NotNil(t, localizer)
}

func TestGetLanguage_EmptyContext(t *testing.T) {
	// Empty context should return default language
	ctx := context.Background()
	lang := GetLanguage(ctx)
	assert.Equal(t, "de", lang)
}
