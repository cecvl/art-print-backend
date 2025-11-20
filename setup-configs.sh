#!/bin/bash
# Setup script to create symlink for firebase-service-account.json in configs directory
# This allows the file to be accessible when mounting the configs directory

set -e

# Ensure configs directory exists
mkdir -p configs

# Create symlink if it doesn't exist or is broken
if [ ! -f "configs/firebase-service-account.json" ] || [ ! -L "configs/firebase-service-account.json" ]; then
    if [ -f "firebase-service-account.json" ]; then
        echo "🔗 Creating symlink: configs/firebase-service-account.json -> ../firebase-service-account.json"
        ln -sf ../firebase-service-account.json configs/firebase-service-account.json
        echo "✅ Symlink created successfully"
    else
        echo "❌ Error: firebase-service-account.json not found in project root"
        exit 1
    fi
else
    echo "✅ Symlink already exists: configs/firebase-service-account.json"
fi

# Verify the symlink works
if [ -f "configs/firebase-service-account.json" ]; then
    echo "✅ Firebase credentials file is accessible in configs directory"
else
    echo "⚠️  Warning: Symlink created but file not accessible"
fi

