# ---------------------------------------------------------
# Build Stage
# ---------------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies and CA certs
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o /build/bin/api ./cmd/api

# ---------------------------------------------------------
# Production Runtime Stage
# ---------------------------------------------------------
FROM alpine:3.20

# Security: Create non-root user and group (UID/GID 10001)
RUN addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup

WORKDIR /app

# Install runtime utilities (curl for health check probe, certs, timezone)
RUN apk add --no-cache ca-certificates tzdata curl

# Copy binary from builder
COPY --from=builder /build/bin/api /app/api
COPY --from=builder /build/migrations /app/migrations

# Assign ownership to non-root user
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

# Container health probe
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health/live || exit 1

ENTRYPOINT ["/app/api"]
