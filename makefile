# ─────────────────────────────
# Art Print Backend — Makefile
# ─────────────────────────────

APP_NAME=art-print-backend
BIN_DIR=bin
CMD_SERVER=cmd/server
CMD_SEED=cmd/seed
CONFIG_DIR=configs
GO_ENV?=dev

# ─────────────────────────────
# Default target
# ─────────────────────────────
.PHONY: run
run:
	@echo "🚀 Running $(APP_NAME) ..."
	@cp $(CONFIG_DIR)/.env.$(GO_ENV) .env 2>/dev/null || echo "⚠️ No $(CONFIG_DIR)/.env.$(GO_ENV) found, using defaults"
	env GO_ENV=$(GO_ENV) go run ./$(CMD_SERVER)
	@rm -f .env

# ─────────────────────────────
# Build targets
# ─────────────────────────────
.PHONY: build
build:
	@echo "🏗️  Building $(APP_NAME)..."
	go build -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_SERVER)
	@echo "✅ Build complete: $(BIN_DIR)/$(APP_NAME)"

.PHONY: build-win
build-win:
	@echo "🏗️  Building Windows binary..."
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME).exe ./$(CMD_SERVER)
	@echo "✅ Windows binary ready: $(BIN_DIR)/$(APP_NAME).exe"

.PHONY: build-linux
build-linux:
	@echo "🏗️  Building Linux binary..."
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_SERVER)
	@echo "✅ Linux binary ready: $(BIN_DIR)/$(APP_NAME)"

# ─────────────────────────────
# Seeder only (manual use)
# ─────────────────────────────
.PHONY: seed
seed:
	@echo "🌱 Running seeders..."
	env GO_ENV=$(GO_ENV) go run ./$(CMD_SEED)

# ─────────────────────────────
# Firebase Emulators
# ─────────────────────────────
.PHONY: emulators
emulators:
	@echo "🔥 Starting Firebase emulators..."
	@cp $(CONFIG_DIR)/.env.$(GO_ENV) .env 2>/dev/null || echo "⚠️ No $(CONFIG_DIR)/.env.$(GO_ENV) found, using defaults"

	# Start emulators in the background
	firebase emulators:start &
	EMULATOR_PID=$$!

	@echo "⏳ Waiting for Firebase emulators to boot..."
	sleep 5

	@echo "🌱 Running seeder tool..."
	env GO_ENV=$(GO_ENV) go run ./$(CMD_SEED)

	@echo "📡 Emulator logs:"
	wait $$EMULATOR_PID

	@rm -f .env

# ─────────────────────────────
# Linting & Cleanup
# ─────────────────────────────
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	@echo "🧹 Cleaning build files..."
	rm -rf $(BIN_DIR)/*
