# Release Tool

Platform-independent Go tool for building and releasing Savvy.

## Features

✅ **Version Sync**: Synchronizes `package.json` with Git tags
✅ **Client Build**: npm install + vite build + asset copy
✅ **Full Build**: All steps in one command
✅ **Platform-Independent**: Works on Windows, Linux, macOS
✅ **GoReleaser Integration**: Automatic version detection

## Usage

### Commands

```bash
# Sync package.json version with Git tag
go run cmd/release-tool/main.go sync-version v1.0.0

# Build SvelteKit client and copy to assets
go run cmd/release-tool/main.go build-client

# Full build (sync + build)
go run cmd/release-tool/main.go build-all v1.0.0
```

### Environment Variables

- `GORELEASER_CURRENT_TAG`: Auto-detected by GoReleaser (e.g., `v1.0.0`)

### Version Detection Priority

1. Command-line argument: `go run ... sync-version v1.0.0`
2. Environment variable: `GORELEASER_CURRENT_TAG=v1.0.0`
3. Git describe: `git describe --tags --abbrev=0`

## Examples

### Local Development

```bash
# Test version sync
go run cmd/release-tool/main.go sync-version v1.0.0-dev
grep version client/package.json

# Build client
go run cmd/release-tool/main.go build-client
ls -lh internal/assets/client/

# Full build
go run cmd/release-tool/main.go build-all v1.0.0
```

### GoReleaser Integration

In `.goreleaser.yaml`:

```yaml
before:
  hooks:
    - go mod download
    - go run cmd/release-tool/main.go build-all
    - go mod tidy
```

GoReleaser automatically sets `GORELEASER_CURRENT_TAG`, so no version argument is needed.

### CI/CD

```bash
# GitHub Actions
- name: Build
  run: go run cmd/release-tool/main.go build-all
  env:
    GORELEASER_CURRENT_TAG: ${{ github.ref_name }}
```

## What it Does

### `sync-version`

1. Reads version from args/env/git
2. Removes `v` prefix (v1.0.0 → 1.9.0)
3. Parses `client/package.json`
4. Updates version field
5. Writes back with proper formatting (2-space indent, trailing newline)

### `build-client`

1. `npm ci --quiet` in `client/`
2. `npm run build` (Vite build)
3. Removes `internal/assets/client/`
4. Copies `client/build/` → `internal/assets/client/`

### `build-all`

1. Runs `sync-version`
2. Runs `build-client`
3. Combined output with progress indicators

## Error Handling

- ❌ No version found → Exit 1
- ❌ package.json parse error → Exit 1
- ❌ npm build failed → Exit 1
- ❌ Asset copy failed → Exit 1

All errors are logged to stderr with descriptive messages.

## Advantages over Shell Scripts

| Feature         | Shell Script  | Go Tool       |
| --------------- | ------------- | ------------- |
| Platform        | Linux/macOS   | ✅ All        |
| Dependencies    | bash, npm, jq | ✅ Go + npm   |
| Error Handling  | Basic         | ✅ Robust     |
| Type Safety     | ❌            | ✅            |
| Maintainability | Medium        | ✅ High       |
| Testing         | Hard          | ✅ Easy       |
| Consistency     | ❌            | ✅ Go project |

## Testing

```bash
# Run tests
go test ./cmd/release-tool/...

# Build binary
go build -o bin/release-tool cmd/release-tool/main.go

# Install globally
go install ./cmd/release-tool

# Use anywhere
release-tool build-all v1.0.0
```

## Development

```bash
# Add new command
func main() {
    switch os.Args[1] {
    case "new-command":
        handleNewCommand()
    }
}

# Add flags
import "flag"
var verbose = flag.Bool("verbose", false, "verbose output")
flag.Parse()
```

## License

Part of the Savvy project. See [LICENSE](../../LICENSE).
