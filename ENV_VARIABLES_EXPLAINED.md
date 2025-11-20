# Environment Variables Architecture Explained

## How FIREBASE_PROJECT_ID Works

The Firebase project ID (`cloudinary-trial` in your case) is **correctly NOT hardcoded** in Dockerfiles. Here's how it flows through the system:

### 1. Dockerfiles (Dockerfile.server, Dockerfile.seed)
✅ **Correct**: These files do NOT contain the project ID
- Dockerfiles build the application binary
- They should be environment-agnostic (work for dev, staging, prod)
- The project ID is provided at **runtime**, not build time

### 2. docker-compose.yml (Main config)
✅ **Correct**: Uses environment variable substitution
```yaml
environment:
  - FIREBASE_PROJECT_ID=${FIREBASE_PROJECT_ID}
```
- This reads from your shell environment or `docker-compose.override.yml`
- Allows different values for different environments

### 3. docker-compose.override.yml (Local override)
✅ **Correct**: You've set it to `cloudinary-trial` here
```yaml
services:
  server:
    environment:
      - FIREBASE_PROJECT_ID=cloudinary-trial
```
- This file is gitignored (safe to store your project ID)
- Automatically merged with `docker-compose.yml` by docker-compose
- Perfect for local development

### 4. docker-compose.prod.yml (Production)
✅ **Correct**: Uses environment variable
```yaml
environment:
  - FIREBASE_PROJECT_ID=${FIREBASE_PROJECT_ID}
```
- For production, you'd set this via:
  - Environment variable: `export FIREBASE_PROJECT_ID=cloudinary-trial`
  - Or create a production override file
  - Or use `.env` file

### 5. cloudbuild.yaml (GCP Deployment)
✅ **Fixed**: Now includes FIREBASE_PROJECT_ID
- Uses `${PROJECT_ID}` which is automatically set by Cloud Build
- This is the GCP project ID, which should match your Firebase project ID

## Your Current Setup

You have `cloudinary-trial` set in `docker-compose.override.yml`, which is **perfect** for local development.

## How to Use Different Environments

### Local Development (Current)
```bash
# docker-compose.override.yml has cloudinary-trial
sudo docker-compose up server
```

### Production (Different project)
```bash
# Option 1: Set environment variable
export FIREBASE_PROJECT_ID=production-project-id
sudo docker-compose -f docker-compose.prod.yml up

# Option 2: Create production override
# docker-compose.prod.override.yml
services:
  server:
    environment:
      - FIREBASE_PROJECT_ID=production-project-id
```

### GCP Deployment
The `cloudbuild.yaml` automatically uses the GCP project ID, which should match your Firebase project ID.

## Summary

✅ **Dockerfiles**: Correctly don't have project ID (build-time)
✅ **docker-compose.yml**: Correctly uses environment variable (runtime)
✅ **docker-compose.override.yml**: Correctly has your project ID (local dev)
✅ **cloudbuild.yaml**: Now fixed to include project ID (GCP deployment)

Everything is set up correctly! The project ID flows from:
- `docker-compose.override.yml` → `docker-compose.yml` → Container environment → Your Go application

## Verify It's Working

```bash
# Check that the container has the right project ID
sudo docker exec art-print-server env | grep FIREBASE_PROJECT_ID
# Should show: FIREBASE_PROJECT_ID=cloudinary-trial
```

