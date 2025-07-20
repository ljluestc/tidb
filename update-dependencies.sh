#!/bin/bash

# This script updates project dependencies

set -e

echo "Updating dependencies..."

# Clean the module cache for the problematic package
go clean -modcache github.com/stathat/consistent

# Update go.mod and go.sum
go mod tidy

# Download all dependencies
go mod download

echo "Dependencies updated successfully"
