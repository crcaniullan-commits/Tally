# Build stage
FROM golang:1.27-alpine AS builder

WORKDIR /build

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY cmd ./cmd
COPY internal ./internal
COPY docs ./docs
COPY migrations ./migrations

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -tags 'postgres' -o /migrate github.com/golang-migrate/migrate/v4/cmd/migrate

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 app

WORKDIR /app

COPY --from=builder /server /app/server
COPY --from=builder /migrate /app/migrate
COPY --from=builder /build/migrations /app/migrations
COPY docker/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh

USER app

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]