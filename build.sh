#!/bin/bash
# db-gen Build Script for Linux/macOS
# Builds db-gen for Windows and Linux

BINARY_NAME="db-gen"
BUILD_DIR="build"
BUILD_FLAGS="-trimpath"
LD_FLAGS="-s -w"

build_windows() {
    echo "Building for Windows (amd64)..."
    mkdir -p "$BUILD_DIR"
    GOOS=windows GOARCH=amd64 go build $BUILD_FLAGS -ldflags "$LD_FLAGS" -o "$BUILD_DIR/$BINARY_NAME-win.exe" .
    if [ $? -eq 0 ]; then
        echo "Windows build complete: $BUILD_DIR/$BINARY_NAME-win.exe"
    else
        echo "Windows build failed"
        exit 1
    fi
}

build_linux() {
    echo "Building for Linux (amd64)..."
    mkdir -p "$BUILD_DIR"
    GOOS=linux GOARCH=amd64 go build $BUILD_FLAGS -ldflags "$LD_FLAGS" -o "$BUILD_DIR/$BINARY_NAME-linux" .
    if [ $? -eq 0 ]; then
        echo "Linux build complete: $BUILD_DIR/$BINARY_NAME-linux"
    else
        echo "Linux build failed"
        exit 1
    fi
}

clean_build() {
    echo "Cleaning build artifacts..."
    rm -rf "$BUILD_DIR"
    rm -f "$BINARY_NAME" "$BINARY_NAME.exe"
    echo "Clean complete"
}

# Main execution
case "${1:-all}" in
    windows)
        build_windows
        ;;
    linux)
        build_linux
        ;;
    clean)
        clean_build
        ;;
    all)
        clean_build
        build_windows
        build_linux
        ;;
    *)
        echo "Usage: ./build.sh [all|windows|linux|clean]"
        echo "  all     - Build for Windows and Linux (default)"
        echo "  windows - Build for Windows only"
        echo "  linux   - Build for Linux only"
        echo "  clean   - Remove build artifacts"
        exit 1
        ;;
esac

echo ""
echo "Build complete!"
