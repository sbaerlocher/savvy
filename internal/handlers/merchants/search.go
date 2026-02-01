// Package merchants contains HTTP request handlers for merchant operations.
package merchants

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// Search returns merchants as HTML options for autocomplete
func (h *Handler) Search(c echo.Context) error {
	query := c.QueryParam("q")

	// Sanitize query to prevent LIKE-pattern injection
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, "%", "\\%")
	query = strings.ReplaceAll(query, "_", "\\_")

	// Limit query length to prevent DoS
	if len(query) > 100 {
		return c.String(http.StatusBadRequest, "Query too long")
	}

	merchants, err := h.merchantService.SearchMerchants(c.Request().Context(), query)
	if err != nil {
		c.Logger().Errorf("Failed to search merchants: %v", err)
		return c.String(http.StatusInternalServerError, "Search failed")
	}

	// Return HTML options for datalist
	html := ""
	for _, m := range merchants {
		html += fmt.Sprintf(`<option value="%s" data-id="%s">`, m.Name, m.ID.String())
	}

	return c.HTML(http.StatusOK, html)
}
