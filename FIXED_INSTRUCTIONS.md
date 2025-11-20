# Fixed: How to Run Containers with Firebase Project ID

The issue was that `FIREBASE_PROJECT_ID` wasn't being set when using `sudo`. Here are the solutions:

## Solution 1: Use the Helper Script (Easiest)

The `run-server.sh` script now automatically:
1. Extracts your project ID from `firebase-service-account.json`
2. Creates a `docker-compose.override.yml` file with the project ID
3. Starts the server

```bash
./run-server.sh
```

## Solution 2: Manual Setup

### Step 1: Extract Your Project ID

```bash
# Get your project ID from the JSON file
grep -o '"project_id": "[^"]*' firebase-service-account.json | cut -d'"' -f4
```

### Step 2: Create docker-compose.override.yml

Create a file `docker-compose.override.yml` in the project root:

```yaml
version: '3.8'

services:
  server:
    environment:
      - FIREBASE_PROJECT_ID=your-actual-project-id-here
  seed:
    environment:
      - FIREBASE_PROJECT_ID=your-actual-project-id-here
```

Replace `your-actual-project-id-here` with your actual project ID.

### Step 3: Ensure configs directory exists

```bash
mkdir -p configs
```

### Step 4: Run the server

```bash
sudo docker-compose up server
```

## Solution 3: Use Environment Variable (if sudo preserves env)

Some systems allow `sudo -E` to preserve environment variables:

```bash
# Extract project ID
export FIREBASE_PROJECT_ID=$(grep -o '"project_id": "[^"]*' firebase-service-account.json | cut -d'"' -f4)

# Run with sudo -E
sudo -E docker-compose up server
```

## Verify It's Working

```bash
# Check container is running
sudo docker ps

# Test health endpoint
curl http://localhost:3001/health

# Should return: {"status":"ok","service":"art-print-backend"}
```

## What Was Fixed

1. **Dockerfile**: Now creates `/app/configs` directory in the container
2. **docker-compose.yml**: Fixed volume mount order
3. **run-server.sh**: Now creates `docker-compose.override.yml` automatically
4. **docker-compose.override.yml**: Added to `.gitignore` so you can safely store your project ID locally

## Troubleshooting

If you still get "FIREBASE_PROJECT_ID not set":

1. Check that `docker-compose.override.yml` exists and has the correct project ID
2. Verify the project ID matches what's in `firebase-service-account.json`:
   ```bash
   cat firebase-service-account.json | grep project_id
   ```
3. Rebuild the container:
   ```bash
   sudo docker-compose build --no-cache server
   sudo docker-compose up server
   ```

