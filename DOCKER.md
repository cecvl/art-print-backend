# Docker Setup Guide

This project uses Docker for containerization with separate containers for different modules, optimized for GCP deployment.

## Architecture

- **Server Container**: Main API server (`Dockerfile.server`)
- **Seed Container**: Database seeding utility (`Dockerfile.seed`)
- **Data**: Seed data is embedded in binaries using Go's `embed` package (modern approach)

## Docker Files

### Dockerfile.server
Multi-stage build for the server module:
- **Stage 1 (builder)**: Compiles Go binary with optimizations
- **Stage 2 (runtime)**: Minimal Alpine image with only runtime dependencies

### Dockerfile.seed
Multi-stage build for the seed module:
- Similar structure to server Dockerfile
- Contains embedded seed data in the binary

### Dockerfile
Default Dockerfile that builds the server (for backward compatibility)

## Local Development

### Using Docker Compose

1. **Development mode** (with Firebase emulators):
```bash
docker-compose --profile emulator up
```

2. **Run server only** (connects to Firebase and Cloudinary):
```bash
docker-compose up server
```
Note: Ensure your Cloudinary credentials are set in `configs/.env.dev` or `docker-compose.override.yml`:
- `CLOUDINARY_CLOUD_NAME`
- `CLOUDINARY_API_KEY`
- `CLOUDINARY_API_SECRET`

3. **Run seed job**:
```bash
docker-compose --profile seed run seed
```

4. **Production-like setup** (no seed data, production environment):
```bash
# Ensure configs directory and symlink are set up
./setup-configs.sh

# Set your Firebase project ID
export FIREBASE_PROJECT_ID=cloudinary-trial

# Set Cloudinary credentials (or use configs/.env.production)
export CLOUDINARY_CLOUD_NAME=your-cloud-name
export CLOUDINARY_API_KEY=your-api-key
export CLOUDINARY_API_SECRET=your-api-secret

# Run production-like setup
docker-compose -f docker-compose.prod.yml up

# Or run in background
docker-compose -f docker-compose.prod.yml up -d
```

**Note**: Seed data will NOT run automatically (only runs when `APP_ENV=dev`). The server will connect to:
- ✅ Firebase (production credentials)
- ✅ Cloudinary (from environment variables)

### Quick Start Scripts

**Development mode** (with seed data):
```bash
./run-server.sh
```

**Production-like mode** (no seed data):
```bash
./run-prod.sh
```

### Building Images Manually

```bash
# Build server image
docker build -f Dockerfile.server -t art-print-server:latest .

# Build seed image
docker build -f Dockerfile.seed -t art-print-seed:latest .

# Run server container (development)
docker run -p 8080:8080 \
  -e APP_ENV=dev \
  -e GOOGLE_APPLICATION_CREDENTIALS=/app/configs/firebase-service-account.json \
  -e FIREBASE_PROJECT_ID=your-project-id \
  -v $(pwd)/configs:/app/configs:ro \
  art-print-server:latest

# Run server container (production-like, no seed data)
docker run -p 8080:8080 \
  -e APP_ENV=production \
  -e GOOGLE_APPLICATION_CREDENTIALS=/app/configs/firebase-service-account.json \
  -e FIREBASE_PROJECT_ID=your-project-id \
  -e CLOUDINARY_CLOUD_NAME=your-cloud-name \
  -e CLOUDINARY_API_KEY=your-api-key \
  -e CLOUDINARY_API_SECRET=your-api-secret \
  -v $(pwd)/configs:/app/configs:ro \
  art-print-server:latest

# Run seed container
docker run --rm \
  -e APP_ENV=dev \
  -e GOOGLE_APPLICATION_CREDENTIALS=/app/configs/firebase-service-account.json \
  -e FIREBASE_PROJECT_ID=your-project-id \
  -v $(pwd)/configs:/app/configs:ro \
  art-print-seed:latest
```

## GCP Deployment

### Prerequisites

1. **Artifact Registry Repository**:
```bash
gcloud artifacts repositories create art-print-backend \
  --repository-format=docker \
  --location=us-central1 \
  --description="Art Print Backend Docker repository"
```

2. **Service Account Secret** (for Firebase credentials):
```bash
# Create secret in Secret Manager
gcloud secrets create firebase-service-account \
  --data-file=firebase-service-account.json
```

3. **Cloud Run Service Account**:
```bash
# Grant necessary permissions
gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:SERVICE_ACCOUNT@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### Using Cloud Build

1. **Configure substitutions** in Cloud Build trigger or use defaults:
   - `_REGION`: GCP region (default: `us-central1`)
   - `_REPO_NAME`: Artifact Registry repository name (default: `art-print-backend`)
   - `_SERVICE_NAME`: Cloud Run service name (default: `art-print-server`)
   - `_APP_ENV`: Application environment (default: `production`)

2. **Trigger build**:
```bash
gcloud builds submit --config=cloudbuild.yaml
```

3. **Deploy seed job** (optional):
```bash
gcloud builds submit --config=cloudbuild-seed.yaml
```

### Manual GCP Deployment

1. **Build and push images**:
```bash
# Set variables
export PROJECT_ID=your-project-id
export REGION=us-central1
export REPO_NAME=art-print-backend

# Build and push server
docker build -f Dockerfile.server -t ${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/server:latest .
docker push ${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/server:latest

# Build and push seed
docker build -f Dockerfile.seed -t ${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/seed:latest .
docker push ${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/seed:latest
```

2. **Deploy to Cloud Run**:
```bash
gcloud run deploy art-print-server \
  --image ${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/server:latest \
  --region ${REGION} \
  --platform managed \
  --allow-unauthenticated \
  --port 8080 \
  --memory 512Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 10 \
  --set-env-vars APP_ENV=production,PORT=8080 \
  --set-secrets GOOGLE_APPLICATION_CREDENTIALS=firebase-service-account:latest
```

3. **Run seed job** (Cloud Run Job):
```bash
gcloud run jobs create art-print-seed \
  --image ${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/seed:latest \
  --region ${REGION} \
  --set-env-vars APP_ENV=production \
  --set-secrets GOOGLE_APPLICATION_CREDENTIALS=firebase-service-account:latest

# Execute the job
gcloud run jobs execute art-print-seed --region ${REGION}
```

## Docker Layer Optimization

The Dockerfiles are optimized for layer caching:

1. **Dependencies layer**: `go.mod` and `go.sum` are copied first and dependencies downloaded separately
2. **Source code layer**: Source code is copied after dependencies
3. **Build layer**: Binary is built in a separate stage
4. **Runtime layer**: Only necessary runtime files are copied to final image

This ensures that dependency changes don't invalidate source code layers and vice versa.

## Seed Data

Seed data is embedded directly into the binaries using Go's `embed` package:
- Data files are located in `internal/seeders/data/`
- JSON files are embedded at compile time
- No external file dependencies at runtime
- Modern Go approach for including static data

## Health Checks

The server includes a `/health` endpoint that returns:
```json
{
  "status": "ok",
  "service": "art-print-backend"
}
```

Docker health checks are configured to use this endpoint.

## Environment Variables

### Server Container
- `APP_ENV`: Application environment (`dev`, `production`)
- `PORT`: Server port (default: `8080`)
- `GOOGLE_APPLICATION_CREDENTIALS`: Path to Firebase service account JSON
- `FIREBASE_PROJECT_ID`: Firebase project ID
- `FIRESTORE_EMULATOR_HOST`: Firestore emulator host (for local dev)
- `FIREBASE_AUTH_EMULATOR_HOST`: Auth emulator host (for local dev)

### Seed Container
- Same as server, but `PORT` is not needed

## Troubleshooting

### Build Issues
- Ensure Go version matches `go.mod` (1.23.2)
- Check that all dependencies are available
- Verify `.dockerignore` isn't excluding necessary files

### Runtime Issues
- Check logs: `docker logs <container-name>`
- Verify environment variables are set correctly
- Ensure Firebase credentials are accessible
- Check health endpoint: `curl http://localhost:8080/health`

### GCP Deployment Issues
- Verify Artifact Registry repository exists
- Check IAM permissions for Cloud Build service account
- Ensure Secret Manager secret exists and is accessible
- Review Cloud Build logs in GCP Console

