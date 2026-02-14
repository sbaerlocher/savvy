package debug //nolint:revive // Standard package name, false positive

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintRoutes(t *testing.T) {
	e := echo.New()

	// Register some test routes
	e.GET("/", func(c echo.Context) error {
		return c.String(200, "home")
	})
	e.POST("/cards", func(c echo.Context) error {
		return c.String(201, "created")
	})
	e.GET("/cards/:id", func(c echo.Context) error {
		return c.String(200, "card")
	})
	e.PUT("/cards/:id", func(c echo.Context) error {
		return c.String(200, "updated")
	})
	e.DELETE("/cards/:id", func(c echo.Context) error {
		return c.String(204, "deleted")
	})

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call PrintRoutes
	PrintRoutes(e)

	// Restore stdout
	_ = w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected elements
	assert.Contains(t, output, "Registered Routes")
	assert.Contains(t, output, "METHOD")
	assert.Contains(t, output, "PATH")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "Total Routes:")

	// Verify routes are in output
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "POST")
	assert.Contains(t, output, "PUT")
	assert.Contains(t, output, "DELETE")
	assert.Contains(t, output, "/cards")
}

func TestPrintRoutes_EmptyRoutes(t *testing.T) {
	e := echo.New()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintRoutes(e)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should still print headers even with no routes
	assert.Contains(t, output, "Registered Routes")
	assert.Contains(t, output, "Total Routes: 0")
}

func TestPrintRoutes_Sorting(t *testing.T) {
	e := echo.New()

	// Register routes in random order
	e.POST("/zebra", func(_ echo.Context) error { return nil })
	e.GET("/apple", func(_ echo.Context) error { return nil })
	e.PUT("/banana", func(_ echo.Context) error { return nil })
	e.DELETE("/apple", func(_ echo.Context) error { return nil })
	e.GET("/banana", func(_ echo.Context) error { return nil })

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintRoutes(e)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Split into lines
	lines := strings.Split(output, "\n")

	// Find the route lines (skip headers and separators)
	var routeLines []string
	inRoutes := false
	for _, line := range lines {
		if strings.Contains(line, "----------------------------------------") {
			inRoutes = true
			continue
		}
		if inRoutes && strings.TrimSpace(line) != "" && !strings.Contains(line, "Total Routes") && !strings.Contains(line, "====") {
			routeLines = append(routeLines, line)
		}
	}

	// Verify routes are sorted by path, then by method
	require.Greater(t, len(routeLines), 0)

	// Check that /apple comes before /banana and /zebra
	appleIndex := -1
	bananaIndex := -1
	zebraIndex := -1

	for i, line := range routeLines {
		if strings.Contains(line, "/apple") {
			if appleIndex == -1 {
				appleIndex = i
			}
		}
		if strings.Contains(line, "/banana") {
			if bananaIndex == -1 {
				bananaIndex = i
			}
		}
		if strings.Contains(line, "/zebra") {
			if zebraIndex == -1 {
				zebraIndex = i
			}
		}
	}

	// Paths should be sorted alphabetically
	if appleIndex != -1 && bananaIndex != -1 {
		assert.Less(t, appleIndex, bananaIndex, "apple should come before banana")
	}
	if bananaIndex != -1 && zebraIndex != -1 {
		assert.Less(t, bananaIndex, zebraIndex, "banana should come before zebra")
	}
}

func TestPrintRoutes_WithParameters(t *testing.T) {
	e := echo.New()

	// Register routes with parameters
	e.GET("/users/:id", func(_ echo.Context) error { return nil })
	e.GET("/users/:id/posts/:postId", func(_ echo.Context) error { return nil })

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintRoutes(e)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify parameter routes are displayed
	assert.Contains(t, output, "/users/:id")
	assert.Contains(t, output, "/users/:id/posts/:postId")
}

func TestPrintRoutes_MultipleMethodsSamePath(t *testing.T) {
	e := echo.New()

	// Register multiple methods on same path
	e.GET("/api/resource", func(_ echo.Context) error { return nil })
	e.POST("/api/resource", func(_ echo.Context) error { return nil })
	e.PUT("/api/resource", func(_ echo.Context) error { return nil })
	e.DELETE("/api/resource", func(_ echo.Context) error { return nil })

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintRoutes(e)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// All methods should be listed
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "POST")
	assert.Contains(t, output, "PUT")
	assert.Contains(t, output, "DELETE")
	assert.Contains(t, output, "/api/resource")

	// Count occurrences of the path
	count := strings.Count(output, "/api/resource")
	assert.GreaterOrEqual(t, count, 4, "should have at least 4 routes for /api/resource")
}

func TestPrintRoutes_FormattingWidth(t *testing.T) {
	e := echo.New()

	// Register route with very long path
	longPath := "/api/v1/very/long/path/that/might/exceed/formatting/width"
	e.GET(longPath, func(_ echo.Context) error { return nil })

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintRoutes(e)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should still contain the long path
	assert.Contains(t, output, longPath)
}

func TestPrintRoutes_GroupRoutes(t *testing.T) {
	e := echo.New()

	// Create a group
	api := e.Group("/api")
	api.GET("/users", func(_ echo.Context) error { return nil })
	api.POST("/users", func(_ echo.Context) error { return nil })

	v1 := api.Group("/v1")
	v1.GET("/cards", func(_ echo.Context) error { return nil })

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintRoutes(e)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify grouped routes are displayed with full path
	assert.Contains(t, output, "/api/users")
	assert.Contains(t, output, "/api/v1/cards")
}

func TestPrintRoutes_OutputFormat(t *testing.T) {
	e := echo.New()
	e.GET("/test", func(_ echo.Context) error { return nil })

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintRoutes(e)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify format structure
	assert.Contains(t, output, "========================================") // Top border
	assert.Contains(t, output, "Registered Routes")                        // Title
	assert.Contains(t, output, "METHOD")                                   // Column header
	assert.Contains(t, output, "PATH")                                     // Column header
	assert.Contains(t, output, "NAME")                                     // Column header
	assert.Contains(t, output, "----------------------------------------") // Separator
	assert.Contains(t, output, "Total Routes:")                            // Footer
}
