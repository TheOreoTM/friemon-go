#!/bin/bash

echo "Updating Friemon Bot..."

# Pull latest changes
echo "Pulling latest changes from GitHub..."
git pull

# Rebuild and restart the bot
echo "Rebuilding bot container..."
docker-compose build bot

echo "Restarting bot..."
docker-compose up -d --no-deps bot

echo "Showing recent logs..."
docker-compose logs --tail=50 bot

echo "Update complete! Use 'docker-compose logs -f bot' to follow logs" 