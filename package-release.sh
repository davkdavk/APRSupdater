#!/bin/bash
# Package daemon binaries for release

set -e

VERSION=${1:-"v1.0.0"}
OUTPUT_DIR="release-${VERSION}"

echo "Packaging APRS Updater Daemon ${VERSION}..."

mkdir -p "${OUTPUT_DIR}"

# Files to include in each package
for binary in aprsupdater-daemon-x86_64 aprsupdater-daemon-arm64 aprsupdater-daemon-arm32; do
    if [ ! -f "$binary" ]; then
        echo "Warning: $binary not found, skipping"
        continue
    fi
    
    pkg="${OUTPUT_DIR}/${binary}.tar.gz"
    echo "Creating ${pkg}..."
    tar czf "$pkg" "$binary" README.md daemon-app/ 2>/dev/null || tar czf "$pkg" "$binary"
    echo "  Created: $pkg"
done

echo ""
echo "Release packages in: ${OUTPUT_DIR}/"
ls -lh "${OUTPUT_DIR}/"

echo ""
echo "To upload manually:"
echo "  gh release upload ${VERSION} ${OUTPUT_DIR}/*"
