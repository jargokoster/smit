# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smit-api ./

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/smit-api .

# Create data directory
RUN mkdir -p /app/data && chown -R appuser:appuser /app

# Copy data file (optional, can be mounted as volume)
COPY --chown=appuser:appuser data/data.json /app/data/

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 1234

# Set environment variables
ENV SERVER_PORT=1234
ENV DATA_FILE_PATH=/app/data/data.json

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:1234/health || exit 1

# Run the application
CMD ["./smit-api"]