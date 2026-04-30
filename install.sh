#!/bin/bash
# Copy this file to Pi, then: bash install.sh
set -e
cd "$(dirname "$0")"

echo "=== APRS Updater Installer ==="
echo "Expected main.go MD5: 19229375682936f781174cd331b5b4bb"
echo ""

# Check if main.go exists and is valid
if [ -f main.go ]; then
    CHECKSUM=$(md5sum main.go | cut -d' ' -f1)
    echo "Found main.go with checksum: $CHECKSUM"
    if [ "$CHECKSUM" != "19229375682936f781174cd331b5b4bb" ]; then
        echo "ERROR: main.go is corrupted! Please re-copy from host."
        exit 1
    fi
    echo "main.go verified OK!"
else
    echo "ERROR: main.go not found! Please copy from host:"
    echo "scp /home/davey/Desktop/APRSupdater/main.go pi@IP:/home/pi/"
    exit 1
fi
echo ""

# Run build script
if [ -f pi-build.sh ]; then
    bash pi-build.sh
else
    echo "ERROR: pi-build.sh not found!"
    exit 1
fi
