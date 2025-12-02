#!/bin/bash
# Cross-platform shell script for running sync in Docker

echo "🐳 Running LeetCode Sync in Docker..."
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found"
    echo "Please copy .env.example to .env and configure it"
    exit 1
fi

# Run docker-compose
docker-compose up --build

echo ""
echo "✅ Sync complete!"
