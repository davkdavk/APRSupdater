# APRS Updater Daemon

Headless daemon version of APRS Updater with web UI for managing APRS objects.

## Features
- No GUI dependencies (pure Go + web interface)
- LAN-accessible web UI at http://<pi-ip>:8080/
- HTTP API for config/object management
- Sends APRS object packets via APRS-IS
- Config stored in `~/.aprsupdater.json` (shared with GUI version)

## API Endpoints
- `GET /` - Web UI
- `GET /api/config` - Get config
- `POST /api/config` - Update config
- `GET/POST/DELETE /api/objects/{idx}` - Manage objects (0-9)
- `POST /api/send` - Send all enabled objects
- `POST /api/daemon/start` - Start daemon (periodic sending)
- `POST /api/daemon/stop` - Stop daemon
- `GET /api/status` - Get daemon status

## Building
```bash
cd daemon-app
go build -o ../aprsupdater-daemon .
```

## Cross-Compilation
GitHub Actions automatically builds for:
- linux/amd64 (x86_64)
- linux/arm64 (Raspberry Pi 3+/4/5)
- linux/arm/v7 (Raspberry Pi Zero/1/2/3)

Push a tag to trigger release:
```bash
git tag v1.0.0
git push origin v1.0.0
```

## Usage
1. Start daemon: `./aprsupdater-daemon`
2. Open browser: `http://localhost:8080`
3. Configure callsign, passcode, server
4. Add/edit objects
5. Click "Send Now" or "Start Daemon"
