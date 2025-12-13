#!/bin/bash
# Helper script to run the server in production-like mode (no seed data)

set -e

echo "🏭 Starting server in production-like mode..."
echo "   - APP_ENV=production (seed data will NOT run)"
echo "   - Connects to Firebase and Cloudinary"
echo ""

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
    exit 1
fi

echo "✅ Found Firebase project: $PROJECT_ID"

# Ensure configs directory exists
echo "📁 Checking configs directory..."
if [ ! -d "configs" ]; then
    echo "❌ Error: configs directory missing"
    exit 1
fi

# No need to copy service account as we expect it in configs/

# Check for Cloudinary credentials
if [ -z "$CLOUDINARY_CLOUD_NAME" ] || [ -z "$CLOUDINARY_API_KEY" ] || [ -z "$CLOUDINARY_API_SECRET" ]; then
    echo "⚠️  Warning: Cloudinary credentials not set in environment"
    echo "   Set these variables or add them to configs/.env.production:"
    echo "   - CLOUDINARY_CLOUD_NAME"
    echo "   - CLOUDINARY_API_KEY"
    echo "   - CLOUDINARY_API_SECRET"
    echo ""
    echo "   Continuing anyway (credentials may be in .env file)..."
fi

# Create or update docker-compose.override.yml with project ID
cat > docker-compose.override.yml << EOF
version: '3.8'

services:
  server:
    environment:
      - FIREBASE_PROJECT_ID=$PROJECT_ID
EOF

echo "✅ Created docker-compose.override.yml with project ID"

# Rebuild if needed, then start
echo "📦 Building/updating containers..."
docker compose -f docker-compose.prod.yml build

echo "🚀 Starting services in production mode..."
echo "   Seed data will NOT run (APP_ENV=production)"
echo ""
docker compose -f docker-compose.prod.yml up -d




