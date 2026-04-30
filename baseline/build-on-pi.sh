#!/bin/bash
# Build script - RUN THIS ON THE RASPBERRY PI
# Copy this script and the main.go file to your Pi, then run it

set -e

echo "=== APRS Updater Pi Builder ==="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go not found. Installing..."
    # Detect Pi architecture
    ARCH=$(uname -m)
    if [ "$ARCH" = "aarch64" ]; then
        GO_ARCH="arm64"
    else
        GO_ARCH="armv6l"
    fi
    # Download and install Go
    wget "https://go.dev/dl/go1.21.6.linux-${GO_ARCH}.tar.gz"
    sudo tar -C /usr/local -xzf "go1.21.6.linux-${GO_ARCH}.tar.gz"
    export PATH=$PATH:/usr/local/go/bin
    echo "Go installed!"
fi

export PATH=$HOME/go/bin:/usr/local/go/bin:$PATH

echo "Go version:"
go version
echo ""

# Install dependencies (Fyne needs X11/OpenGL libs)
echo "Installing dependencies..."
sudo apt-get update
sudo apt-get install -y xorg-dev libgl1-mesa-dev libxrandr-dev libxxf86vm-dev libxi-dev libxcursor-dev

echo ""
echo "Building 64-bit binary (for Pi 3+/4/5)..."
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o aprsupdater-pi64 .

# Check if 32-bit build is needed
if [ "$(uname -m)" = "armv6l" ] || [ "$(uname -m)" = "armv7l" ]; then
    echo ""
    echo "Building 32-bit binary (for Pi Zero/1/2/3)..."
    CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build -o aprsupdater-pi32 .
fi

echo ""
echo "=== Build Complete! ==="
ls -lh aprsupdater-pi*
echo ""
echo "To run: ./aprsupdater-pi64 (or ./aprsupdater-pi32 on older Pis)"
