# Development Guide

**Last Updated**: 2026-03-02
**Status**: ✅ Production-Ready Development Setup

---

## 📚 Table of Contents

- [Prerequisites](#prerequisites)
- [Git Hooks in Worktrees](#git-hooks-in-worktrees)
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

# Make (optional, for convenience commands)
make --version
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

## Git Hooks in Worktrees

The repo ships [`lefthook.yml`](lefthook.yml) with pre-commit (prettier,
eslint) and pre-push (format-check, lint, typecheck, go test) hooks. The
frontend hooks run `npm run …` in `client/`, which needs `client/node_modules`
— git-ignored and **not shared across git worktrees** (each worktree has its
own working tree).

A worktree created for frontend changes therefore has no `node_modules`. When
a push includes `client/` files, the frontend hooks can't resolve the
prettier/eslint/svelte-check binaries and **fail the push with a
missing-binary error** — the gate never actually runs and the check only
happens later in CI. (lefthook only _skips_ these `root: client/` commands
when no pushed file falls under `client/`, i.e. when there's nothing to gate.)
Before touching frontend files in a fresh worktree, install once:

```bash
cd client && npm ci
```

Node/npm come from `.nvmrc`; Docker stays the canonical dev path. DB and E2E
gates remain CI-only (they need Postgres/Playwright). lefthook is opt-in and
skippable with `LEFTHOOK=0 git commit` or `--no-verify`.

---

## Quick Start

### Docker Compose (Required for Development)

```bash
# Start all services (PostgreSQL + Backend + Frontend)
docker compose up

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
make test-coverage

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
docker compose up

# Air (Go) and Vite (SvelteKit) automatically detect changes
# No manual rebuild needed during development
```

### Production Build

```bash
# Build frontend + backend binary
make build

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

# Restart Docker Compose
docker compose restart
```

### Frontend not loading / 404 errors

**Problem**: SvelteKit pages return 404

**Solution (Development)**:

```bash
# Restart the client container (Vite dev server)
docker compose restart client

# Or rebuild if needed
docker compose up --build client
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
docker compose up  # ✅ Correct - Vite runs inside Docker
```

**Why?** Vite's proxy config points to `http://api:8080`, which is only
resolvable inside the Docker network.

### Port already in use (5173, 8080, 5432)

**Problem**: `Error: Port already in use`

**Solution**:

```bash
# Stop Docker containers (recommended)
docker compose down

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
# Check PostgreSQL is running
docker compose ps

# View PostgreSQL logs
docker compose logs postgres

# Restart services
docker compose restart
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
docker compose down
docker compose up --build
```

---

## Best Practices

### ✅ DO

- **ALWAYS use Docker Compose** for development (required, not optional)
- Edit files locally, let Air/Vite in Docker handle reloads
- Run tests before committing (`go test ./...`, `npm run test:e2e`)
- Write tests for new features (handlers, services, repositories)
- Follow Clean Architecture (handlers → services → repositories)
- Use `make` commands for common tasks (`make up`, `make test-coverage`)

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
make dev                     # Start development (recommended, alias for make up)
make up                      # Start Docker containers with hot reload
make down                    # Stop Docker containers
make restart                 # Restart containers
make rebuild                 # Rebuild containers from scratch
make ps                      # Show running containers

# Logs
make logs                    # API logs (api container)
make logs-client             # Frontend logs (client container)
make logs-db                 # Database logs (postgres container)
make logs-all                # All logs (all containers)

# Shell Access
make shell                   # Open shell in API container
dde project:db:open          # Open PostgreSQL shell

# Testing
make test                    # Run all Go tests
make test-core               # Run core tests (services + models)
make test-coverage           # Run tests with coverage report
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
make build                   # Production build (frontend + backend binary)
npm run build:embed          # Frontend build + copy to Go assets
make clean                   # Remove build artifacts

# Quality
make lint                    # Run golangci-lint
make fmt                     # Format Go code
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
