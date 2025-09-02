#!/bin/bash
# Build script for the Go wrapper

set -e

echo "Building shaka-packager Go wrapper..."

# Create build directory
BUILD_DIR="build"
if [ -d "$BUILD_DIR" ]; then
    rm -rf "$BUILD_DIR"
fi
mkdir -p "$BUILD_DIR"

cd "$BUILD_DIR"

# Configure with CMake
echo "Configuring C++ wrapper..."
cmake .. -DCMAKE_BUILD_TYPE=Release

# Build the C++ wrapper
echo "Building C++ wrapper..."
make -j$(nproc)

# Go back to wrapper directory
cd ..

# Download Go dependencies
echo "Downloading Go dependencies..."
go mod download

# Run Go tests
echo "Running Go tests..."
go test -v ./...

# Build Go package
echo "Building Go package..."
go build -v

echo "Build completed successfully!"