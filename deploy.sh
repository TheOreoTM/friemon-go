#!/bin/bash

echo "🚀 Starting Friemon deployment..."

# Function to check if command succeeded
check_status() {
    if [ $? -eq 0 ]; then
        echo "✅ $1 successful"
    else
        echo "❌ $1 failed"
        exit 1
    fi
}

# Check if migrate CLI is installed
if ! command -v migrate &> /dev/null; then
    echo "📦 Installing migrate CLI tool..."
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
    sudo mv migrate /usr/local/bin/
    rm -f README.md LICENSE
    check_status "Installing migrate CLI"
fi

# Pull latest changes
echo "📥 Pulling latest changes from git..."
git pull
check_status "Git pull"

# Generate SQLC code
echo "📝 Generating SQLC code..."
sqlc generate
check_status "SQLC generation"

# Stop running containers
echo "🛑 Stopping running containers..."
docker-compose down
check_status "Stopping containers"

# Build fresh images
echo "🏗️ Building fresh images..."
docker-compose build --no-cache
check_status "Building images"

# Start containers in background
echo "🚀 Starting containers..."
docker-compose up -d
check_status "Starting containers"

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
max_attempts=30
attempt=1
while ! docker-compose exec postgres pg_isready -U friemon -d friemon >/dev/null 2>&1; do
    if [ $attempt -eq $max_attempts ]; then
        echo "❌ PostgreSQL failed to become ready in time"
        exit 1
    fi
    echo "Waiting for PostgreSQL... (attempt $attempt/$max_attempts)"
    sleep 2
    attempt=$((attempt + 1))
done
echo "✅ PostgreSQL is ready"

# Run database migrations
echo "🔄 Running database migrations..."
cd friemon && migrate \
    -database "postgres://friemon:friemonpass@postgres:5431/friemon?sslmode=disable" \
    -path db/migrations up
check_status "Database migrations"
cd ..

# Show recent logs
echo "📋 Showing recent logs..."
echo "--------------------"
docker-compose logs --tail=50 bot
echo "--------------------"

echo "✨ Deployment complete!"
echo "📝 To view logs in real-time, run: docker-compose logs -f bot"
echo "🔄 To restart the bot only, run: docker-compose restart bot"
echo "⏹️ To stop everything, run: docker-compose down" 