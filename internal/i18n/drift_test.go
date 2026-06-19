package i18n_test

import (
	"encoding/json"
	"sort"
	"testing"

	"savvy/internal/assets"
	"savvy/internal/i18n"

	"github.com/stretchr/testify/require"
)

// localeMessage matches one entry in a locale JSON file.
type localeMessage struct {
	ID string `json:"id"`
}

// loadKeys reads the embedded locale file for lang and returns its sorted set of message IDs.
func loadKeys(t *testing.T, lang string) []string {
	t.Helper()
	data, err := assets.Locales.ReadFile("locales/" + lang + ".json")
	require.NoError(t, err, "read locale %s", lang)

	var msgs []localeMessage
	require.NoError(t, json.Unmarshal(data, &msgs), "unmarshal locale %s", lang)

	keys := make([]string, 0, len(msgs))
	for _, m := range msgs {
		keys = append(keys, m.ID)
	}
	sort.Strings(keys)
	return keys
}

// TestLocaleKeysInSync guards against backend i18n drift: every supported language
// must define the exact same set of message IDs. A German-only developer adding a key
// without the EN/FR counterparts otherwise ships empty strings to users. The TS
// frontend locales are already covered by the strict TranslationKeys type check; this
// test closes the equivalent gap for the untyped backend JSON files.
func TestLocaleKeysInSync(t *testing.T) {
	ref := i18n.SupportedLanguages[0].String()
	refKeys := loadKeys(t, ref)
	require.NotEmpty(t, refKeys, "reference locale %s has no keys", ref)

	refSet := make(map[string]struct{}, len(refKeys))
	for _, k := range refKeys {
		refSet[k] = struct{}{}
	}

	for _, tag := range i18n.SupportedLanguages[1:] {
		lang := tag.String()
		keys := loadKeys(t, lang)

		set := make(map[string]struct{}, len(keys))
		for _, k := range keys {
			set[k] = struct{}{}
		}

		var missing, extra []string
		for _, k := range refKeys {
			if _, ok := set[k]; !ok {
				missing = append(missing, k)
			}
		}
		for _, k := range keys {
			if _, ok := refSet[k]; !ok {
				extra = append(extra, k)
			}
		}

		require.Emptyf(t, missing, "locale %s is missing keys present in %s: %v", lang, ref, missing)
		require.Emptyf(t, extra, "locale %s has keys absent in %s: %v", lang, ref, extra)
	}
}
