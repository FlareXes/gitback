# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/gitback ./cmd/gitback

# Final stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache \
    git \
    openssh-client

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/bin/gitback /app/

# Create data directory with correct permissions
RUN mkdir -p /data && \
    chown -R appuser:appgroup /data

# Switch to non-root user
USER appuser

# Set volume for persistent data
VOLUME ["/data"]

# Set environment variables with defaults
ENV GITBACK_OUTPUT_DIR=/data \
    GITBACK_THREADS=5 \
    GITBACK_TIMEOUT=30 \
    GITBACK_INCLUDE_GISTS=true

# Set the entrypoint
ENTRYPOINT ["/app/gitback"]

# Default command (can be overridden)
CMD ["--help"]
