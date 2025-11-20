# How to Run Docker Containers

Since your `firebase-service-account.json` is in the root directory, here's how to run the containers:

## Step 1: Set Your Firebase Project ID

You need to set the `FIREBASE_PROJECT_ID` environment variable. You can find this in your Firebase Console or in your `firebase-service-account.json` file (look for the `project_id` field).

**Option A: Export as environment variable**
```bash
export FIREBASE_PROJECT_ID=your-actual-project-id
```

**Option B: Create a .env file**
```bash
echo "FIREBASE_PROJECT_ID=your-actual-project-id" > .env
```

**Option C: Pass it directly to docker-compose**
```bash
FIREBASE_PROJECT_ID=your-actual-project-id sudo docker-compose up
```

## Step 2: Ensure configs Directory Exists

```bash
mkdir -p configs
```

## Step 3: Run the Server

```bash
# Set your project ID first
export FIREBASE_PROJECT_ID=your-actual-project-id

# Start the server
sudo docker-compose up server

# Or run in background
sudo docker-compose up -d server
```

## Step 4: Verify It's Running

```bash
# Check running containers
sudo docker ps

# Should show:
# CONTAINER ID   IMAGE                    STATUS          PORTS                    NAMES
# abc123...      art-print-backend-server Up X minutes    0.0.0.0:3001->3001/tcp   art-print-server

# Test the health endpoint
curl http://localhost:3001/health

# Should return: {"status":"ok","service":"art-print-backend"}
```

## Check Logs if Container Exits

If the container exits, check the logs:

```bash
# View server logs
sudo docker-compose logs server

# Or view all logs
sudo docker-compose logs

# Check what went wrong
sudo docker logs art-print-server
```

## Common Issues

### Issue: "FIREBASE_PROJECT_ID not set" or "your-project-id"
**Solution**: Set the environment variable:
```bash
export FIREBASE_PROJECT_ID=$(grep -o '"project_id": "[^"]*' firebase-service-account.json | cut -d'"' -f4)
sudo docker-compose up server
```

### Issue: "GOOGLE_APPLICATION_CREDENTIALS not set"
**Solution**: The file should be mounted automatically. Check if it exists:
```bash
ls -la firebase-service-account.json
```

### Issue: "Failed to create Firebase App"
**Solution**: Check that:
1. The `firebase-service-account.json` file is valid JSON
2. The `FIREBASE_PROJECT_ID` matches the project_id in the JSON file
3. The file has proper permissions

## Quick Start Script

Create a file `run-server.sh`:

```bash
#!/bin/bash
# Extract project ID from firebase-service-account.json
PROJECT_ID=$(grep -o '"project_id": "[^"]*' firebase-service-account.json | cut -d'"' -f4)

if [ -z "$PROJECT_ID" ]; then
    echo "❌ Could not extract project_id from firebase-service-account.json"
    exit 1
fi

echo "🚀 Starting server with project: $PROJECT_ID"
export FIREBASE_PROJECT_ID=$PROJECT_ID
sudo docker-compose up server
```

Make it executable and run:
```bash
chmod +x run-server.sh
./run-server.sh
```

