#!/bin/bash
# Restore APRS Updater from baseline
set -e

echo "Restoring APRS Updater from baseline..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$SCRIPT_DIR")"

# Restore source
cp "$SCRIPT_DIR/main.go.baseline" "$BASE_DIR/main.go"
echo "✓ Restored main.go"

# Restore config (optional)
read -p "Restore config file too? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    cp "$SCRIPT_DIR/aprsupdater-config.baseline.json" ~/.aprsupdater.json
    echo "✓ Restored ~/.aprsupdater.json"
fi

# Rebuild
cd "$BASE_DIR"
export PATH=$HOME/go/bin:$PATH
go build -o aprsupdater . 2>&1 && echo "✓ Built aprsupdater binary"

echo ""
echo "Restore complete! Run with: ./aprsupdater"
