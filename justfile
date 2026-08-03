# justfile for Savvy — org-wide task-runner standard (sbaerlocher/.github).
#
# Runner layering (do not collapse these into one another):
#   1. dde project:*   Runtime layer, priority 1, untouched (db, migrations,
#                      observability, mailpit, e2e). just delegates to dde,
#                      never replaces it.
#   2. just <verb>     Task entrypoint. The same verbs mean the same thing in
#                      every repo.
#   3. go / npm        Toolchain underneath, called *by* just, never a peer
#                      runner.
#
# Development Workflow (Docker-based):
#   just dev      # Start Docker development (alias for just up)
#   just logs     # View application logs
#   just down     # Stop containers
#
# Production:
#   - Production images built via GitHub Actions (.github/workflows/release.yml)
#   - Use `just build` for local binary builds

APP_NAME := "savvy"
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
BUILD_TIME := `date -u '+%Y-%m-%d_%H:%M:%S'`

# Container names (development)
APP_CONTAINER := "savvy-api-1"

# Helm configuration
HELM_RELEASE := "savvy"
HELM_CHART := "deploy/helm"
HELM_TEST_VALUES := "deploy/helm/values-test.yaml"

# default → list available recipes
default:
    @just --list

# Show recipes plus the dde-driven commands just does not wrap
help:
    @just --list
    @echo ""
    @echo "Database (via dde plugins):"
    @echo "  dde project:db:migrate-up      Apply all pending migrations"
    @echo "  dde project:db:migrate-down    Rollback last migration"
    @echo "  dde project:db:migrate-reset   Rollback all migrations"
    @echo "  dde project:db:migrate-status  Show applied migrations"
    @echo "  dde project:db:migrate-to      Migrate to version (-- TARGET_VERSION)"
    @echo "  dde project:db:seed            Seed database with test data"
    @echo "  dde project:db:open            Open PostgreSQL shell"
    @echo ""
    @echo "Observability + mail (via dde):"
    @echo "  dde project:observability      Start Grafana/Prometheus/Loki/Tempo"
    @echo "  Mailpit UI: dde stock service (see dde dashboard)"

# ==============================================================================
# STANDARD VERBS
# ==============================================================================

# dev → start the local dev loop (alias for up)
dev: up

# build → build binary and JS bundle
build:
    @echo "🔨 Building {{ APP_NAME }}..."
    @echo "📦 Building client (SvelteKit) using release-tool..."
    go run cmd/release-tool/main.go build-client
    @echo "🔧 Building Go binary with embedded client..."
    go build -tags production -ldflags="-s -w -X main.version={{ VERSION }} -X main.buildTime={{ BUILD_TIME }}" -o bin/server cmd/server/main.go
    @echo "✓ Build complete (version: {{ VERSION }})"

# test → run all tests
test:
    @echo "🧪 Running all tests..."
    go test -mod=mod -v -p=1 ./...

# lint → run golangci-lint
lint:
    @echo "🔍 Running golangci-lint..."
    golangci-lint run

# fmt → format Go code
fmt:
    @echo "✨ Formatting Go code..."
    go fmt ./...

# ==============================================================================
# CORE
# ==============================================================================

# Run core tests (services + models)
test-core:
    @echo "🧪 Running core tests (services + models)..."
    go test -mod=mod -v -p=1 ./internal/services ./internal/models

# Run tests with coverage report
test-coverage:
    @echo "📊 Running tests with coverage..."
    go test -mod=mod -p=1 -coverprofile=coverage.out -covermode=atomic ./internal/services ./internal/models
    go tool cover -html=coverage.out -o coverage.html
    @echo "✓ Coverage report generated: coverage.html"
    @echo ""
    @go tool cover -func=coverage.out | grep total

# Run tests with coverage (CI mode)
test-coverage-ci:
    @echo "📊 Running tests with coverage (CI mode)..."
    go test -mod=mod -coverprofile=coverage.out -covermode=atomic ./internal/services ./internal/models
    @go tool cover -func=coverage.out | grep total

# Regenerate the OpenAPI spec from handler annotations
openapi:
    @echo "📖 Generating OpenAPI spec..."
    go tool swag init --v3.1 \
        --generalInfo internal/setup/routes.go \
        --dir ./ \
        --exclude docs \
        --parseInternal \
        --output docs/openapi \
        --outputTypes go,yaml \
        --packageName openapi \
        --propertyStrategy snakecase \
        --quiet
    @mv docs/openapi/swagger.yaml docs/openapi/openapi.yaml
    @go run ./cmd/openapi-fix
    @echo "✓ Spec written to docs/openapi/openapi.yaml"

# Remove build artifacts
clean:
    @echo "🧹 Cleaning up..."
    rm -rf bin/
    rm -rf client/build/
    rm -rf client/.svelte-kit/
    rm -rf internal/assets/client/
    @echo "✓ Cleanup complete"

# ==============================================================================
# DEVELOPMENT
# ==============================================================================
# IMPORTANT: Development MUST run in Docker (Vite proxy requires Docker network)
# See AGENTS.md for details

# Start Docker containers with hot reload
up:
    @echo "🚀 Starting development containers with hot reload..."
    @echo "📝 Watching: Go (Air) + SvelteKit (Vite HMR)"
    @echo "🔄 Auto-rebuild on file changes"
    @echo ""
    @echo "🔍 Checking for active E2E containers..."
    @if docker ps -q -f name=savvy-app-e2e 2>/dev/null | grep -q .; then \
        echo "⚠️  E2E container detected - stopping to avoid port conflicts..."; \
        docker compose stop app-e2e; \
        docker compose rm -f app-e2e; \
        echo "✓ E2E container stopped"; \
    else \
        echo "✓ No E2E container running"; \
    fi
    @echo ""
    docker compose up -d
    @echo "✓ Containers started"
    @echo "  Frontend: http://localhost:5173 (Vite dev server)"
    @echo "  Backend:  http://localhost:8080 (Go API)"
    @echo "  Mailpit:  http://localhost:8025 (Mail catcher)"
    @echo ""
    @echo "💡 View logs: just logs (API) | just logs-client (Frontend)"

# Stop Docker containers
down:
    @echo "⏹️  Stopping containers..."
    docker compose down
    @echo "✓ Containers stopped"

# Restart Docker containers
restart: down up

# Rebuild Docker containers from scratch
rebuild:
    @echo "🔄 Rebuilding development containers..."
    docker compose down
    docker compose build --no-cache
    docker compose up -d
    @echo "✓ Rebuild complete"
    @echo "💡 View logs: just logs"

# Show API logs
logs:
    docker compose logs -f api

# Show client logs
logs-client:
    docker compose logs -f client

# Show database logs
logs-db:
    docker compose logs -f postgres

# Show all logs
logs-all:
    docker compose logs -f

# Open shell in app container
shell:
    docker exec -it {{ APP_CONTAINER }} sh

# Database shell: use `dde project:db:open` (dde stock postgres, db savvy)

# Show running containers
ps:
    docker compose ps

# ==============================================================================
# DATABASE MIGRATIONS & SEED
# ==============================================================================
# Migrations and seeding are driven through dde plugins (.dde/plugins/db.*.sh):
#   dde project:db:migrate-up      Apply all pending migrations
#   dde project:db:migrate-down    Rollback last migration
#   dde project:db:migrate-reset   Rollback all migrations
#   dde project:db:migrate-status  Show applied migrations
#   dde project:db:migrate-to -- TARGET_VERSION
#   dde project:db:seed            Seed database with test data
#
# Mailpit + observability via dde:
#   dde project:observability      Start Grafana/Prometheus/Loki/Tempo
#   Mailpit UI: dde stock service (see dde dashboard)

# ==============================================================================
# HELM
# ==============================================================================

# Install with Helm
helm-install:
    @echo "📦 Installing Helm chart..."
    helm install {{ HELM_RELEASE }} {{ HELM_CHART }}

# Upgrade with Helm
helm-upgrade:
    @echo "🔄 Upgrading Helm release..."
    helm upgrade {{ HELM_RELEASE }} {{ HELM_CHART }}

# Uninstall Helm release
helm-uninstall:
    @echo "🗑️  Uninstalling Helm release..."
    helm uninstall {{ HELM_RELEASE }}

# Preview Helm templates (with test values)
helm-template:
    @echo "🔍 Rendering Helm templates..."
    helm template {{ HELM_RELEASE }} {{ HELM_CHART }} --values {{ HELM_TEST_VALUES }}

# Lint Helm chart (with test values)
helm-lint:
    @echo "🔍 Linting Helm chart..."
    helm lint {{ HELM_CHART }} --values {{ HELM_TEST_VALUES }}

# ==============================================================================
# KUSTOMIZE
# ==============================================================================

# Deploy development
kustomize-dev:
    @echo "🚀 Deploying to development..."
    kubectl apply -k deploy/kustomize/overlays/development

# Deploy staging
kustomize-staging:
    @echo "🚀 Deploying to staging..."
    kubectl apply -k deploy/kustomize/overlays/staging

# Deploy production
kustomize-prod:
    @echo "🚀 Deploying to production..."
    kubectl apply -k deploy/kustomize/overlays/production

# Preview development manifests
kustomize-preview-dev:
    @echo "🔍 Previewing development manifests..."
    kubectl kustomize deploy/kustomize/overlays/development

# Preview staging manifests
kustomize-preview-staging:
    @echo "🔍 Previewing staging manifests..."
    kubectl kustomize deploy/kustomize/overlays/staging

# Preview production manifests
kustomize-preview-prod:
    @echo "🔍 Previewing production manifests..."
    kubectl kustomize deploy/kustomize/overlays/production

# ==============================================================================
# MAINTENANCE
# ==============================================================================

# Update dependencies
deps:
    @echo "📦 Updating dependencies..."
    go mod tidy && go mod download
