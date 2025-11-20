#!/bin/bash
# Setup script to extract and set Firebase project ID
# Run this before docker-compose commands: source setup-env.sh

if [ ! -f "firebase-service-account.json" ]; then
    echo "❌ Error: firebase-service-account.json not found"
    return 1 2>/dev/null || exit 1
fi

# Extract project ID
PROJECT_ID=$(grep -o '"project_id": "[^"]*' firebase-service-account.json | cut -d'"' -f4)

if [ -z "$PROJECT_ID" ]; then
    echo "❌ Error: Could not extract project_id from firebase-service-account.json"
    return 1 2>/dev/null || exit 1
fi

export FIREBASE_PROJECT_ID=$PROJECT_ID
echo "✅ FIREBASE_PROJECT_ID set to: $PROJECT_ID"
echo "   You can now run: sudo docker-compose up server"

