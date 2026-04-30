# APRS Updater - Raspberry Pi Setup

## Files needed:
- `main.go` - Source code (21KB)
- `pi-build.sh` - Build script (1.2KB)
- OR `aprsupdater-pi-package.tar.gz` - Both files packaged (6.2KB)

## Quick Start (easiest):

### Option 1: Use the tar package
```bash
# 1. Copy to Pi
scp aprsupdater-pi-package.tar.gz pi@PI-ADDRESS:/home/pi/

# 2. On the Pi:
ssh pi@PI-ADDRESS
tar -xzf aprsupdater-pi-package.tar.gz
bash pi-build.sh
```

### Option 2: Copy files separately
```bash
# 1. Copy both files to Pi
scp main.go pi-build.sh pi@PI-ADDRESS:/home/pi/

# 2. On the Pi:
ssh pi@PI-ADDRESS
bash pi-build.sh
```

## What the script does:
1. ✅ Detects your Pi architecture (ARM64 or ARMv6/v7)
2. ✅ Installs Go if not present
3. ✅ Installs build dependencies (OpenGL, X11 libs)
4. ✅ Builds `aprsupdater-pi` binary

## After build:
```bash
# Run the app
./aprsupdater-pi

# Or check help
./aprsupdater-pi --help
```

## Notes:
- Tested on Pi 3, 4, 5 (ARM64) and Pi Zero/1/2 (ARMv6)
- Requires Raspberry Pi OS with desktop (for GUI)
- For headless mode (no monitor), the app will auto-start web interface on port 8080

## Troubleshooting:
- If build fails, ensure you have internet access
- If "permission denied", run: `chmod +x pi-build.sh`
- For sudo password prompts, you'll need to type your Pi password
