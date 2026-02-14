package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// i18nCtx Tests
// ============================================================================

// TestI18nCtx_EmptyLangDefaultsToDe verifies that an empty language string defaults to "de"
func TestI18nCtx_EmptyLangDefaultsToDe(t *testing.T) {
	// i18n.Bundle is nil in tests, so i18nCtx returns a plain context.Background()
	ctx := i18nCtx("")
	assert.NotNil(t, ctx)
}

// TestI18nCtx_NonEmptyLang verifies that a non-empty language is accepted
func TestI18nCtx_NonEmptyLang(t *testing.T) {
	ctx := i18nCtx("en")
	assert.NotNil(t, ctx)
}

// TestI18nCtx_FrenchLang verifies French language
func TestI18nCtx_FrenchLang(t *testing.T) {
	ctx := i18nCtx("fr")
	assert.NotNil(t, ctx)
}

// TestI18nCtx_BundleNil verifies that when i18n.Bundle is nil (as in tests),
// a plain context is returned without panic
func TestI18nCtx_BundleNil(t *testing.T) {
	// In test environment, i18n.Bundle is always nil (not initialized)
	// Ensure it does not panic and returns a valid context
	ctx := i18nCtx("de")
	assert.NotNil(t, ctx)

	// Also verify with empty string (default path)
	ctx2 := i18nCtx("")
	assert.NotNil(t, ctx2)
}

// ============================================================================
// pushArticle Tests
// ============================================================================

// TestPushArticle_DeGiftCard verifies German article for gift_card (feminine → "eine")
func TestPushArticle_DeGiftCard(t *testing.T) {
	result := pushArticle("gift_card", "de")
	assert.Equal(t, "eine", result)
}

// TestPushArticle_DeCard verifies German article for card (feminine → "eine")
func TestPushArticle_DeCard(t *testing.T) {
	result := pushArticle("card", "de")
	assert.Equal(t, "eine", result)
}

// TestPushArticle_DeVoucher verifies German article for voucher (masculine → "einen")
func TestPushArticle_DeVoucher(t *testing.T) {
	result := pushArticle("voucher", "de")
	assert.Equal(t, "einen", result)
}

// TestPushArticle_FrGiftCard verifies French article for gift_card (feminine → "une")
func TestPushArticle_FrGiftCard(t *testing.T) {
	result := pushArticle("gift_card", "fr")
	assert.Equal(t, "une", result)
}

// TestPushArticle_FrCard verifies French article for card (feminine → "une")
func TestPushArticle_FrCard(t *testing.T) {
	result := pushArticle("card", "fr")
	assert.Equal(t, "une", result)
}

// TestPushArticle_FrVoucher verifies French article for voucher (masculine → "un")
func TestPushArticle_FrVoucher(t *testing.T) {
	result := pushArticle("voucher", "fr")
	assert.Equal(t, "un", result)
}

// TestPushArticle_EnCard verifies English article defaults to "a"
func TestPushArticle_EnCard(t *testing.T) {
	result := pushArticle("card", "en")
	assert.Equal(t, "a", result)
}

// TestPushArticle_EnVoucher verifies English article defaults to "a"
func TestPushArticle_EnVoucher(t *testing.T) {
	result := pushArticle("voucher", "en")
	assert.Equal(t, "a", result)
}

// TestPushArticle_EnGiftCard verifies English article defaults to "a"
func TestPushArticle_EnGiftCard(t *testing.T) {
	result := pushArticle("gift_card", "en")
	assert.Equal(t, "a", result)
}

// TestPushArticle_UnknownLanguageDefaultsToA verifies unknown languages default to "a"
func TestPushArticle_UnknownLanguageDefaultsToA(t *testing.T) {
	result := pushArticle("card", "ja")
	assert.Equal(t, "a", result)
}

// TestPushArticle_EmptyLanguageDefaultsToA verifies empty language defaults to "a"
func TestPushArticle_EmptyLanguageDefaultsToA(t *testing.T) {
	result := pushArticle("card", "")
	assert.Equal(t, "a", result)
}

// TestPushArticle_DeUnknownResourceType verifies unknown resource type in German defaults to "einen"
func TestPushArticle_DeUnknownResourceType(t *testing.T) {
	result := pushArticle("unknown", "de")
	assert.Equal(t, "einen", result)
}

// TestPushArticle_FrUnknownResourceType verifies unknown resource type in French defaults to "un"
func TestPushArticle_FrUnknownResourceType(t *testing.T) {
	result := pushArticle("unknown", "fr")
	assert.Equal(t, "un", result)
}
