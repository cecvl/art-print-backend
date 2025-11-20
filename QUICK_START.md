# Quick Start Guide - Running Docker Containers

## Problem: Containers Exit Immediately

If `docker ps` shows empty after `docker-compose up`, the containers likely exited due to errors.

## Solution: Check Logs First

```bash
# Check what happened
sudo docker-compose logs server

# Or check all containers
sudo docker-compose logs

# Check exited containers
sudo docker ps -a
```

## Option 1: Use Firebase Emulator (Easiest for Development)

This is the recommended approach for local development - no Firebase credentials needed!

```bash
# Start server with Firebase emulator
sudo docker-compose -f docker-compose.dev.yml up

# This will:
# 1. Start Firebase emulator
# 2. Wait for it to be ready
# 3. Start the server connected to the emulator
```

The server will be available at: `http://localhost:3001`

## Option 2: Run Server Only (If You Have Firebase Credentials)

```bash
# Make sure firebase-service-account.json exists
ls -la firebase-service-account.json

# Start server in detached mode
sudo docker-compose up -d server

# Check if it's running
sudo docker ps

# View logs
sudo docker-compose logs -f server
```

## Option 3: Run Server Without Firebase (For Testing)

If you just want to test the container build, you can modify the server to skip Firebase initialization temporarily, or use the emulator approach above.

## Verify It's Working

```bash
# Check running containers
sudo docker ps

# Test health endpoint
curl http://localhost:3001/health

# Should return: {"status":"ok","service":"art-print-backend"}
```

## Common Commands

```bash
# Start in foreground (see logs)
sudo docker-compose -f docker-compose.dev.yml up

# Start in background
sudo docker-compose -f docker-compose.dev.yml up -d

# Stop containers
sudo docker-compose down

# Rebuild and start
sudo docker-compose -f docker-compose.dev.yml up --build

# View logs
sudo docker-compose logs -f server
```

## If Still Having Issues

1. **Check the logs**:
   ```bash
   sudo docker-compose logs server | tail -50
   ```

2. **Rebuild containers**:
   ```bash
   sudo docker-compose down
   sudo docker-compose build --no-cache
   sudo docker-compose -f docker-compose.dev.yml up
   ```

3. **Check Docker is running**:
   ```bash
   sudo systemctl status docker
   ```

