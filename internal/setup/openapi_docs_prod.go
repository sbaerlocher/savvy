//go:build production
// +build production

package setup

import "github.com/labstack/echo/v5"

// registerOpenAPIDocs is a no-op in production builds: the Swagger UI is a
// development tool, and leaving the generated docs package out keeps swag's
// parser (and golang.org/x/tools) out of the shipped binary.
// See openapi_docs_dev.go for the development implementation.
func registerOpenAPIDocs(_ *echo.Echo) {}
