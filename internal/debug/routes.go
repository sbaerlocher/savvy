// Package debug provides development-only debugging utilities.
package debug

import (
	"fmt"
	"sort"

	"github.com/labstack/echo/v4"
)

// PrintRoutes prints all registered routes in a formatted table.
// Only call this in development mode (GO_ENV=development).
//
// Output format:
//
//	METHOD   PATH                                            NAME
//	GET      /                                               github.com/sbaerlocher/savvy/internal/handlers.HomeIndex
//	POST     /cards                                          github.com/sbaerlocher/savvy/internal/handlers/cards.(*Handler).Create
//
// Routes are sorted alphabetically by path, then by method.
func PrintRoutes(e *echo.Echo) {
	routes := e.Routes()

	// Sort by path first, then by method
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	fmt.Println("\n========================================")
	fmt.Println("Registered Routes")
	fmt.Println("========================================")
	fmt.Printf("%-8s %-50s %s\n", "METHOD", "PATH", "NAME")
	fmt.Println("----------------------------------------")

	for _, route := range routes {
		fmt.Printf("%-8s %-50s %s\n", route.Method, route.Path, route.Name)
	}

	fmt.Printf("\nTotal Routes: %d\n", len(routes))
	fmt.Println("========================================\n")
}
