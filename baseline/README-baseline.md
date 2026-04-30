# APRS Updater - BASELINE (Stable Release)
**Date:** April 30, 2026
**Status:** WORKING - Do not modify without testing

## What's in this baseline:
- ✅ `main.go.baseline` - Working source with all bug fixes
- ✅ `aprsupdater.baseline` - Compiled binary (Linux AMD64)
- ✅ `aprsupdater-config.baseline.json` - Working config (VK5ARC, VK5RSV, VK5ARC/P)
- ✅ `aprsupdater-pi-install.sh` - Self-contained Pi installer (ONE FILE)
- ✅ `build-on-pi.sh` - Alternative Pi build script

## Bug fixes included:
1. ✅ SendObject fmt.Sprintf fix (was dropping description)
2. ✅ APRS symbol table corrected (Repeater Tower = /r, Lighthouse = \L)
3. ✅ Symbol migration in loadObject (handles old "Radio Tower" key)
4. ✅ Enable/Disable checkboxes for objects
5. ✅ Lighthouse icon working on aprs.f.i

## Verified working symbols:
- VK5ARC: House w/ HF (Alt) `\-`
- VK5RSV: Repeater Tower `/r`
- VK5ARC/P: Lighthouse (Alt) `\L` ✅ (confirmed on aprs.f.i)

## To restore from baseline:
```bash
cd /home/davey/Desktop/APRSupdater
cp baseline/main.go.baseline main.go
cp baseline/aprsupdater.baseline aprsupdater
# OR rebuild: go build -o aprsupdater .
```

## Pi Installation:
Just give someone `aprsupdater-pi-install.sh` - it's self-contained!
