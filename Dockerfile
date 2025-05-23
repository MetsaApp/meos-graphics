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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o meos-graphics ./cmd/meos-graphics

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S meos && \
    adduser -u 1000 -S meos -G meos

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/meos-graphics .

# Create logs directory
RUN mkdir -p /app/logs && chown -R meos:meos /app

# Switch to non-root user
USER meos

# Expose port
EXPOSE 8090

# Run the application
ENTRYPOINT ["./meos-graphics"]