#!/bin/bash
# Build script for the pure Go implementation

set -e

echo "Building shaka-packager pure Go implementation..."

# Download Go dependencies
echo "Downloading Go dependencies..."
go mod download

# Verify Go module
echo "Verifying Go module..."
go mod verify

# Run Go linting
echo "Running Go linting..."
go fmt ./...
go vet ./...

# Build Go package
echo "Building Go package..."
go build -v

# Run Go tests
echo "Running Go tests..."
go test -v ./...

echo "Build completed successfully!"