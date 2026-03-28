package handlers

import (
	"embed"
	"net/http"

	"github.com/labstack/echo/v5"
)

//go:embed templates/reset.html
var resetFS embed.FS

// ResetHandler serves a standalone HTML page that clears all PWA state.
// This bypasses the SPA and service worker entirely.
type ResetHandler struct {
	html []byte
}

// NewResetHandler creates a new reset handler.
func NewResetHandler() *ResetHandler {
	html, err := resetFS.ReadFile("templates/reset.html")
	if err != nil {
		panic("failed to load reset.html: " + err.Error())
	}
	return &ResetHandler{html: html}
}

// ServeResetPage serves a self-contained HTML page that unregisters service workers,
// clears all caches, and redirects to login.
// GET /reset
func (h *ResetHandler) ServeResetPage(c *echo.Context) error {
	return c.HTMLBlob(http.StatusOK, h.html)
}
