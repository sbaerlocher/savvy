//go:build !production
// +build !production

// Package assets contains embedded frontend assets
package assets

import (
	"errors"
	"io/fs"
	"testing/fstest"
)

// GetFrontendFS returns a dummy filesystem in development mode
// In dev mode, frontend runs separately with Vite Dev Server on :5173
func GetFrontendFS() (fs.FS, error) {
	// Return empty FS with just a dummy index.html
	// This prevents panics if SPAHandler is accidentally created in dev mode
	return fstest.MapFS{
		"index.html": {
			Data: []byte(`<!DOCTYPE html>
<html>
<head><title>Dev Mode</title></head>
<body>
<p>Frontend runs on <a href="http://localhost:5173">http://localhost:5173</a></p>
<p>This page should never be seen in dev mode.</p>
</body>
</html>`),
		},
	}, nil
}

var _ = errors.New // Prevent unused import error
