#!/bin/bash
set -e

# Script to generate proto code
# Usage: ./scripts/generate-proto.sh

echo "Checking prerequisites..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "ERROR: protoc is not installed."
    echo "Install it with:"
    echo "  sudo apt-get update && sudo apt-get install -y protobuf-compiler"
    echo "  or visit https://grpc.io/docs/protoc-installation/"
    exit 1
fi

echo "✓ protoc found: $(protoc --version)"

# Get GOPATH and add to PATH
GOPATH=$(go env GOPATH)
export PATH=$PATH:$GOPATH/bin

# Install Go plugins if not present
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

echo "✓ All plugins installed"

# Generate proto code
echo "Generating proto code..."
protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    --proto_path=. \
    proto/product/v1/product_service.proto

echo "✓ Proto code generated successfully!"
echo "Generated files:"
ls -la proto/product/v1/*.go 2>/dev/null || echo "  (check proto/product/v1/ directory)"
