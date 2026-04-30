#!/bin/bash
# APRS Updater - Pi Builder
# Copy main.go and this script to Pi, then run: bash pi-build.sh

set -e

cd "$(dirname "$0")"

echo "=== APRS Updater Pi Builder ==="
echo ""

# Detect architecture
ARCH=$(uname -m)
echo "Detected architecture: $ARCH"
echo ""

# Install Go if needed
if ! command -v go &> /dev/null; then
    echo "Go not found. Installing..."
    
    if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        GO_ARCH="arm64"
    else
        GO_ARCH="armv6l"
    fi
    
    FILE="go1.21.6.linux-${GO_ARCH}.tar.gz"
    URL="https://go.dev/dl/${FILE}"
    
    echo "Downloading Go 1.21.6 for ${GO_ARCH}..."
    wget -q --show-progress "$URL" -O "$FILE"
    
    echo "Extracting Go..."
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$FILE"
    rm -f "$FILE"
    
    # Add to PATH permanently
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    
    export PATH=$PATH:/usr/local/go/bin
    echo "Go installed!"
fi

# Ensure PATH has Go
export PATH=$HOME/go/bin:/usr/local/go/bin:$PATH

echo "Go version:"
go version
echo ""

# Initialize Go module if needed
if [ ! -f go.mod ]; then
    echo "Initializing Go module..."
    go mod init aprsupdater
    go mod tidy
    echo "Module initialized!"
fi
echo ""

# Install dependencies
echo "Installing dependencies..."
sudo apt-get update
sudo apt-get install -y xorg-dev libgl1-mesa-dev libxrandr-dev libxxf86vm-dev libxi-dev libxcursor-dev
echo "Dependencies installed!"
echo ""

# Build
echo "Building APRS Updater..."
CGO_ENABLED=1 go build -o aprsupdater-pi .
echo ""
echo "=== Build Complete! ==="
ls -lh aprsupdater-pi
echo ""
echo "To run: ./aprsupdater-pi"
