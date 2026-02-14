// Package main provides a release tool for building and versioning the Savvy application.
// It synchronizes versions across package.json and version.json, and builds the SvelteKit client.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	clientDir   = "client"
	packageJSON = "client/package.json"
	assetsDir   = "internal/assets/client"
	buildDir    = "client/build"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "sync-version":
		if err := syncVersion(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
	case "build-client":
		if err := buildClient(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
	case "build-all":
		if err := buildAll(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown command: %s\n", command) // #nosec G705 -- CLI tool, no XSS risk
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Savvy Release Tool

Usage:
  release-tool <command> [options]

Commands:
  sync-version [version]    Sync package.json version with Git tag
  build-client              Build SvelteKit client and copy to assets
  build-all [version]       Sync version + build client (full build)

Examples:
  release-tool sync-version v1.9.0
  release-tool build-client
  release-tool build-all v1.9.0

Environment Variables:
  GORELEASER_CURRENT_TAG    Git tag from GoReleaser (auto-detected)`)
}

// syncVersion synchronizes package.json version with Git tag
func syncVersion() error {
	// Get version from args, env, or git
	version := getVersion()
	if version == "" {
		return fmt.Errorf("no version found (provide as argument, GORELEASER_CURRENT_TAG, or git tag)")
	}

	// Remove 'v' prefix
	version = strings.TrimPrefix(version, "v")

	fmt.Printf("🏷️  Syncing package.json version to: %s\n", version)

	// Read package.json
	data, err := os.ReadFile(packageJSON)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", packageJSON, err)
	}

	// Parse as raw JSON to preserve field order
	var rawPackage map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawPackage); err != nil {
		return fmt.Errorf("failed to parse package.json: %w", err)
	}

	// Update version
	versionJSON, err := json.Marshal(version)
	if err != nil {
		return fmt.Errorf("failed to marshal version: %w", err)
	}
	rawPackage["version"] = versionJSON

	// Write back with formatting
	output, err := json.MarshalIndent(rawPackage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package.json: %w", err)
	}
	output = append(output, '\n')

	if err := os.WriteFile(packageJSON, output, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", packageJSON, err)
	}

	fmt.Printf("✅ package.json version updated to %s\n", version)
	return nil
}

// buildClient builds SvelteKit client and copies to assets
func buildClient() error {
	fmt.Println("📦 Building SvelteKit client...")

	// 1. Install dependencies
	fmt.Println("   → Installing dependencies...")
	if err := runCommand(clientDir, "npm", "ci", "--quiet"); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	// 2. Build client
	fmt.Println("   → Building with Vite...")
	if err := runCommand(clientDir, "npm", "run", "build"); err != nil {
		return fmt.Errorf("npm build failed: %w", err)
	}

	// 3. Remove old assets
	fmt.Println("   → Cleaning old assets...")
	if err := os.RemoveAll(assetsDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old assets: %w", err)
	}

	// 4. Copy build to assets
	fmt.Println("   → Copying to internal/assets/client/...")
	if err := copyDir(buildDir, assetsDir); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	fmt.Println("✅ Client build completed")
	return nil
}

// buildAll runs full build: sync version + build client
func buildAll() error {
	fmt.Println("🚀 Starting full build...")

	if err := syncVersion(); err != nil {
		return err
	}

	fmt.Println()

	if err := buildClient(); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("✅ Full build completed successfully!")
	return nil
}

// getVersion retrieves version from args, env, or git
func getVersion() string {
	// 1. Command-line argument
	if len(os.Args) > 2 {
		return os.Args[2]
	}

	// 2. GoReleaser environment
	if version := os.Getenv("GORELEASER_CURRENT_TAG"); version != "" {
		return version
	}

	// 3. Git describe
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}

	return ""
}

// runCommand executes a command in a specific directory
func runCommand(dir string, name string, args ...string) error {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204 -- args are hardcoded in callers, not user input
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0750); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			data, err := os.ReadFile(srcPath) // #nosec G304 - srcPath is controlled (copyDir internal use)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0600); err != nil {
				return err
			}
		}
	}

	return nil
}
