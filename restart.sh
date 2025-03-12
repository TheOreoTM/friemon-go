#!/bin/bash

echo "🔄 Restarting Friemon bot..."
docker-compose restart bot

echo "📋 Recent logs:"
echo "--------------------"
docker-compose logs --tail=20 bot
echo "--------------------"

echo "✨ Restart complete!"
echo "📝 To view logs in real-time, run: docker-compose logs -f bot" 