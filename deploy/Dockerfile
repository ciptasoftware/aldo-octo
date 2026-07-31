# ============================================================
# Go-AI Framework — Production Dockerfile (Multi-Stage Build)
# ============================================================
# Binary size: ~8-13 MB. Final image: ~15 MB (Alpine-based).
# Templates and static assets are embedded via go:embed.
# ============================================================

# --- Stage 1: Build ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies (if any C deps needed in future)
RUN apk add --no-cache ca-certificates

# Copy mod files first for Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build optimized binary
# CGO_ENABLED=0: modernc.org/sqlite is pure Go, no CGO needed
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o /app/server main.go

# --- Stage 2: Production ---
FROM alpine:3.21

# Install CA certificates for HTTPS requests
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S goai && adduser -S goai -G goai

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Set ownership to non-root user
RUN chown -R goai:goai /app

# Switch to non-root user
USER goai

# Expose port (configurable via APP_PORT env var)
EXPOSE 8080

# Health check — uses Core-level /healthz endpoint (always available)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:${APP_PORT:-8080}/healthz || exit 1

# Production defaults
ENV APP_ENV=production

CMD ["./server"]
