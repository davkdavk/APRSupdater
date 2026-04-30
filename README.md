# APRS Updater

APRS object packet updater with GUI (Fyne) and headless daemon with web UI.

## Features
- GUI version (main.go) - Desktop app with Fyne toolkit
- Daemon version (daemon-app/) - Headless with web UI at http://:8080/
- Sends APRS object packets via APRS-IS
- Config file: `~/.aprsupdater.json`
- Cross-compiled for x86_64, ARM64, ARM32

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

## Build for Raspberry Pi (Cross-compile)
GitHub Actions automatically builds on tag push:
```bash
git tag v1.0.0
git push origin v1.0.0
```

## API (Daemon)
- `GET /api/config` - Get/Update config
- `GET/POST/DELETE /api/objects/{idx}` - Manage objects
- `POST /api/send` - Send all enabled objects
- `POST /api/daemon/start` - Start daemon
- `POST /api/daemon/stop` - Stop daemon

## Bug Fixes Applied
- SendObject now includes description field
- APRS symbol table corrected per spec
- Config migration for stale symbol names

## License
MIT
