# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o friemon

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create config directory
RUN mkdir -p /app/config

# Copy the binary from builder
COPY --from=builder /app/friemon .

# Copy default config
COPY config.toml /app/config/config.toml

# Run the application
CMD ["./friemon", "--config", "/app/config/config.toml"] 