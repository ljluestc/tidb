#!/bin/bash

# This script builds the tidb-server binary needed for tests

set -e

echo "Building tidb-server binary..."

# Make sure bin directory exists
mkdir -p bin

# Build tidb-server
go build -o bin/tidb-server cmd/tidb-server/main.go

echo "tidb-server binary built successfully at bin/tidb-server"
chmod +x bin/tidb-server

# Verify the binary was created
if [ -f bin/tidb-server ]; then
    echo "Verified: bin/tidb-server exists and is executable"
else
    echo "Error: Failed to create bin/tidb-server"
    exit 1
fi
