#!/bin/bash

echo "Building DevPorts Pro for multiple platforms..."

# Create dist directory
mkdir -p dist

# Build for Windows (64-bit)
echo "Building for Windows (64-bit)..."
GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui -w -s" -o dist/devports-pro-windows.exe main.go port_scanner.go

# Build for macOS (64-bit Intel)
echo "Building for macOS (64-bit Intel)..."  
GOOS=darwin GOARCH=amd64 go build -ldflags "-w -s" -o dist/devports-pro-macos-intel main.go port_scanner.go

# Build for macOS (ARM64 - Apple Silicon)
echo "Building for macOS (ARM64 - Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-w -s" -o dist/devports-pro-macos-arm64 main.go port_scanner.go

# Build for Linux (64-bit)
echo "Building for Linux (64-bit)..."
GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o dist/devports-pro-linux main.go port_scanner.go

echo "Build complete! Check the dist/ directory for executables."
echo ""
echo "Files created:"
ls -la dist/