# Development Guide

**Last Updated**: 2026-03-02
**Status**: ✅ Production-Ready Development Setup

---

## 📚 Table of Contents

- [Prerequisites](#prerequisites)
- [Git Hooks](#git-hooks)
- [Quick Start](#quick-start)
- [Hot Reload (Air)](#hot-reload-air)
- [Testing](#testing)
- [PWA Development](#pwa-development)
- [Build Process](#build-process)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

---

## Prerequisites

### Required (Docker Development)

```bash
# Docker & Docker Compose
docker --version
docker compose version

# just (task runner, see https://just.systems for install options)
just --version
```

**Note**: Go and Node.js are NOT required on your host machine. All development
happens inside Docker containers with hot reload enabled.

### Optional (Advanced Use Cases Only)

```bash
# Go 1.25+ (only if building outside Docker)
go version

# Node.js 18+ & npm (only if building outside Docker)
node --version
npm --version

# Air (only needed for local non-Docker development - NOT recommended)
go install github.com/air-verse/air@latest
```

---

## Git Hooks

The repo ships [`lefthook.yml`](lefthook.yml) with pre-commit (gofmt,
golangci-lint, prettier, eslint, markdownlint) and pre-push (go test,
format-check, lint, typecheck) hooks. Like every project command they run in
the dde containers, not on the host: [`.lefthook/dde-run.sh`](.lefthook/dde-run.sh)
executes Go tooling in `api` and Node tooling in `client`, as the container's
non-root user, so rewritten files stay owned by the host user.

- **Stack running** — the hook runs in the container; its exit code gates the
  commit/push.
- **Stack not running** — the hook fails with a hint instead of passing
  unchecked. Start the stack with `dde project:up`, or skip one hook with
  `LEFTHOOK_EXCLUDE=<command> git commit`.
- **dde not installed** — the hooks are skipped with a notice. CI does not run
  lefthook; it runs its own gates.

The pre-push hooks only run for affected files: `go-test` for everything except
`client/` and Markdown, the frontend checks for files under `client/`. A
Markdown-only push therefore needs no running stack; other files such as
`LICENSE`, `deploy/**` or `.github/**` still trigger `go-test`. The file list
comes from the checked-out branch, so pushing a different ref
(`git push origin <branch>`) is gated by the current checkout.

dde resolves the stack from the current directory, so a **git worktree needs
its own stack** (`dde project:up` in the worktree): the main clone's
containers mount the main clone's files and would check the wrong tree. In
return a worktree needs no host `npm ci` — the frontend hooks use the client
container's `node_modules`.

Notes:

- `go-test` points `DATABASE_URL` at an unresolvable `.invalid` host, so the
  DB-backed tests skip instead of migrating and cleaning the dev database.
  Tests also skip when `DATABASE_URL` is unset. DB and E2E gates remain CI-only.
- `golangci-lint` is not part of the api image. The hook builds the version
  pinned in `lefthook.yml` (`GOLANGCI_LINT_VERSION`, bumped by Renovate) with
  `go run` on first use; the build cache lives in the container, so the first
  lint after recreating the stack takes a while.
- `markdownlint` runs `npx markdownlint-cli` in the client container. That
  container only mounts `client/`, so
  [`.lefthook/markdownlint.sh`](.lefthook/markdownlint.sh) copies the staged
  files and `.markdownlint.json` into a git-ignored scratch directory under
  `client/` for the run.

lefthook is opt-in and skippable with `LEFTHOOK=0 git commit`.

### Root-owned files in the bind mount

dde starts the containers as root, so files written into the bind mount as
root (`dde project:exec --root`, git inside the container) end up root-owned
on the host and break host git and lefthook with `EACCES`. Two dde hooks keep
the working tree user-owned:

- `.dde/hooks/project.up.pre/10-volume-mountpoints.sh` pre-creates the
  anonymous-volume mountpoints (`tmp/`, `client/node_modules/`, derived from
  `docker compose config`) as the host user, so Docker does not create them as
  root. An empty root-owned mountpoint is recreated; a non-empty one gets a
  one-time fix hint.
- `.dde/hooks/project.down.pre/10-chown-bind-mount.sh` chowns root-owned
  files under `/app` (and, in a worktree, the main clone's `.git`) back to the
  host user before the stack stops.

Both are best effort and only warn. The chown needs the running `api`
container; after a crash or a plain `docker stop`, run
`dde project:up && dde project:down` once.

---

## Quick Start

### dde (Required for Development)

```bash
# Start all services (PostgreSQL + Backend + Frontend)
dde project:up

# Access (routed by dde traefik on a single origin):
# - App (Frontend + API): https://savvy.test
# - PostgreSQL: dde stock postgres (no host port; reach it inside the dde network)
```

**Hot Reload enabled**: Air (Go) + Vite (SvelteKit) detect changes automatically.

### ⚠️ Why Docker is Required

**CRITICAL**: Local development without Docker does NOT work due to network
architecture:

```bash
# ❌ WRONG: NEVER start local dev servers!
npm run dev          # NEVER! Causes proxy errors
cd client && npm run dev  # NEVER! Vite cannot access Docker network
air                  # NEVER! Go backend must run in Docker
```

**Why Docker is mandatory**:

- ✅ Vite proxy only works in Docker network (`http://api:8080`)
- ✅ Consistent development environment (PostgreSQL, Go, Node.js)
- ✅ Prevents port conflicts and network issues
- ❌ Local processes cannot access Docker containers
  (`api:8080` is only resolvable in Docker network)

**When making changes**:

1. Edit files locally (e.g., in VSCode)
2. Air (Go) and Vite (Node.js) in Docker detect changes automatically
3. No manual restarts needed

---

## Hot Reload (Air)

Savvy uses **[Air](https://github.com/air-verse/air)** for automatic rebuild
and restart on code changes in Docker.

### Configuration (`.air.toml`)

The project uses a single `.air.toml` file with **polling mode** enabled for
Docker compatibility:

```toml
# .air.toml (excerpt)
[build]
  poll = true              # ✅ Required for Docker volumes
  poll_interval = 500      # Check every 500ms
  delay = 1000             # Wait 1s after change before rebuild
  exclude_unchanged = true # Skip files with unchanged content
```

**Why polling?**

- Docker volume mounts don't support native filesystem events (inotify)
- Polling checks for changes every 500ms
- Works reliably in Docker across all platforms

### Watched Directories

- `cmd/` - Application entrypoint
- `internal/` - All Go code (handlers, services, repositories, models)
- `client/src/` - SvelteKit frontend (watched by Vite, not Air)

### Ignored

- `tmp/` - Air build artifacts
- `vendor/`, `node_modules/` - Dependencies
- `*_test.go` - Test files (use `go test` instead)
- `frontend/` - Old frontend directory (deprecated)

---

## Testing

### Go Backend Tests

```bash
# Run all tests
go test ./...

# Run with coverage
just test-coverage

# Run with race detection
go test -race ./...
```

**Coverage**: Services 80.0%, Handlers 80.4%, Models 100%

### Frontend Unit Tests (Vitest)

```bash
cd client

# Run tests
npm test

# Run with coverage
npm run test:coverage
```

### E2E Tests (Playwright)

```bash
cd client

# Install browsers (first time only)
npm run playwright:install

# Run all tests (all browsers in parallel)
npm run test:e2e

# Run individual test file across all browsers (recommended for development)
npm run test:e2e:run -- tests/e2e/auth.spec.ts

# List tests without running (instant, no Docker startup)
npm run test:e2e:run -- tests/e2e/auth.spec.ts --list

# Run with specific browser (faster for quick testing)
npm run test:e2e:run -- tests/e2e/cards.spec.ts --project=chromium

# Run in headed mode (show browser)
npm run test:e2e:run -- tests/e2e/sharing.spec.ts --headed

# Debug mode (step-through debugging)
npm run test:e2e:run -- tests/e2e/import.spec.ts --debug

# Run in UI mode
npm run test:e2e:ui

# Run mobile tests
npm run test:e2e:mobile

# Browser-specific
npm run test:e2e:chromium    # Chromium only
npm run test:e2e:firefox     # Firefox only
```

**Coverage**: 70+ E2E tests in 23 test files
(Auth, CRUD, Sharing, Favorites, Dashboard, Batch Ops, 2FA, Import/Export, Profile, Security)

**Test Files**: `admin`, `auth`, `batch-operations`, `cards`, `config-and-features`, `dashboard`, `error-handling`,
`favorites`, `form-validation`, `gift-cards`, `import`, `internationalization`, `merchant-management`,
`notifications`, `offline-storage`, `pagination`, `password-reset`, `profile`, `security`, `sharing`,
`two-factor`, `verify-email`, `vouchers`

**Docker Environment**: Tests automatically start PostgreSQL + app-e2e containers via `globalSetup`. Use `--list` flag
to skip Docker startup for quick test listing.

---

## PWA Development

Savvy is a **Progressive Web App** with offline support.

### Features

- ✅ **Offline Viewing** - Cached cards, vouchers, gift cards
- ✅ **Service Worker** - Auto-caching via Vite PWA Plugin
- ✅ **Installable** - Add to homescreen

### Build

```bash
cd client
npm run build        # Build with Service Worker
npm run build:embed  # Build + copy to Go assets
```

### Testing PWA

1. `npm run preview` - Serve production build
2. DevTools → Application → Service Workers
3. Test offline mode

---

## Build Process

### Development (Docker)

```bash
# Start all services with hot reload
dde project:up

# Air (Go) and Vite (SvelteKit) automatically detect changes
# No manual rebuild needed during development
```

### Production Build

```bash
# Build frontend + backend binary
just build

# Or manually:
cd client && npm run build:embed  # Build SvelteKit + copy to Go assets
go build -o bin/server ./cmd/server
```

### Docker Production Build

```bash
# Build production Docker image (multi-stage build)
docker build -t savvy:latest .
```

---

## Troubleshooting

### Air doesn't detect changes in Docker

**Problem**: Code changes not triggering rebuild

**Solution**:

```bash
# Check polling is enabled in .air.toml
grep "poll = " .air.toml
# Should show: poll = true

# Restart the development stack
dde project:down
dde project:up
```

### Frontend not loading / 404 errors

**Problem**: SvelteKit pages return 404

**Solution (Development)**:

```bash
# Restart the client container (Vite dev server)
docker compose restart client

# Or rebuild only the client service (after dde project:up)
docker compose up -d --build client
```

**Solution (Production/Embedded Build)**:

```bash
# Rebuild frontend and restart
cd client && npm run build:embed
docker compose restart api
```

### "Vite proxy error" or API requests failing

**Problem**: Frontend cannot reach backend API

**Symptoms**:

- Console shows `Failed to fetch` or proxy errors
- API calls to `/api/v1/*` fail

**Solution**: This happens when running Vite **outside Docker**. You MUST use:

```bash
dde project:up  # ✅ Correct - Vite runs inside Docker
```

**Why?** Vite's proxy config points to `http://api:8080`, which is only
resolvable inside the Docker network.

### Port already in use (5173, 8080, 5432)

**Problem**: `Error: Port already in use`

**Solution**:

```bash
# Stop the development stack (recommended)
dde project:down

# Or manually check what's using the port
lsof -i :5173  # Frontend (Vite)
lsof -i :8080  # Backend (Go API)
lsof -i :5432  # PostgreSQL

# Kill process if needed (replace PID with actual process ID)
kill -9 <PID>
```

### Database connection refused

**Problem**: Backend cannot connect to PostgreSQL

**Solution**:

```bash
# Check configured services and restart the development stack
dde project:service:list
dde project:down
dde project:up
```

### Changes not reflecting after edit

**Problem**: File changes don't trigger reload

**Solution**:

```bash
# Check Air logs
docker compose logs -f api

# Check Vite logs
docker compose logs -f client

# If still not working, rebuild
dde project:down
docker compose build --no-cache
dde project:up
```

---

## Best Practices

### ✅ DO

- **ALWAYS use `dde project:up`** for development (required, not optional)
- Edit files locally, let Air/Vite in Docker handle reloads
- Run tests before committing (`go test ./...`, `npm run test:e2e`)
- Write tests for new features (handlers, services, repositories)
- Follow Clean Architecture (handlers → services → repositories)
- Use `just` recipes for common tasks (`just up`, `just test-coverage`)

### ❌ DON'T

- **NEVER run local dev servers** (`npm run dev`, `air`) - Vite proxy breaks
- **NEVER disable polling** in `.air.toml` - Volume changes won't detect in Docker
- Don't commit generated files (`tmp/`, `node_modules/`, `client/dist/`)
- Don't use direct database access in handlers (use services/repositories)
- Don't modify `.air.toml` poll settings without testing

---

## Commands Cheatsheet

```bash
# Development (Docker)
just                         # List all available recipes
just dev                     # Start development (recommended, alias for just up)
just up                      # Start Docker containers with hot reload
just down                    # Stop Docker containers
just restart                 # Restart containers
just rebuild                 # Rebuild containers from scratch
just ps                      # Show running containers

# Logs
just logs                    # API logs (api container)
just logs-client             # Frontend logs (client container)
just logs-all                # All logs (all containers)

# Shell Access
just shell                   # Open shell in API container
dde project:db               # Open PostgreSQL shell

# Testing
just test                    # Run all Go tests
just test-core               # Run core tests (services + models)
just test-coverage           # Run tests with coverage report
go test ./...                # Direct: all Go tests
go test -race ./...          # Direct: all tests with race detection
npm run test:e2e             # Frontend E2E tests (Playwright)
npm test                     # Frontend unit tests (Vitest)

# Database (via dde plugins)
dde project:db:seed            # Seed database with test data
dde project:db:migrate-up      # Apply all pending migrations
dde project:db:migrate-down    # Rollback last migration
dde project:db:migrate-status  # Show migration status

# Build
just build                   # Production build (frontend + backend binary)
npm run build:embed          # Frontend build + copy to Go assets
just clean                   # Remove build artifacts

# Quality
just lint                    # Run golangci-lint
just fmt                     # Format Go code
```

---

## References

- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Contribution Guidelines
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System Architecture
- **[OPERATIONS.md](OPERATIONS.md)** - Deployment Guide
- **[Air Docs](https://github.com/air-verse/air)** - Hot Reload
- **[Playwright Docs](https://playwright.dev/)** - E2E Testing

---

**Need help?** Check [SUPPORT.md](SUPPORT.md)
