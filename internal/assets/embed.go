// Package assets provides embedded locales for the application
package assets

import "embed"

// Locales embeds the locale JSON files (translations)
//
//go:embed all:locales
var Locales embed.FS
