# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.24 AS builder

WORKDIR /app

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically-linked binary
# CGO_ENABLED=1 is required for go-sqlite3
COPY . .
RUN CGO_ENABLED=1 GOOS=linux \
    go build -ldflags="-s -w" -o /server ./cmd/server

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM debian:bookworm-slim

# ca-certificates is needed for any outbound HTTPS calls
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Binary
COPY --from=builder /server ./server

# Migrations — must be present at runtime for ApplyMigrations()
COPY pkg/db/migrations/sqlite ./pkg/db/migrations/sqlite

# Pre-create upload directories so the server can write on first run
RUN mkdir -p uploads/avatars uploads/posts uploads/comments

# ── Environment defaults ──────────────────────────────────────────────────────
# DB_PATH points into the mounted data volume (see docker-compose.yml)
ENV SERVER_ADDR=:8080
ENV DB_PATH=/app/data/social-network.db
ENV MIGRATIONS_DIR=/app/pkg/db/migrations/sqlite

EXPOSE 8080

ENTRYPOINT ["./server"]
