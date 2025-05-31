FROM golang:1.23-alpine AS builder

# Install git and file utility
RUN apk add --no-cache git file

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Show build environment info
RUN echo "=== BUILD ENVIRONMENT ===" && \
    go env GOOS GOARCH && \
    uname -m && \
    echo "========================="

# Build for the target platform
ARG COMMIT=unknown
ARG BRANCH=unknown
RUN go build \
    -ldflags="-X main.commit=${COMMIT} -X main.branch=${BRANCH}" \
    -o friemon \
    ./cmd/friemon/main.go

# Check what we built
RUN echo "=== BUILT BINARY INFO ===" && \
    ls -la friemon && \
    file friemon && \
    echo "========================="

# Runtime stage
FROM alpine:3.20

# Install utilities for debugging
RUN apk --no-cache add ca-certificates file

WORKDIR /app

# Copy binary
COPY --from=builder /app/friemon .
COPY ./assets ./assets

# Show runtime environment info
RUN echo "=== RUNTIME ENVIRONMENT ===" && \
    uname -m && \
    file friemon && \
    echo "============================"

# Make executable
RUN chmod +x friemon

# Create user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup && \
    chown -R appuser:appgroup /app

USER appuser

# Try to run the binary and show any errors
CMD echo "=== ATTEMPTING TO RUN ===" && \
    echo "Container arch: $(uname -m)" && \
    echo "Binary info: $(file ./friemon)" && \
    echo "=========================" && \
    ./friemon