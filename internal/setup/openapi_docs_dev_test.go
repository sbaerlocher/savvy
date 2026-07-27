//go:build !production
// +build !production

package setup

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes needs a fully wired dependency container (and a database), so
// the docs registration is exercised directly. The production counterpart is
// covered by openapi_docs_prod_test.go under the production build tag.
func TestRegisterOpenAPIDocsServesSpec(t *testing.T) {
	e := echo.New()
	registerOpenAPIDocs(e)

	tests := []struct {
		name string
		path string
		want int
	}{
		{name: "UI entrypoint", path: "/api/v1/docs/index.html", want: http.StatusOK},
		{name: "spec document", path: "/api/v1/docs/doc.json", want: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != tt.want {
				t.Errorf("GET %s = %d, want %d", tt.path, rec.Code, tt.want)
			}
		})
	}
}

// Guards the blank import of the generated package: without it the UI still
// loads, but doc.json serves swag's placeholder instead of Savvy's spec — a
// failure mode that returns a perfectly healthy 200.
func TestGeneratedSpecIsRegistered(t *testing.T) {
	e := echo.New()
	registerOpenAPIDocs(e)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/docs/doc.json", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/docs/doc.json = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	for _, want := range []string{"Savvy API", "/cards/{id}/transfer"} {
		if !strings.Contains(body, want) {
			t.Errorf("served spec does not contain %q", want)
		}
	}
}
