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
	assetManifestErr  error
)

// GetAssetPath returns the versioned asset path for production or fallback for dev
// Example: "bundle" -> "bundle.abc123.js" in production, "bundle.js" in dev
func GetAssetPath(name string) string {
	loadAssetManifest()

	// If manifest exists and has the asset, use it
	if assetManifest != nil {
		if hashed, ok := assetManifest[name]; ok {
			return "/static/js/" + hashed
		}
	}

	// Fallback to non-hashed filename (development)
	return "/static/js/" + name + ".js"
}

// loadAssetManifest reads the manifest.json file (once)
func loadAssetManifest() {
	assetManifestOnce.Do(func() {
		manifestPath := filepath.Join("internal", "assets", "static", "js", "manifest.json")

		data, err := os.ReadFile(manifestPath)
		if err != nil {
			// File doesn't exist in dev mode - that's OK
			if !os.IsNotExist(err) {
				assetManifestErr = err
			}
			return
		}

		var manifest map[string]string
		if err := json.Unmarshal(data, &manifest); err != nil {
			assetManifestErr = err
			return
		}

		assetManifest = manifest
	})
}

// ReloadAssetManifest forces a reload of the manifest (useful for hot reload)
func ReloadAssetManifest() {
	assetManifestOnce = sync.Once{}
	assetManifest = nil
	assetManifestErr = nil
}
