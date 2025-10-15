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

# Run tests
.PHONY: test
test:
	go test -v ./...

# Help
.PHONY: help
help:
	@echo "db-gen Makefile targets:"
	@echo "  make build         - Build for Windows and Linux"
	@echo "  make build-windows - Build for Windows only"
	@echo "  make build-linux   - Build for Linux only"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make test          - Run tests"
	@echo ""
	@echo "Output directory: $(BUILD_DIR)/"
