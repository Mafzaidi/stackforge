# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /myapp ./cmd/api

# Optional: Compress the binary
RUN apk add --no-cache upx && upx /myapp

# Final stage
FROM alpine:latest

# Security: Run as non-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

COPY --from=builder /myapp /myapp

# Runtime configuration
ENV PORT=4000
EXPOSE $PORT

HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:$PORT/health || exit 1

ENTRYPOINT ["/myapp"]