#!/bin/bash
# Helper script to run the Docker server container
# This script extracts the project ID from firebase-service-account.json

set -e

# Check if firebase-service-account.json exists in configs/
if [ ! -f "configs/firebase-service-account.json" ]; then
    echo "❌ Error: configs/firebase-service-account.json not found"
    echo "   Please place your service account file in the configs/ directory"
    exit 1
fi

# Extract project ID from configs/firebase-service-account.json
PROJECT_ID=$(grep -o '"project_id": "[^"]*' configs/firebase-service-account.json | cut -d'"' -f4)

if [ -z "$PROJECT_ID" ]; then
    echo "❌ Error: Could not extract project_id from configs/firebase-service-account.json"
    echo "   Please check that the file is valid JSON and contains a 'project_id' field"
    exit 1
fi

echo "✅ Found Firebase project: $PROJECT_ID"
echo "🚀 Starting server container..."

# Ensure configs directory exists
echo "📁 Checking configs directory..."
if [ ! -d "configs" ]; then
    echo "❌ Error: configs directory missing"
    exit 1
fi

# No need to copy service account as we expect it in configs/

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
docker compose build server

echo "🚀 Starting server..."
docker compose up server

