# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build optimized binary for slim image
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w" \
  -trimpath \
  -o api-slim ./api

# Build debug binary (no optimization flags)
RUN CGO_ENABLED=0 GOOS=linux go build \
  -o api-debug ./api

# Slim image (distroless with optimized binary)
FROM gcr.io/distroless/static-debian12:nonroot AS slim
COPY --from=builder /app/api-slim /api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]

# Debug image (alpine with unoptimized binary)
FROM alpine:latest AS debug
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/api-debug ./api
EXPOSE 8080
CMD ["./api"]