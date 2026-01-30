// Package config provides configuration management for asset manifest handling.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

var (
	assetManifest     map[string]string
	assetManifestOnce sync.Once
)

// GetAssetPath returns the versioned asset path for production or fallback for dev
// Example: "bundle" -> "bundle.abc123.js" in production, "bundle.js" in dev
func GetAssetPath(name string) string {
	loadAssetManifest()

	// If manifest exists and has the asset, verify the file exists before using it
	if assetManifest != nil {
		if hashed, ok := assetManifest[name]; ok {
			// Extract just the filename (remove .js extension if present in manifest)
			filename := hashed
			if filepath.Ext(filename) != ".js" {
				filename += ".js"
			}

			// Check if the hashed file actually exists
			hashedPath := filepath.Join("internal", "assets", "static", "js", filename)
			if _, err := os.Stat(hashedPath); err == nil {
				return "/static/js/" + hashed
			}
			// File doesn't exist, fall through to development fallback
		}
	}

	// Fallback to non-hashed filename (development)
	return "/static/js/" + name + ".js"
}

// loadAssetManifest reads the manifest.json file (once)
func loadAssetManifest() {
	assetManifestOnce.Do(func() {
		manifestPath := filepath.Join("internal", "assets", "static", "js", "manifest.json")

		// #nosec G304 - manifestPath is constructed from hardcoded strings, not user input
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			// File doesn't exist in dev mode - that's OK
			return
		}

		var manifest map[string]string
		if err := json.Unmarshal(data, &manifest); err != nil {
			return
		}

		assetManifest = manifest
	})
}

// ReloadAssetManifest forces a reload of the manifest (useful for hot reload)
func ReloadAssetManifest() {
	assetManifestOnce = sync.Once{}
	assetManifest = nil
}
