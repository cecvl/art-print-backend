# ─────────────────────────────
# Art Print Backend — Makefile
# ─────────────────────────────

APP_NAME=art-print-backend
BIN_DIR=bin
CMD_DIR=cmd/server
CONFIG_DIR=configs
GO_ENV?=dev

# Default target
.PHONY: run
run:
	@echo "🚀 Running $(APP_NAME) ..."
	@cp $(CONFIG_DIR)/.env.$(GO_ENV) .env 2>/dev/null || echo "⚠️ No $(CONFIG_DIR)/.env.$(GO_ENV) found, using defaults"
	env GO_ENV=$(GO_ENV) go run ./$(CMD_DIR)
	@rm -f .env

# ─────────────────────────────
# Build targets
# ─────────────────────────────
.PHONY: build
build:
	@echo "🏗️  Building $(APP_NAME)..."
	go build -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "✅ Build complete: $(BIN_DIR)/$(APP_NAME)"

.PHONY: build-win
build-win:
	@echo "🏗️  Building Windows binary..."
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME).exe ./$(CMD_DIR)
	@echo "✅ Windows binary ready: $(BIN_DIR)/$(APP_NAME).exe"

.PHONY: build-linux
build-linux:
	@echo "🏗️  Building Linux binary..."
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "✅ Linux binary ready: $(BIN_DIR)/$(APP_NAME)"

# ─────────────────────────────
# Firebase Emulators
# ─────────────────────────────
.PHONY: emulators
emulators:
	@echo "🔥 Starting Firebase emulators..."
	@cp $(CONFIG_DIR)/.env.$(GO_ENV) .env 2>/dev/null || echo "⚠️ No $(CONFIG_DIR)/.env.$(GO_ENV) found, using defaults"
	firebase emulators:start --import=./emulator_data --export-on-exit
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
