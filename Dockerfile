# ==============================================================================
# DEVELOPMENT STAGE (Hot Reload)
# ==============================================================================
FROM golang:1.25-alpine@sha256:98e6cffc31ccc44c7c15d83df1d69891efee8115a5bb7ede2bf30a38af3e3c92 AS development

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
# PRODUCTION STAGE (downloads pre-built binary from GitHub Release)
# ==============================================================================
# Downloads the appropriate binary based on TARGETARCH from GitHub Releases.
# Usage: docker build --target production --build-arg VERSION=v1.2.3 --platform linux/amd64
FROM alpine:3.23@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659 AS builder
ARG TARGETARCH
ARG VERSION
RUN apk add --no-cache curl tar ca-certificates && \
    ARCH=$(case ${TARGETARCH} in amd64) echo "x86_64" ;; arm64) echo "arm64" ;; *) echo ${TARGETARCH} ;; esac) && \
    curl -fsSL "https://github.com/sbaerlocher/savvy/releases/download/${VERSION}/savvy_Linux_${ARCH}.tar.gz" -o /tmp/savvy.tar.gz && \
    tar -xzf /tmp/savvy.tar.gz -C /tmp savvy && \
    chmod +x /tmp/savvy

FROM scratch AS production

LABEL org.opencontainers.image.title="savvy"
LABEL org.opencontainers.image.description="Digital customer card, voucher and gift card management system with sharing functionality"
LABEL org.opencontainers.image.source="https://github.com/sbaerlocher/savvy"
LABEL org.opencontainers.image.licenses="MIT"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /tmp/savvy /savvy
USER 65534:65534
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/savvy", "-health"]
ENTRYPOINT ["/savvy"]
