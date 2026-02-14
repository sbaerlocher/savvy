// Package handlers provides HTTP handlers for the application.
package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"savvy/internal/assets"

	"github.com/labstack/echo/v4"
)

// SPAHandler serves the SvelteKit SPA
type SPAHandler struct {
	fileSystem http.FileSystem
}

// NewSPAHandler creates a new SPA handler
func NewSPAHandler() *SPAHandler {
	frontendFS, err := assets.GetFrontendFS()
	if err != nil {
		panic("failed to load frontend assets: " + err.Error())
	}

	return &SPAHandler{
		fileSystem: http.FS(frontendFS),
	}
}

// ServeStatic serves static files (JS, CSS, images, etc.)
func (h *SPAHandler) ServeStatic(c echo.Context) error {
	path := c.Request().URL.Path

	// Try to serve the file
	file, err := h.fileSystem.Open(path)
	if err != nil {
		// File not found - this is handled by ServeSPA
		return echo.ErrNotFound
	}
	defer func() { _ = file.Close() }()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		return echo.ErrNotFound
	}

	// Don't serve directories
	if stat.IsDir() {
		return echo.ErrNotFound
	}

	// Type assert to io.ReadSeeker for ServeContent
	readSeeker, ok := file.(io.ReadSeeker)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "File does not support seeking")
	}

	// Serve the file with proper content type
	http.ServeContent(c.Response(), c.Request(), path, stat.ModTime(), readSeeker)
	return nil
}

// ServeSPA serves the index.html for all SPA routes (fallback)
func (h *SPAHandler) ServeSPA(c echo.Context) error {
	// Serve index.html
	file, err := h.fileSystem.Open("index.html")
	if err != nil {
		slog.Error("Failed to open SPA index.html", "error", err) //nolint:gosec // URI is not user-controlled input
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load SPA")
	}
	defer func() { _ = file.Close() }()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		slog.Error("Failed to stat SPA index.html", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load SPA")
	}

	// Type assert to io.ReadSeeker for ServeContent
	readSeeker, ok := file.(io.ReadSeeker)
	if !ok {
		slog.Error("SPA index.html does not support seeking")
		return echo.NewHTTPError(http.StatusInternalServerError, "File does not support seeking")
	}

	// Serve index.html
	http.ServeContent(c.Response(), c.Request(), "index.html", stat.ModTime(), readSeeker)
	return nil
}
