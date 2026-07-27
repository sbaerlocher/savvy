//go:build production
// +build production

package setup

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

// The production build must not expose the docs at all — that is the whole
// point of keeping the generated spec package out of the binary.
func TestRegisterOpenAPIDocsIsNoOpInProduction(t *testing.T) {
	e := echo.New()
	registerOpenAPIDocs(e)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/docs/index.html", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /api/v1/docs/index.html = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
