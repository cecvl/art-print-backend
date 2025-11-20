# Docker Troubleshooting Guide

## Issue: Containers Exit Immediately After Starting

If `docker ps` shows no running containers after `docker-compose up`, the containers likely exited due to errors.

### Check Container Logs

```bash
# Check all container logs
sudo docker-compose logs

# Check specific service logs
sudo docker-compose logs server
sudo docker-compose logs seed

# Check exited containers
sudo docker ps -a

# View logs of a specific container
sudo docker logs art-print-server
sudo docker logs art-print-seed
```

### Common Issues and Solutions

#### 1. Missing Firebase Credentials

**Error**: `GOOGLE_APPLICATION_CREDENTIALS not set` or file not found

**Solution**: You have two options:

**Option A: Use Firebase Emulator (Recommended for Development)**
```bash
# Use the dev compose file with emulator
sudo docker-compose -f docker-compose.dev.yml up
```

**Option B: Provide Firebase Credentials**
```bash
# Ensure firebase-service-account.json exists
ls -la firebase-service-account.json

# If missing, create it or set environment variables
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/firebase-service-account.json
```

#### 2. Missing Config Files

**Error**: Config files not found

**Solution**: Create the configs directory and .env file:
```bash
mkdir -p configs
# Create configs/.env.dev with your environment variables
```

#### 3. Port Already in Use

**Error**: `bind: address already in use`

**Solution**: 
```bash
# Check what's using port 3001
sudo lsof -i :3001
# Or
sudo netstat -tulpn | grep 3001

# Stop the process or change the port in docker-compose.yml
```

#### 4. Container Build Failed

**Error**: Build errors during `docker-compose up`

**Solution**:
```bash
# Rebuild without cache
sudo docker-compose build --no-cache

# Check build logs
sudo docker-compose build server
```

### Running Containers Properly

#### Development with Firebase Emulator (Easiest)
```bash
# This starts Firebase emulator and server together
sudo docker-compose -f docker-compose.dev.yml up

# In another terminal, run seed if needed
sudo docker-compose -f docker-compose.dev.yml run --rm seed
```

#### Development with Real Firebase
```bash
# Ensure firebase-service-account.json exists
# Then run:
sudo docker-compose up -d server

# Check if it's running
sudo docker ps

# View logs
sudo docker-compose logs -f server
```

#### Run Server Only (No Seed)
```bash
# Start server in detached mode
sudo docker-compose up -d server

# Check status
sudo docker-compose ps

# View logs
sudo docker-compose logs -f server

# Stop
sudo docker-compose down
```

#### Run Seed Job Separately
```bash
# Run seed once (it will exit after completion)
sudo docker-compose --profile seed run --rm seed

# Check seed logs
sudo docker-compose logs seed
```

### Verify Container is Running

```bash
# List running containers
sudo docker ps

# Should show something like:
# CONTAINER ID   IMAGE                    STATUS          PORTS                    NAMES
# abc123def456   art-print-backend-server   Up 2 minutes   0.0.0.0:3001->3001/tcp   art-print-server

# Test the health endpoint
curl http://localhost:3001/health

# Should return:
# {"status":"ok","service":"art-print-backend"}
```

### Debugging Steps

1. **Check if containers were created**:
   ```bash
   sudo docker ps -a
   ```

2. **Check exit codes**:
   ```bash
   sudo docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.ExitCode}}"
   ```

3. **Run container interactively**:
   ```bash
   sudo docker run -it --rm art-print-server:latest sh
   ```

4. **Check environment variables**:
   ```bash
   sudo docker exec art-print-server env
   ```

5. **Check mounted volumes**:
   ```bash
   sudo docker exec art-print-server ls -la /app/configs
   ```

### Quick Fix: Start with Emulator

The easiest way to get started is using the Firebase emulator:

```bash
# 1. Ensure you have a firebase.json file (or create a minimal one)
# 2. Start with emulator
sudo docker-compose -f docker-compose.dev.yml up

# 3. In another terminal, test the server
curl http://localhost:3001/health
```

### Still Having Issues?

1. Check Docker daemon is running:
   ```bash
   sudo systemctl status docker
   ```

2. Check Docker Compose version:
   ```bash
   sudo docker-compose --version
   ```

3. Try rebuilding everything:
   ```bash
   sudo docker-compose down -v
   sudo docker-compose build --no-cache
   sudo docker-compose up
   ```

