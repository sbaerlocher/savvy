package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"savvy/internal/assets"
	"savvy/internal/i18n"
)

// initI18n loads the real translation bundle so the push-body helpers render
// actual strings instead of returning the message key.
func initI18n(t *testing.T) {
	t.Helper()
	require.NoError(t, i18n.Init(assets.Locales))
	t.Cleanup(func() { i18n.Bundle = nil })
}

// TestSharePushBody_ContainsMerchant_NeverValue is the datenschutz invariant:
// the push body names the merchant (and description) but must never contain the
// monetary value — amounts must not appear on a lockscreen.
func TestSharePushBody_ContainsMerchant_NeverValue(t *testing.T) {
	initI18n(t)

	for _, lang := range []string{"de", "en", "fr"} {
		body := sharePushBody(lang, "Alice", "voucher", "IKEA", "")
		assert.Contains(t, body, "IKEA", "%s: merchant must appear in share push body", lang)
		// A value like "50" / "CHF 50.00" must never leak — the helper has no
		// amount parameter, so no amount digits can appear.
		assert.NotContains(t, body, "50", "%s: value must never appear in push body", lang)
		assert.NotContains(t, body, "CHF", "%s: currency must never appear in push body", lang)
	}
}

// TestTransferPushBody_ContainsMerchant_NeverValue mirrors the share case.
func TestTransferPushBody_ContainsMerchant_NeverValue(t *testing.T) {
	initI18n(t)

	for _, lang := range []string{"de", "en", "fr"} {
		body := transferPushBody(lang, "Alice", "gift_card", "IKEA", "")
		assert.Contains(t, body, "IKEA", "%s: merchant must appear in transfer push body", lang)
		assert.NotContains(t, body, "CHF", "%s: currency must never appear in push body", lang)
	}
}

// TestPushBody_WithDescription includes the description when present.
func TestSharePushBody_WithDescription(t *testing.T) {
	initI18n(t)

	body := sharePushBody("en", "Alice", "voucher", "IKEA", "20% off")
	assert.Contains(t, body, "IKEA")
	assert.Contains(t, body, "20% off")
}

// TestPushBody_EmptyMerchantFallsBackToGeneric uses the generic wording (no
// merchant mention) when no merchant is known.
func TestSharePushBody_EmptyMerchantFallsBackToGeneric(t *testing.T) {
	initI18n(t)

	generic := sharePushBody("en", "Alice", "voucher", "", "")
	withMerchant := sharePushBody("en", "Alice", "voucher", "IKEA", "")
	assert.NotContains(t, generic, "IKEA")
	assert.NotEqual(t, generic, withMerchant)
	// generic must still name the sharer
	assert.True(t, strings.Contains(generic, "Alice"))
}
