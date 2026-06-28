# db-gen Makefile
# Build and publish db-gen for Windows and Linux

BINARY_NAME := db-gen
BUILD_DIR := build

# Detect OS and use appropriate build script
ifeq ($(OS),Windows_NT)
    BUILD_SCRIPT = powershell -ExecutionPolicy Bypass -File build.ps1
else
    BUILD_SCRIPT = ./build.sh
endif

# Default target
.PHONY: all
all: build

# Build for all platforms
.PHONY: build
build:
	$(BUILD_SCRIPT) all

# Build for Windows (64-bit)
.PHONY: build-windows
build-windows:
	$(BUILD_SCRIPT) windows

# Build for Linux (64-bit)
.PHONY: build-linux
build-linux:
	$(BUILD_SCRIPT) linux

# Clean build artifacts
.PHONY: clean
clean:
	$(BUILD_SCRIPT) clean

# Run all tests (DB-backed tests auto-skip when no database is reachable).
# Set DBGEN_TEST_DSN to point at a Postgres instance, or rely on the
# ConnectionString in test/db-gen-copy.json.
.PHONY: test
test:
	go test ./...

# Same as `test` but bypasses the Go test cache (always re-runs).
.PHONY: test-force
test-force:
	go test -count=1 ./...

# Fast, pure-Go unit tests only (no database needed).
.PHONY: test-unit
test-unit:
	go test -short ./private/...

# Integration + e2e tests (require a reachable test database with fixtures).
.PHONY: test-integration
test-integration:
	go test -v ./private/dbGen -run 'Integration|E2E'

# Regenerate e2e golden files (run after intentionally changing templates/output).
.PHONY: test-update-golden
test-update-golden:
	go test ./private/dbGen -run E2E -update

# Help
.PHONY: help
help:
	@echo "db-gen Makefile targets:"
	@echo "  make build              - Build for Windows and Linux"
	@echo "  make build-windows      - Build for Windows only"
	@echo "  make build-linux        - Build for Linux only"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make test               - Run all tests (DB tests skip if no DB)"
	@echo "  make test-force         - Run all tests, ignoring the test cache"
	@echo "  make test-unit          - Run pure unit tests only"
	@echo "  make test-integration   - Run DB-backed integration + e2e tests"
	@echo "  make test-update-golden - Regenerate e2e golden files"
	@echo ""
	@echo "Output directory: $(BUILD_DIR)/"
