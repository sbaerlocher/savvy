package email

import (
	"testing"

	"savvy/internal/assets"
	"savvy/internal/i18n"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initTestI18n(t *testing.T) {
	t.Helper()
	i18n.Bundle = nil
	require.NoError(t, i18n.Init(assets.Locales))
}

func TestPasswordResetStrings(t *testing.T) {
	initTestI18n(t)

	tests := []struct {
		lang            string
		wantSubject     string
		wantGreeting    string
		wantButtonText  string
		wantExpiresText string
	}{
		{"de", "Passwort zurücksetzen - Savvy", "Hallo Alice,", "Passwort zurücksetzen", "Dieser Link läuft in 1 Stunde ab."},
		{"en", "Reset Your Password - Savvy", "Hi Alice,", "Reset Password", "This link will expire in 1 hour."},
		{"fr", "Réinitialisation du mot de passe - Savvy", "Bonjour Alice,", "Réinitialiser le mot de passe", "Ce lien expirera dans 1 heure."},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			strs := passwordResetStrings(tt.lang, "Alice", "https://example.com/reset", "1 Stunde")
			assert.Equal(t, tt.wantSubject, strs.Subject)
			assert.Equal(t, tt.wantGreeting, strs.Data["Greeting"])
			assert.Equal(t, tt.wantButtonText, strs.Data["ButtonText"])
			assert.Equal(t, tt.lang, strs.Data["Lang"])
			assert.Equal(t, "https://example.com/reset", strs.Data["ResetURL"])
			assert.NotEmpty(t, strs.Data["Footer"])
			assert.NotEmpty(t, strs.Data["FallbackText"])
		})
	}
}

func TestEmailVerificationStrings(t *testing.T) {
	initTestI18n(t)

	tests := []struct {
		lang           string
		wantSubject    string
		wantButtonText string
	}{
		{"de", "E-Mail bestätigen - Savvy", "E-Mail bestätigen"},
		{"en", "Verify Your Email - Savvy", "Verify Email"},
		{"fr", "Vérification de l'e-mail - Savvy", "Vérifier l'e-mail"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			strs := emailVerificationStrings(tt.lang, "Bob", "https://example.com/verify")
			assert.Equal(t, tt.wantSubject, strs.Subject)
			assert.Equal(t, tt.wantButtonText, strs.Data["ButtonText"])
			assert.NotEmpty(t, strs.Data["ExpiresText"])
		})
	}
}

func TestAccountDeletedStrings(t *testing.T) {
	initTestI18n(t)

	tests := []struct {
		lang        string
		wantSubject string
		wantTitle   string
	}{
		{"de", "Konto gelöscht - Savvy", "Konto gelöscht"},
		{"en", "Account Deleted - Savvy", "Account Deleted"},
		{"fr", "Compte supprimé - Savvy", "Compte supprimé"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			strs := accountDeletedStrings(tt.lang, "Charlie")
			assert.Equal(t, tt.wantSubject, strs.Subject)
			assert.Equal(t, tt.wantTitle, strs.Data["Title"])
			assert.NotEmpty(t, strs.Data["Message"])
			assert.NotEmpty(t, strs.Data["IgnoreText"])
		})
	}
}

func TestTestEmailStrings(t *testing.T) {
	initTestI18n(t)

	tests := []struct {
		lang        string
		wantSubject string
	}{
		{"de", "SMTP Test erfolgreich - Savvy"},
		{"en", "SMTP Test Successful - Savvy"},
		{"fr", "Test SMTP réussi - Savvy"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			strs := testEmailStrings(tt.lang, "Admin")
			assert.Equal(t, tt.wantSubject, strs.Subject)
			assert.NotEmpty(t, strs.Data["StatusTitle"])
			assert.NotEmpty(t, strs.Data["StatusLine1"])
			assert.NotEmpty(t, strs.Data["StatusLine2"])
			assert.NotEmpty(t, strs.Data["StatusLine3"])
			assert.NotEmpty(t, strs.Data["SuccessMessage"])
			assert.NotEmpty(t, strs.Data["AutomatedMessage"])
		})
	}
}

func TestExpiryReminderStrings(t *testing.T) {
	initTestI18n(t)

	data := ExpiryReminderData{
		MerchantName: "IKEA",
		ResourceType: "voucher",
		DaysLeft:     3,
		ExpiresAt:    "26. Februar 2026",
		Code:         "ABC123",
		Value:        "CHF 50.00",
		ResourceURL:  "https://example.com/vouchers/123",
	}

	tests := []struct {
		lang         string
		wantDaysText string
		wantTitle    string
	}{
		{"de", "3 Tagen", "Ablauferinnerung"},
		{"en", "3 days", "Expiry Reminder"},
		{"fr", "3 jours", "Rappel d'expiration"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			strs := expiryReminderStrings(tt.lang, "User", "https://example.com/unsub", data)
			assert.Contains(t, strs.Subject, "IKEA")
			assert.Equal(t, tt.wantTitle, strs.Data["Title"])
			assert.Equal(t, tt.wantDaysText, strs.Data["DaysLeftText"])
			assert.Contains(t, strs.Data["Message"], "IKEA")
			assert.NotEmpty(t, strs.Data["CodeLabel"])
			assert.NotEmpty(t, strs.Data["ValueLabel"])
		})
	}
}

func TestExpiryReminderStrings_Pluralization(t *testing.T) {
	initTestI18n(t)

	data := ExpiryReminderData{
		MerchantName: "Test",
		ResourceType: "voucher",
		Code:         "X",
		Value:        "10%",
		ResourceURL:  "https://example.com",
	}

	tests := []struct {
		days     int
		lang     string
		wantText string
	}{
		{0, "de", "heute"},
		{1, "de", "1 Tag"},
		{7, "de", "7 Tagen"},
		{0, "en", "today"},
		{1, "en", "1 day"},
		{7, "en", "7 days"},
		{0, "fr", "aujourd'hui"},
		{1, "fr", "1 jour"},
		{7, "fr", "7 jours"},
	}

	for _, tt := range tests {
		t.Run(tt.lang+"_"+tt.wantText, func(t *testing.T) {
			data.DaysLeft = tt.days
			strs := expiryReminderStrings(tt.lang, "User", "", data)
			assert.Equal(t, tt.wantText, strs.Data["DaysLeftText"])
		})
	}
}

func TestExpiryReminderStrings_GiftCardLabels(t *testing.T) {
	initTestI18n(t)

	data := ExpiryReminderData{
		MerchantName: "Test",
		ResourceType: "gift_card",
		DaysLeft:     3,
		Code:         "1234",
		Value:        "CHF 100.00",
		ResourceURL:  "https://example.com",
	}

	tests := []struct {
		lang          string
		wantCodeLabel string
	}{
		{"de", "Kartennummer"},
		{"en", "Card Number"},
		{"fr", "Numéro de carte"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			strs := expiryReminderStrings(tt.lang, "User", "", data)
			assert.Equal(t, tt.wantCodeLabel, strs.Data["CodeLabel"])
		})
	}
}

func TestShareNotificationStrings(t *testing.T) {
	initTestI18n(t)

	tests := []struct {
		lang        string
		wantTitle   string
		wantSubject string
	}{
		{"de", "Neue Freigabe", "Alice hat etwas mit dir geteilt - Savvy"},
		{"en", "New Share", "Alice shared something with you - Savvy"},
		{"fr", "Nouveau partage", "Alice a partagé quelque chose avec vous - Savvy"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			strs := shareNotificationStrings(tt.lang, "Bob", "Alice", "voucher", "IKEA", "20% off", 50, "CHF", "https://example.com/v/1", "https://example.com/unsub")
			assert.Equal(t, tt.wantSubject, strs.Subject)
			assert.Equal(t, tt.wantTitle, strs.Data["Title"])
			assert.Contains(t, strs.Data["Message"], "Alice")
			assert.NotEmpty(t, strs.Data["UnsubscribeText"])
			// merchant + description + value are carried through
			assert.Equal(t, "IKEA", strs.Data["Merchant"])
			assert.Equal(t, "20% off", strs.Data["Description"])
			assert.Equal(t, "CHF 50.00", strs.Data["Value"])
			assert.NotEmpty(t, strs.Data["MerchantLabel"])
		})
	}
}

// TestShareNotificationStrings_NoValue verifies an empty value yields no value
// string so the template omits the row.
func TestShareNotificationStrings_NoValue(t *testing.T) {
	initTestI18n(t)
	strs := shareNotificationStrings("en", "Bob", "Alice", "card", "", "Some notes", 0, "", "https://example.com/c/1", "")
	assert.Equal(t, "", strs.Data["Merchant"])
	assert.Equal(t, "Some notes", strs.Data["Description"])
	assert.Equal(t, "", strs.Data["Value"])
}

func TestTransferNotificationStrings(t *testing.T) {
	initTestI18n(t)

	tests := []struct {
		lang        string
		wantTitle   string
		wantSubject string
	}{
		{"de", "Neue Übertragung", "Alice hat dir etwas übertragen - Savvy"},
		{"en", "New Transfer", "Alice transferred something to you - Savvy"},
		{"fr", "Nouveau transfert", "Alice vous a transféré quelque chose - Savvy"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			strs := transferNotificationStrings(tt.lang, "Bob", "Alice", "card", "IKEA", "", 0, "", "https://example.com/c/1", "https://example.com/unsub")
			assert.Equal(t, tt.wantSubject, strs.Subject)
			assert.Equal(t, tt.wantTitle, strs.Data["Title"])
			assert.Contains(t, strs.Data["Message"], "Alice")
			assert.Equal(t, "IKEA", strs.Data["Merchant"])
		})
	}
}

func TestEmailCtx_NilBundle(t *testing.T) {
	i18n.Bundle = nil
	ctx := emailCtx("de")
	// Should not panic, returns plain context
	assert.NotNil(t, ctx)

	// et should return messageID as fallback when bundle is nil
	result := et(ctx, "email.common.greeting", map[string]any{"UserName": "Test"})
	assert.Equal(t, "email.common.greeting", result)
}
