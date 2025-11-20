#!/bin/bash
# Helper script to run the Docker server container
# This script extracts the project ID from firebase-service-account.json

set -e

# Check if firebase-service-account.json exists
if [ ! -f "firebase-service-account.json" ]; then
    echo "❌ Error: firebase-service-account.json not found in current directory"
    exit 1
fi

# Extract project ID from firebase-service-account.json
PROJECT_ID=$(grep -o '"project_id": "[^"]*' firebase-service-account.json | cut -d'"' -f4)

if [ -z "$PROJECT_ID" ]; then
    echo "❌ Error: Could not extract project_id from firebase-service-account.json"
    echo "   Please check that the file is valid JSON and contains a 'project_id' field"
    exit 1
fi

echo "✅ Found Firebase project: $PROJECT_ID"
echo "🚀 Starting server container..."

# Ensure configs directory and symlink exist
echo "📁 Setting up configs directory..."
./setup-configs.sh

# Create or update docker-compose.override.yml with project ID
cat > docker-compose.override.yml << EOF
version: '3.8'

services:
  server:
    environment:
      - FIREBASE_PROJECT_ID=$PROJECT_ID
  seed:
    environment:
      - FIREBASE_PROJECT_ID=$PROJECT_ID
EOF

echo "✅ Created docker-compose.override.yml with project ID"

# Rebuild if needed, then start
echo "📦 Building/updating container..."
sudo docker-compose build server

echo "🚀 Starting server..."
sudo docker-compose up server

