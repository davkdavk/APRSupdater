# APRS Updater - Build Rules

## Overview
Build a headless daemon version with web UI, cross-compiled for multiple platforms via GitHub Actions.

## Build Targets
- **x86_64** (Linux PC)
- **ARM64** (Raspberry Pi 3+/4/5)
- **ARM32** (Raspberry Pi Zero/1/2/3)

## Step-by-Step Plan

### Step 1: Create `main-daemon.go`
- Start from working `main.go` baseline
- Remove Fyne imports (no GUI)
- Add `html/template` and `net/http` imports
- Keep all APRS logic (APRSClient, config, symbols)

### Step 2: Add Web UI
- Create `web.html` with same layout as Fyne GUI
  - Connection settings (callsign, passcode, server, port, interval)
  - Object editor (name, symbol dropdown, lat/lon, description, enabled checkbox)
  - Action buttons (Send All, Start/Stop Daemon)
  - Log display
- Or embed as Go string constant `webTemplate`

### Step 3: Add Web Server to `main()`
- Serve `web.html` at `/`
- API endpoints:
  - `GET /api/config` - Get/Update config
  - `GET/POST/DELETE /api/objects/:id` - Manage objects
  - `POST /api/send` - Send all enabled objects
  - `POST /api/daemon/start` - Start daemon
  - `POST /api/daemon/stop` - Stop daemon
- Listen on port `8080`

### Step 4: GitHub Actions Workflow
Create `.github/workflows/build.yml`:

```yaml
name: Build APRS Daemon

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goos: linux
            goarch: amd64
            output: aprsupdater-daemon-x86_64
          - goos: linux
            goarch: arm64
            output: aprsupdater-daemon-arm64
          - goos: linux
            goarch: arm
            goarm: 7
            output: aprsupdater-daemon-arm32
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install cross-compilation tools
        run: |
          sudo apt-get update
          sudo apt-get install -y \
            gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
            gcc-arm-linux-gnueabihf \
            libgl1-mesa-dev-arm64-cross \
            libx11-dev-arm64-cross \
            libxrandr-dev-arm64-cross \
            libxxf86vm-dev-arm64-cross \
            libxi-dev-arm64-cross \
            libxcursor-dev-arm64-cross \
            libgl1-mesa-dev-armhf-cross \
            libx11-dev-armhf-cross \
            libxrandr-dev-armhf-cross \
            libxxf86vm-dev-armhf-cross \
            libxi-dev-armhf-cross \
            libxcursor-dev-armhf-cross
      
      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          GOARM: ${{ matrix.goarm }}
          CGO_ENABLED: 1
          CC: ${{ matrix.goarch == 'arm64' && 'aarch64-linux-gnu-gcc' || matrix.goarch == 'arm' && 'arm-linux-gnueabihf-gcc' || '' }}
        run: |
          go mod tidy
          go build -o ${{ matrix.output }} main-daemon.go
      
      - name: Upload Release Asset
        uses: softprops/action-gh-release@v1
        if: startsWith(github.ref, 'refs/tags/')
        with:
          files: ${{ matrix.output }}
          generate_release_notes: true
```

### Step 5: Create Release
```bash
git tag v1.0.0
git push origin v1.0.0
# GitHub Actions builds all 3 binaries and creates release
```

## File Structure
```
APRSupdater/
├── main.go              # GUI version (Fyne) - baseline
├── main-daemon.go       # Headless daemon + web UI
├── web.html             # Web UI template
├── go.mod
├── go.sum
├── baseline/            # Protected backups
└── .github/
    └── workflows/
        └── build.yml      # Cross-compilation workflow
```

## Key Points
- ✅ `main.go` stays untouched (working baseline)
- ✅ `main-daemon.go` is separate (no Fyne, pure Go + web)
- ✅ GitHub Actions handles cross-compilation (no local setup needed)
- ✅ Release on tag creates 3 binaries automatically
