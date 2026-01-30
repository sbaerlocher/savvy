// Package assets provides embedded static files and locales for the application
package assets

import "embed"

// Static embeds the static directory (CSS, JS, PWA files)
//
//go:embed all:static
var Static embed.FS

// Locales embeds the locale JSON files (translations)
//
//go:embed all:locales
var Locales embed.FS
