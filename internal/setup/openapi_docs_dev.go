//go:build !production
// +build !production

package setup

import (
	_ "savvy/docs/openapi" // registers the generated spec with swag

	"github.com/labstack/echo/v5"
	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

// registerOpenAPIDocs mounts the Swagger UI for the generated spec.
//
// Development only, for two reasons: the API is session-authenticated and
// internal, and the generated docs package pulls swag's parser (and with it
// golang.org/x/tools) into the binary — roughly 11 MB that production has no
// use for. The production build gets the no-op twin in openapi_docs_prod.go.
//
// V3 is the OpenAPI 3.x handler; the plain EchoWrapHandler reads swag's v1
// registry, which our v2-generated spec never populates, and would answer
// doc.json with a 500.
func registerOpenAPIDocs(e *echo.Echo) {
	e.GET("/api/v1/docs/*", echoSwagger.EchoWrapHandlerV3())
}
