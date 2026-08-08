# ── Stage 1: build both Go binaries ───────────────────────────────────────────
FROM golang:alpine AS builder

WORKDIR /app

# Copy module files first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build API server binary — CGO_ENABLED=0 for a fully static binary (no libc needed)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./cmd/

# Build worker binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/worker ./cmd/worker/

# ── Stage 2: API server production image ──────────────────────────────────────
FROM alpine:3.21 AS api

# ca-certificates: needed for HTTPS calls to external APIs (Gemini, Groq, etc.)
# tini: proper PID 1 signal handling for graceful shutdown
RUN apk add --no-cache ca-certificates tini

WORKDIR /app

# Create storage directories used by the app at runtime
RUN mkdir -p storage/attachments storage/reports

# Run as non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown -R appuser:appgroup /app
USER appuser

# Copy only the API binary from the builder stage
COPY --from=builder /bin/api .

EXPOSE 8080

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./api"]

# ── Stage 3: background worker production image ────────────────────────────────
FROM alpine:3.21 AS worker_image

RUN apk add --no-cache ca-certificates tini

WORKDIR /app

RUN mkdir -p storage/attachments storage/reports

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown -R appuser:appgroup /app
USER appuser

# Copy only the worker binary from the builder stage
COPY --from=builder /bin/worker .

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./worker"]
