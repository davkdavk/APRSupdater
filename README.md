# APRS Updater

APRS object packet updater with two interfaces:
- **GUI version** - Desktop app built with Fyne toolkit
- **Daemon version** - Headless server with web UI at `http://:8080/`

Sends APRS object packets via APRS-IS protocol. Supports up to 10 objects with configurable symbols, coordinates, and descriptions.

## Features
- 🖥️ **GUI Mode** (`main.go`) - Full desktop interface with Fyne
- 🌐 **Web Mode** (`daemon-app/`) - LAN-accessible web UI at port 8080
- 📡 **Real APRS-IS** - TCP connection to APRS servers (rotate.aprs.net)
- ⏱️ **Auto-send** - Configurable interval (default 15 min)
- 🔧 **Cross-platform** - Builds for Linux (x86/ARM) and Windows
- 💾 **Shared config** - `~/.aprsupdater.json` or `aprsupdater.json` next to binary

## Quick Start

### GUI Version
```bash
go build -o aprsupdater .
./aprsupdater
```

### Daemon Version
```bash
cd daemon-app
go build -o ../aprsupdater-daemon .
./aprsupdater-daemon
# Open http://localhost:8080/
```

## Build Targets (Cross-compile)
GitHub Actions automatically builds on tag push:
- `aprsupdater-daemon-x86_64` - Linux x86_64
- `aprsupdater-daemon-arm64` - Linux ARM64 (Raspberry Pi 3+/4/5)
- `aprsupdater-daemon-arm32` - Linux ARM32 (Raspberry Pi Zero/1/2/3)
- `aprsupdater-daemon-windows-x86_64.exe` - Windows x86_64

## Deployment (Raspberry Pi)
```bash
# Copy binary and config
scp aprsupdater-daemon-arm64 pi@192.168.1.100:/home/pi/aprsupdater-daemon
scp ~/.aprsupdater.json pi@192.168.1.100:/home/pi/aprsupdater.json

# On Pi:
chmod +x aprsupdater-daemon
./aprsupdater-daemon
# Access web UI: http://pi-ip:8080/
```

## Bug Fixes (GUI Version)
- ✅ SendObject now includes description field
- ✅ APRS symbol table corrected per spec
- ✅ Config migration for stale symbol names

## License
MIT
