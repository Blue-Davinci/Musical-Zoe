# Build stage
FROM golang:1.24.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o musical-zoe ./cmd/api

# Final stage
FROM alpine:3.20

# Install ca-certificates for HTTPS requests and security updates
RUN apk --no-cache add ca-certificates tzdata && \
    apk upgrade --no-cache

# Create non-root user
RUN adduser -D -s /bin/sh musical-zoe

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/musical-zoe .

# Copy migration files (if needed for auto-migration)
COPY --from=builder /app/internal/sql/schema ./internal/sql/schema

# Change ownership
RUN chown -R musical-zoe:musical-zoe /app

# Switch to non-root user
USER musical-zoe

# Expose port
EXPOSE 4000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:4000/v1/health || exit 1

# Run the application
CMD ["./musical-zoe"]
