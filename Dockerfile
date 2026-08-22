# ==============================================================================
# Savvy - Multi-Stage Dockerfile (Optimized)
# ==============================================================================
# Targets:
#   - backend-dev: Go API with Air hot reload
#   - frontend-dev: Vite dev server
#   - production-build: Production image with embedded frontend
#   - production: Minimal image with GitHub release binary
# ==============================================================================

# ==============================================================================
# BASE STAGE - Shared dependencies for dev targets
# ==============================================================================
FROM golang:1.26-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS base-dev

# Install build tools and utilities
RUN apk add --no-cache \
    git \
    build-base \
    curl \
    ca-certificates

WORKDIR /app

# ==============================================================================
# BACKEND-DEV STAGE (Go API with Air Hot Reload)
# ==============================================================================
FROM base-dev AS backend-dev

# Install Air for hot reload
RUN go install github.com/air-verse/air@v1.61.1

# Copy Go dependencies (layer caching optimization)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy Air configuration
COPY .air.toml ./

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    mkdir -p /app/tmp /home/appuser/.cache && \
    chown -R appuser:appuser /app /go /home/appuser

USER appuser

EXPOSE 8080

ENV VERSION=dev \
    GO_ENV=development \
    CGO_ENABLED=1 \
    PORT=8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["/go/bin/air", "-c", ".air.toml"]

# ==============================================================================
# FRONTEND-DEV STAGE (Vite Dev Server with HMR)
# ==============================================================================
FROM node:24.19.0-alpine@sha256:d32cdf619f63fe0471182d08996dd516c6275bb5fd31ae06e55a570bd9e1ad43 AS frontend-dev

# Install wget for healthcheck and su-exec to drop the dev server to the dde
# user (dde's overlay forces the container to start as root; see CMD below)
RUN apk add --no-cache wget su-exec

WORKDIR /app/client

# Install dependencies (cached layer)
COPY client/package.json client/package-lock.json ./
RUN npm ci --loglevel=error

# Copy source (watch syncs changes at runtime)
COPY client/ ./

RUN chown -R node:node /app

USER node

EXPOSE 5173

ENV NODE_ENV=development

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5173 || exit 1

# dde's runtime overlay forces the container to start as root (user 0:0) and
# overrides the entrypoint, but preserves this CMD. Drop the dev server to the
# dde user via su-exec so it owns the Vite caches it creates in node_modules and
# `dde project:exec` (same uid) can write them. sh -c defers $DDE_UID/$DDE_GID
# expansion to runtime, where dde injects the host uid/gid into the container.
CMD ["sh", "-c", "exec su-exec \"$DDE_UID:$DDE_GID\" npm run dev -- --host 0.0.0.0"]

# ==============================================================================
# FRONTEND BUILDER STAGE (Build SvelteKit)
# ==============================================================================
FROM node:24.19.0-alpine@sha256:d32cdf619f63fe0471182d08996dd516c6275bb5fd31ae06e55a570bd9e1ad43 AS frontend-builder

WORKDIR /app/client

# Copy package files (layer caching optimization)
COPY client/package.json client/package-lock.json ./

# Install dependencies and clean existing build artifacts
RUN npm ci --loglevel=error --prefer-offline && \
    rm -rf build/ .svelte-kit/ node_modules/.vite/

# Copy frontend source
COPY client/ ./

# Build SvelteKit app (output: build/)
# This matches the local build process (vite build)
RUN npm run build

# Verify build output exists
RUN ls -la build/ && echo "✅ Frontend build complete"

# ==============================================================================
# GO BUILDER STAGE (Build Go Binaries)
# ==============================================================================
FROM golang:1.26-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS go-builder

ARG VERSION=dev

# Install build dependencies
RUN apk add --no-cache git build-base ca-certificates

WORKDIR /app

# Copy Go dependencies (layer caching optimization)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Clean any existing embedded assets to ensure fresh build
RUN rm -rf ./internal/assets/client

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/client/build ./internal/assets/client

# Verify frontend assets were copied successfully
RUN ls -la ./internal/assets/client && \
    echo "✅ Frontend assets embedded successfully"

# Copy Go source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build Go binaries with embedded frontend
# CGO_ENABLED=0 for static binary (smaller, more portable)
# -tags production enables embedded frontend assets
RUN CGO_ENABLED=0 GOOS=linux go build \
    -tags production \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -trimpath \
    -o /app/savvy \
    ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/seed \
    ./cmd/seed && \
    CGO_ENABLED=0 GOOS=linux go build \
    -tags production \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/e2e \
    ./cmd/e2e

# ==============================================================================
# PRODUCTION BUILD STAGE (Distroless Production Image)
# ==============================================================================
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab AS production-build

LABEL org.opencontainers.image.title="savvy"
LABEL org.opencontainers.image.description="Digital customer card, voucher and gift card management system with sharing functionality"
LABEL org.opencontainers.image.source="https://github.com/sbaerlocher/savvy"
LABEL org.opencontainers.image.licenses="MIT"

WORKDIR /app

# Copy binaries from builder
COPY --from=go-builder --chown=nonroot:nonroot /app/savvy /app/savvy
COPY --from=go-builder --chown=nonroot:nonroot /app/seed /app/seed
COPY --from=go-builder --chown=nonroot:nonroot /app/e2e /app/e2e

# Use nonroot user (uid:65532)
USER nonroot:nonroot

EXPOSE 3000

ENV GO_ENV=production \
    SERVER_PORT=3000

# Note: Distroless doesn't support shell-based healthchecks
# Use Docker Compose or Kubernetes liveness probes instead

ENTRYPOINT ["/app/savvy"]

# ==============================================================================
# PRODUCTION STAGE (GitHub Release Binary)
# ==============================================================================
FROM alpine:3.24@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b AS release-builder

ARG TARGETARCH
ARG VERSION

RUN apk add --no-cache curl tar ca-certificates && \
    ARCH=$(case ${TARGETARCH} in \
    amd64) echo "x86_64" ;; \
    arm64) echo "arm64" ;; \
    *) echo ${TARGETARCH} ;; \
    esac) && \
    curl -fsSL "https://github.com/sbaerlocher/savvy/releases/download/${VERSION}/savvy_Linux_${ARCH}.tar.gz" -o /tmp/savvy.tar.gz && \
    tar -xzf /tmp/savvy.tar.gz -C /tmp savvy && \
    chmod +x /tmp/savvy

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab AS production

LABEL org.opencontainers.image.title="savvy"
LABEL org.opencontainers.image.description="Digital customer card, voucher and gift card management system with sharing functionality"
LABEL org.opencontainers.image.source="https://github.com/sbaerlocher/savvy"
LABEL org.opencontainers.image.licenses="MIT"

COPY --from=release-builder --chown=nonroot:nonroot /tmp/savvy /savvy

USER nonroot:nonroot

EXPOSE 3000

ENTRYPOINT ["/savvy"]
