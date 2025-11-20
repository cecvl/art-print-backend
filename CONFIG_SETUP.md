# Configuration Setup Explained

## Firebase Credentials File Location

Your `firebase-service-account.json` file is in the **project root**, but the Docker container expects it at `/app/configs/firebase-service-account.json`.

## The Solution: Symlink

We've created a **symlink** in the `configs/` directory that points to the file in root:

```
configs/firebase-service-account.json -> ../firebase-service-account.json
```

This allows:
1. ✅ Keep the file in root (where you have it)
2. ✅ Mount the `configs/` directory (which contains `.env` files)
3. ✅ The file is accessible at `/app/configs/firebase-service-account.json` in the container

## How It Works

```
Project Root:
├── firebase-service-account.json  (actual file)
└── configs/
    ├── .env.dev
    ├── .env
    └── firebase-service-account.json  (symlink → ../firebase-service-account.json)
```

When Docker mounts `./configs:/app/configs`, it includes:
- All `.env` files
- The symlink (which resolves to the file in root)

## Setup Script

Run `./setup-configs.sh` to create the symlink automatically:

```bash
./setup-configs.sh
```

Or the `run-server.sh` script will do this automatically.

## Why Not Just Copy the File?

We use a symlink instead of copying because:
- ✅ Keeps a single source of truth (file in root)
- ✅ No need to sync changes
- ✅ Works with git (symlink is tracked, actual file is gitignored)

## Verification

Check that the symlink works:

```bash
# Check symlink exists
ls -la configs/firebase-service-account.json

# Should show: firebase-service-account.json -> ../firebase-service-account.json

# Verify it resolves correctly
cat configs/firebase-service-account.json | head -1
# Should show JSON content
```

## Docker Compose Configuration

The `docker-compose.yml` now simply mounts the configs directory:

```yaml
volumes:
  - ./configs:/app/configs:ro
```

This mounts:
- All `.env` files from `configs/`
- The `firebase-service-account.json` symlink (which resolves to the file in root)

The environment variable is set to:
```yaml
GOOGLE_APPLICATION_CREDENTIALS=/app/configs/firebase-service-account.json
```

And it will correctly resolve to your file in root via the symlink.

