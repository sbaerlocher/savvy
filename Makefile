# Makefile for Savvy System
#
# Development Workflow:
# 1. make up       # Start development with hot reload
# 2. make logs     # View application logs
# 3. make dev      # Run locally with Air
#
# Production Build:
# 1. make docker-build  # Build production image
# 2. make docker-run    # Run production container

APP_NAME := savvy
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
VERSION_CLEAN := $(shell echo $(VERSION) | sed 's/^v//')
VERSION_SHORT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DOCKER_IMAGE := ghcr.io/sbaerlocher/container/$(APP_NAME)

# Container names
APP_CONTAINER := savvy-app
DB_CONTAINER := savvy-postgres

# ==============================================================================
# HELP
# ==============================================================================

.PHONY: help
help:
	@echo "Savvy System Build Commands:"
	@echo ""
	@echo "Core:"
	@echo "  generate       Generate templ templates"
	@echo "  build          Build binary and JS bundle (includes generate)"
	@echo "  test           Run tests"
	@echo "  clean          Remove build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  dev            Start development with Air hot reload"
	@echo "  dev-simple     Run without hot reload"
	@echo "  up             Start Docker containers"
	@echo "  down           Stop Docker containers"
	@echo "  restart        Restart Docker containers"
	@echo "  rebuild        Rebuild Docker containers"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build   Build production image with versioning"
	@echo "  docker-run     Run production container"
	@echo "  build-prod     Build production image (alias)"
	@echo "  logs           Show app logs"
	@echo "  logs-db        Show database logs"
	@echo "  shell          Open shell in app container"
	@echo "  db-shell       Open PostgreSQL shell"
	@echo ""
	@echo "Database:"
	@echo "  migrate-up     Apply all pending migrations"
	@echo "  migrate-down   Rollback last migration"
	@echo "  migrate-reset  Rollback all migrations"
	@echo "  migrate-status Show applied migrations"
	@echo "  migrate-to     Migrate to specific version (VERSION=...)"
	@echo "  seed           Seed database with test data"
	@echo ""
	@echo "Helm:"
	@echo "  helm-install   Install with Helm"
	@echo "  helm-upgrade   Upgrade with Helm"
	@echo "  helm-template  Preview Helm templates"
	@echo "  helm-lint      Lint Helm chart"
	@echo ""
	@echo "Kustomize:"
	@echo "  kustomize-dev        Deploy development"
	@echo "  kustomize-staging    Deploy staging"
	@echo "  kustomize-prod       Deploy production"
	@echo "  kustomize-preview-*  Preview manifests"
	@echo ""
	@echo "Quality:"
	@echo "  lint           Run golangci-lint"
	@echo "  fmt            Format Go code"
	@echo "  deps           Update dependencies"

# ==============================================================================
# CORE TARGETS
# ==============================================================================

.PHONY: generate
generate:
	@echo "⚡ Generating templ templates..."
	templ generate
	@echo "✓ Generation complete"

.PHONY: build
build: generate service-worker
	@echo "🔨 Building $(APP_NAME)..."
	@echo "📦 Installing npm dependencies..."
	npm install
	@echo "📋 Building JS bundle (Alpine.js + html5-qrcode + scanner)..."
	npm run build
	@echo "🔧 Building Go binary..."
	go build -ldflags="-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)" -o bin/server cmd/server/main.go
	@echo "✓ Build complete (version: $(VERSION))"

.PHONY: service-worker
service-worker:
	@echo "⚙️  Generating service-worker.js with version $(VERSION)..."
	@sed 's/__VERSION__/$(VERSION)/g' internal/assets/static/service-worker.js.tmpl > internal/assets/static/service-worker.js
	@echo "✓ Service Worker generated (version: $(VERSION))"

.PHONY: test
test:
	@echo "🧪 Running tests..."
	go test -v ./...

.PHONY: clean
clean:
	@echo "🧹 Cleaning up..."
	rm -rf bin/
	rm -rf static/js/bundle.js
	rm -rf static/js/bundle.js.map
	rm -f internal/templates/*_templ.go
	rm -f internal/assets/static/service-worker.js
	@echo "✓ Cleanup complete"

.PHONY: clean-port
clean-port:
	@echo "🧹 Cleaning up port 3000..."
	@lsof -ti:3000 | xargs -r kill -9 || echo "No process found on port 3000"
	@echo "✓ Port 3000 is now free"

.PHONY: clean-all
clean-all: clean clean-port
	@echo "🧹 Full cleanup completed"

# ==============================================================================
# DEVELOPMENT
# ==============================================================================

.PHONY: dev
dev:
	@echo "🚀 Starting development server with hot reload..."
	@echo "📝 Watching: *.go, *.templ files"
	@echo "🔄 Auto-rebuild on file changes"
	@echo ""
	air

.PHONY: dev-simple
dev-simple:
	@echo "🚀 Starting development server..."
	go run cmd/server/main.go

.PHONY: up
up:
	@echo "🚀 Starting development containers with Air hot reload..."
	@echo "📝 Watching: *.go, *.templ files"
	@echo "🔄 Auto-rebuild on file changes"
	docker compose up -d
	@echo "✓ Containers started"
	@echo "  App: http://localhost:3000"
	@echo ""
	@echo "💡 View logs: make logs"

.PHONY: down
down:
	@echo "⏹️  Stopping containers..."
	docker compose down
	@echo "✓ Containers stopped"

.PHONY: restart
restart: down up

.PHONY: rebuild
rebuild:
	@echo "🔄 Rebuilding development containers..."
	docker compose down
	docker compose build --no-cache
	docker compose up -d
	@echo "✓ Rebuild complete"
	@echo "💡 View logs: make logs"

.PHONY: logs
logs:
	docker compose logs -f app

.PHONY: logs-db
logs-db:
	docker compose logs -f postgres

.PHONY: logs-all
logs-all:
	docker compose logs -f

.PHONY: shell
shell:
	docker exec -it $(APP_CONTAINER) sh

.PHONY: db-shell
db-shell:
	docker exec -it $(DB_CONTAINER) psql -U savvy -d savvy

.PHONY: ps
ps:
	docker compose ps

# ==============================================================================
# DOCKER (Production)
# ==============================================================================

.PHONY: docker-build
docker-build:
	@echo "🐳 Building production Docker image..."
	@echo "   Version: $(VERSION)"
	@echo "   Build Time: $(BUILD_TIME)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--target production .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	@echo "✓ Production image built"
	@echo "   $(DOCKER_IMAGE):$(VERSION)"
	@echo "   $(DOCKER_IMAGE):latest"

.PHONY: docker-run
docker-run:
	@echo "🐳 Starting production Docker container..."
	@echo "Configure via environment variables or .env file"
	docker run --rm -it \
		-e DATABASE_URL=${DATABASE_URL} \
		-e SESSION_SECRET=${SESSION_SECRET} \
		-e GO_ENV=production \
		-p 3000:3000 \
		$(DOCKER_IMAGE):latest

.PHONY: build-prod
build-prod: docker-build

# ==============================================================================
# DATABASE MIGRATIONS
# ==============================================================================

.PHONY: migrate-up
migrate-up:
	@echo "🚀 Running migrations..."
	go run cmd/migrate/main.go up

.PHONY: migrate-down
migrate-down:
	@echo "⏪ Rolling back last migration..."
	go run cmd/migrate/main.go down

.PHONY: migrate-reset
migrate-reset:
	@echo "⚠️  Rolling back all migrations..."
	go run cmd/migrate/main.go reset

.PHONY: migrate-status
migrate-status:
	@echo "📋 Checking migration status..."
	go run cmd/migrate/main.go status

.PHONY: migrate-to
migrate-to:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION parameter required"; \
		echo "Usage: make migrate-to VERSION=202601230001_init_schema"; \
		exit 1; \
	fi
	@echo "🎯 Migrating to version $(VERSION)..."
	go run cmd/migrate/main.go to $(VERSION)

.PHONY: migrate-up-docker
migrate-up-docker:
	@echo "🚀 Running migrations in Docker..."
	docker exec $(APP_CONTAINER) go run -mod=mod /app/cmd/migrate/main.go up

.PHONY: migrate-down-docker
migrate-down-docker:
	@echo "⏪ Rolling back last migration in Docker..."
	docker exec $(APP_CONTAINER) go run -mod=mod /app/cmd/migrate/main.go down

.PHONY: migrate-status-docker
migrate-status-docker:
	@echo "📋 Checking migration status in Docker..."
	docker exec $(APP_CONTAINER) go run -mod=mod /app/cmd/migrate/main.go status

.PHONY: seed
seed:
	@echo "🌱 Seeding database..."
	go run -mod=mod cmd/seed/main.go

.PHONY: seed-docker
seed-docker:
	@echo "🌱 Seeding database in Docker..."
	docker exec $(APP_CONTAINER) go run -mod=mod /app/cmd/seed/main.go

# ==============================================================================
# HELM
# ==============================================================================

.PHONY: helm-install
helm-install:
	@echo "📦 Installing Helm chart..."
	helm install savvy-system deploy/helm/savvy-system

.PHONY: helm-upgrade
helm-upgrade:
	@echo "🔄 Upgrading Helm release..."
	helm upgrade savvy-system deploy/helm/savvy-system

.PHONY: helm-uninstall
helm-uninstall:
	@echo "🗑️  Uninstalling Helm release..."
	helm uninstall savvy-system

.PHONY: helm-template
helm-template:
	@echo "🔍 Rendering Helm templates..."
	helm template savvy-system deploy/helm/savvy-system

.PHONY: helm-lint
helm-lint:
	@echo "🔍 Linting Helm chart..."
	helm lint deploy/helm/savvy-system

# ==============================================================================
# KUSTOMIZE
# ==============================================================================

.PHONY: kustomize-dev
kustomize-dev:
	@echo "🚀 Deploying to development..."
	kubectl apply -k deploy/kustomize/overlays/development

.PHONY: kustomize-staging
kustomize-staging:
	@echo "🚀 Deploying to staging..."
	kubectl apply -k deploy/kustomize/overlays/staging

.PHONY: kustomize-prod
kustomize-prod:
	@echo "🚀 Deploying to production..."
	kubectl apply -k deploy/kustomize/overlays/production

.PHONY: kustomize-preview-dev
kustomize-preview-dev:
	@echo "🔍 Previewing development manifests..."
	kubectl kustomize deploy/kustomize/overlays/development

.PHONY: kustomize-preview-staging
kustomize-preview-staging:
	@echo "🔍 Previewing staging manifests..."
	kubectl kustomize deploy/kustomize/overlays/staging

.PHONY: kustomize-preview-prod
kustomize-preview-prod:
	@echo "🔍 Previewing production manifests..."
	kubectl kustomize deploy/kustomize/overlays/production

# ==============================================================================
# QUALITY & MAINTENANCE
# ==============================================================================

.PHONY: lint
lint:
	@echo "🔍 Running golangci-lint..."
	golangci-lint run

.PHONY: fmt
fmt:
	@echo "✨ Formatting Go code..."
	go fmt ./...

.PHONY: deps
deps:
	@echo "📦 Updating dependencies..."
	go mod tidy && go mod download
