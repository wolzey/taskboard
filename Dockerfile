# Build stage
FROM golang:1.25.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o taskboard ./cmd/taskboard

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections (needed for Redis TLS)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/taskboard .

# Copy example config (optional reference)
COPY --from=builder /build/config.example.yaml .

# Create directory for optional config file
RUN mkdir -p /app/config

# Expose the API port
EXPOSE 1337

# Run the application
CMD ["./taskboard", "serve"]
