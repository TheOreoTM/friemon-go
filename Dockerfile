FROM golang:1.23-alpine AS builder

# Install git
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build for the target platform (let Docker handle architecture)
ARG COMMIT=unknown
ARG BRANCH=unknown
RUN go build \
    -ldflags="-X main.commit=${COMMIT} -X main.branch=${BRANCH}" \
    -o friemon \
    ./cmd/friemon/main.go

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates
WORKDIR /app

# Copy binary
COPY --from=builder /app/friemon .
COPY config.example.toml config.toml
COPY ./assets ./assets

# Make executable
RUN chmod +x friemon

# Create user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup && \
    chown -R appuser:appgroup /app

USER appuser

CMD ["./friemon"]