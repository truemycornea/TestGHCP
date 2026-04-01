# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build tools (required for CGO-free compilation of some deps).
RUN apk add --no-cache git ca-certificates tzdata

# Cache dependency download separately from source compilation.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a static binary with no CGO dependencies.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /app/aura-backend \
    ./cmd/api

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/aura-backend /aura-backend

EXPOSE 8080

ENTRYPOINT ["/aura-backend"]
