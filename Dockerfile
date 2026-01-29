# ==============================================================================
# DEVELOPMENT STAGE (Hot Reload)
# ==============================================================================
FROM golang:1.25-alpine AS development

# Install development dependencies including Node.js
RUN apk add --no-cache git build-base nodejs npm

# Install Air for hot reload and Templ in single layer
RUN go install github.com/air-verse/air@v1.61.1 && \
    go install github.com/a-h/templ/cmd/templ@v0.3.977

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy package files and install npm dependencies
COPY package.json package-lock.json ./
RUN npm ci --quiet

# Copy Air configuration
COPY .air.toml ./

# Create non-root user for development
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    mkdir -p /app/tmp && \
    chown -R appuser:appuser /app /go

USER appuser

EXPOSE 3000

# Use Air for hot reload
CMD ["/go/bin/air", "-c", ".air.toml"]

# ==============================================================================
# BUILDER STAGE (Production)
# ==============================================================================
FROM golang:1.25-alpine AS builder

# Add build arguments for version info
ARG VERSION=dev
ARG BUILD_TIME=unknown

# Install build dependencies including Node.js and git
RUN apk update && \
    apk add --no-cache git build-base nodejs npm

WORKDIR /app

# Copy git directory for build vcs info
COPY .git .git/

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy package files and install npm dependencies
COPY package.json package-lock.json ./
RUN npm ci --quiet

# Copy source (excluding node_modules via .dockerignore)
COPY . .

# Build frontend bundles
RUN npm run build

# Install Templ and generate templates
RUN go install github.com/a-h/templ/cmd/templ@v0.3.977
RUN /go/bin/templ generate

# Build Go binary with version info and build flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -buildvcs=true \
    -buildmode=default \
    -trimpath \
    -o /savvy ./cmd/server

# ==============================================================================
# PRODUCTION STAGE
# ==============================================================================
FROM alpine:3.21 AS production

# Add OCI metadata labels
LABEL org.opencontainers.image.title="savvy"
LABEL org.opencontainers.image.description="Digital customer card, voucher and gift card management system with sharing functionality"
LABEL org.opencontainers.image.source="https://github.com/sbaerlocher/savvy"
LABEL org.opencontainers.image.licenses="MIT"

# Install runtime dependencies
RUN apk add --no-cache ca-certificates wget~1.25

WORKDIR /app

# Copy binary from builder
COPY --from=builder /savvy ./server

# Copy static files and migrations
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/locales ./locales

# Create non-root user
RUN addgroup -g 65534 appuser && \
    adduser -D -u 65534 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 3000

# Health check using wget
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

CMD ["./server"]
