#!/bin/bash

set -e

echo "Building imgr..."
echo

# Check for libheif
echo "→ Checking for libheif (required for HEIC support)..."
if command -v pkg-config &> /dev/null; then
    if pkg-config --exists libheif; then
        echo "  ✓ libheif found: $(pkg-config --modversion libheif)"
    else
        echo "  ⚠ libheif not found - HEIC support will not work"
        echo "    Install with: brew install libheif (macOS) or apt-get install libheif-dev (Linux)"
        read -p "Continue anyway? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
else
    echo "  ⚠ pkg-config not found, cannot verify libheif"
fi

# Initialize Go module if needed
if [ ! -f "go.mod" ]; then
    echo
    echo "→ Initializing Go module..."
    go mod init imgr
fi

# Get dependencies and tidy
echo
echo "→ Installing dependencies..."
go get github.com/urfave/cli/v2
go get golang.org/x/image/draw
go get golang.org/x/image/tiff
go get golang.org/x/image/webp
go get github.com/strukturag/libheif/go/heif
go mod tidy

# Create build directory
mkdir -p build

# Build for current platform
echo
echo "→ Building for current platform..."
CGO_ENABLED=1 go build -o build/imgr imgr.go

if [ $? -eq 0 ]; then
    echo
    echo "✓ Build complete!"
    echo
    ls -lh build/imgr
    
    echo
    echo "Supported formats:"
    echo "  • Read: JPEG, PNG, GIF, TIFF, WebP, HEIF/HEIC"
    echo "  • Write: JPEG, PNG, GIF, TIFF"
    echo
    echo "Usage:"
    echo "  ./build/imgr info image.jpg"
    echo "  ./build/imgr photo.heic photo.jpg"
    echo "  ./build/imgr -w 800 image.jpg small.jpg"
    echo
    echo "Install:"
    echo "  cp build/imgr /usr/local/bin/"
else
    echo
    echo "✗ Build failed"
    exit 1
fi