#!/bin/bash

echo "🚀 Starting Friemon deployment..."

check_status() {
    if [ $? -eq 0 ]; then
        echo "✅ $1 successful"
    else
        echo "❌ $1 failed"
        exit 1
    fi
}

echo "📥 Pulling latest changes from git..."
git pull
check_status "Git pull"

echo "🛑 Stopping running containers..."
docker-compose down
check_status "Stopping containers"

echo "🏗️ Building fresh images..."
docker-compose build --no-cache
check_status "Building images"

echo "🚀 Starting containers..."
docker-compose up -d
check_status "Starting containers"

echo "⏳ Waiting for PostgreSQL to be ready..."
max_attempts=30
attempt=1

while ! docker-compose exec -T postgres pg_isready -U friemon -d friemon >/dev/null 2>&1; do
    if [ $attempt -eq $max_attempts ]; then
        echo "❌ PostgreSQL failed to become ready in time"
        exit 1
    fi
    echo "Waiting for PostgreSQL... (attempt $attempt/$max_attempts)"
    sleep 2
    attempt=$((attempt + 1))
done

echo "✅ PostgreSQL is ready"
echo "✅ GORM will handle database migrations automatically"

echo "📋 Showing recent logs..."
echo "--------------------"
docker-compose logs --tail=50 bot
echo "--------------------"

echo "✨ Deployment complete!"
echo "📝 To view logs in real-time, run: docker-compose logs -f bot"
echo "🔄 To restart the bot only, run: docker-compose restart bot"
echo "⏹️ To stop everything, run: docker-compose down"