# Default Dockerfile - builds server by default
# For production, use Dockerfile.server or Dockerfile.seed explicitly
FROM golang:1.25.4-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o /app/server \
    ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/server /app/server
COPY --from=builder /build/configs /app/configs

WORKDIR /app
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/server"]
