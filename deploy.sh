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

# Show recent logs
echo "📋 Showing recent logs..."
echo "--------------------"
docker-compose logs --tail=50 bot
echo "--------------------"

echo "✨ Deployment complete!"
echo "📝 To view logs in real-time, run: docker-compose logs -f bot"
echo "🔄 To restart the bot only, run: docker-compose restart bot"
echo "⏹️ To stop everything, run: docker-compose down" 