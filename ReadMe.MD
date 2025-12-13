# Web App Server

### Introduction

Back-end for an art prints' sales platform

### Tools

- Go standard libs
- `Cloudinary` for storage
- `Firebase` as backend support, `firestore` as db

`Docker setup`

- Docker & Docker Compose installed

### Curl commands to test server

-`sign up`: curl -X POST http://localhost:8080/signup \
 -H "Content-Type: application/json" \
 -d '{
"email": "useruser@gmail.com",
"password": "securepass123",
"userType": "artist"
}'

- curl -X GET http://localhost:8080/artworks
- curl -X POST http://localhost:8080/cart/add -H "Authorization: Bearer ID_TOKEN" -H "Content-Type: application/json" -d '{
  "artworkId": "7HCcdzetfE69uFLtTfGd",
  "title": "",
  "quantity": 1,
  "price": 1500.00
  }'

- curl -X GET http://localhost:8080/cart -H "Authorization: Bearer ID_TOKEN"
- curl -X POST http://localhost:8080/checkout -H "Authorization: Bearer ID_TOKEN"

# Make commands
Run `make --version` to check installation
##
`make run` - run app server

`make build` - build a binary file

`make build-win`
`make build-linux` - build for windows, linux

`make seed` - seed json data

`make emulators` - run firebase emulators

### Running in Production Mode (Docker)

To run the server in a production-like environment (using Docker, without seed data):

1. Ensure you have `docker` and `docker compose` installed.
2. Place your `firebase-service-account.json` file in the `configs/` directory.
3. Run the helper script:

```bash
./run-prod.sh
```

This script will:
- Verify your credentials exist.
- Build the Docker container.
- Start the server on port 8080.
- Connect to the real Firebase project (not emulators).


# Quick Start with Docker

- 1. Clone the repository & `cd art-print-backend`

- 2. Set up environment configuration
`cp docker-compose.override.yml.example docker-compose.override.yml` then edit docker-compose.override.yml with environment variables

- 3. Build and run for development
`docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build`

- 4. Or for production-like environment
`docker-compose -f docker-compose.yml -f docker-compose.prod.yml up --build`

## Docker commands

- Build the main server image
`docker build -f Dockerfile.server -t art-print-server .`

- Build the seed image
`docker build -f Dockerfile.seed -t art-print-seed .`

- Run a specific service
`docker-compose up server` or `docker-compose up seed`

- View running containers
`docker ps`

- View logs
`docker logs -f art-print-server`
