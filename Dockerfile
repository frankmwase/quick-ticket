# ── Build Stage ────────────────────────────────────────────────────────────────
FROM golang:alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/quick-ticket ./cmd/baas-core

# ── Runtime Stage ─────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/quick-ticket /bin/quick-ticket

EXPOSE 8080

ENTRYPOINT ["/bin/quick-ticket"]
