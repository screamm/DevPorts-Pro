#!/bin/bash

echo "DevPorts Pro Cross-Platform Build Script"
echo "========================================"

# Create dist directory
mkdir -p dist

# Set CGO_ENABLED=1 for GUI builds (requires platform-specific dependencies)
export CGO_ENABLED=1

echo ""
echo "Note: This build requires platform-specific dependencies."
echo "For Windows: Build on Windows machine or use Docker"
echo "For macOS: Build on macOS machine"
echo "For Linux: Install X11/GL development libraries"
echo ""

# Check if we can build for current platform
echo "Building for current platform ($(go env GOOS)/$(go env GOARCH))..."

if go build -ldflags "-w -s" -o dist/devports-pro-$(go env GOOS)-$(go env GOARCH) main.go port_scanner.go; then
    echo "✅ Successfully built for $(go env GOOS)/$(go env GOARCH)"
else
    echo "❌ Failed to build for $(go env GOOS)/$(go env GOARCH)"
    echo "Try installing GUI development dependencies:"
    echo "  Ubuntu/Debian: sudo apt install libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libgl1-mesa-dev pkg-config"
    echo "  CentOS/RHEL: sudo yum install libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel mesa-libGL-devel pkgconfig"
fi

echo ""
echo "For cross-platform builds, use the specific platform's build environment."
echo "Command-line version (no GUI dependencies):"
echo "  go build -o test-scanner test_scanner.go port_scanner.go"

echo ""
echo "Build complete! Check the dist/ directory."
if [ -d "dist" ]; then
    ls -la dist/
fi