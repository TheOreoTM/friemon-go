# ---- Builder Stage ----
# Use a specific version of the golang-alpine image for reproducibility
FROM golang:1.23-alpine AS builder

# Set necessary environment variables for the build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV ASSETS_DIR=/app/assets 

# Install git, which is required to fetch dependencies
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy go module files and download dependencies.
# This is done in a separate layer to leverage Docker's layer caching,
# which speeds up builds when dependencies haven't changed.
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Copy the assets directory
COPY assets /app/assets  

# Build the application, creating a static binary.
# Build arguments can be passed to embed version info into the binary.
ARG COMMIT=unknown
ARG BRANCH=unknown
RUN go build -ldflags="-X main.commit=${COMMIT} -X main.branch=${BRANCH}" -o /friemon ./main.go

# ---- Final Stage ----
# Use a specific version of the alpine image for a small and secure base
FROM alpine:3.20

# Create a non-root user and group for security purposes
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory for the final image
WORKDIR /app

# Copy the built binary and the example config file from the builder stage
COPY --from=builder /friemon /app/friemon
COPY config.example.toml /app/config.toml

# Assign ownership of the application files to the non-root user
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Healthcheck to ensure the bot process is running
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep -x friemon || exit 1

# Command to run the application when the container starts
CMD ["./friemon"]