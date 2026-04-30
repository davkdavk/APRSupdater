# aprsupdater — Symbol Fix Instructions for opencode

## Overview

This is a Go 1.22 Fyne GUI app (`main.go`, single file, module `aprsupdater`) that sends APRS-IS object packets over TCP. There are three bugs causing wrong symbols to appear on APRS maps. Fix all three as described below. Do not change any other logic.

---

## Bug 1 — `SendObject` drops the `desc` field

**Location:** `func (c *APRSClient) SendObject(...)` in `main.go`

**Problem:** The format string has 6 `%s` verbs but 7 arguments. `desc` is the 7th argument and is silently dropped by `fmt.Sprintf`, so every transmitted packet is missing its comment/description field.

**Fix:** Add the missing `%s`:

```go
// BEFORE
body := fmt.Sprintf(";%s*%s%s%s%s%s",
    objNamePadded, ts, latF, symbol[:1], lonF, symbol[1:], desc)

// AFTER
body := fmt.Sprintf(";%s*%s%s%s%s%s%s",
    objNamePadded, ts, latF, symbol[:1], lonF, symbol[1:], desc)
```

---

## Bug 2 — Incorrect entries in `aprsSymbols` map

**Location:** `var aprsSymbols` in `main.go`

**Problem:** Several symbol codes are wrong per the APRS symbol spec:
- `Power Boat` is mapped to `\b` (which is Bicycle) — should be `\^`
- `Sailboat` is mapped to `\/` (invalid) — should be `\s`
- `RV` is mapped to `\y` (Skywarn) — should be `\R`
- `Repeater Tower` should be `/r` (lowercase)

**Fix:** Replace the entire `aprsSymbols` map with:

```go
var aprsSymbols = map[string]string{
    // Primary table '/'
    "House (Primary)":        "/-",
    "Red Cross (Primary)":    "/+",
    "Helicopter (Primary)":   "/X",
    "Plane (Primary)":        "/'",
    "Car (Primary)":          "/>",
    "Repeater Tower":         "/r",
    "Yagi @ QTH (Primary)":  "/Y",

    // Alternate table '\'
    "House w/ HF (Alt)":      "\\-",
    "Red Cross (Alt)":        "\\+",
    "Helicopter (Alt)":       "\\X",
    "Plane (Alt)":            "\\'",
    "Car (Alt)":              "\\>",
    "Power Boat (Alt)":       "\\^",
    "Sailboat (Alt)":         "\\s",
    "RV (Alt)":               "\\R",
    "Ambulance (Alt)":        "\\a",
    "HF Antenna (Alt)":       "\\#",
    "Shack w/ Antenna (Alt)": "\\Y",
    "Restaurant (Alt)":       "\\r",
    "Bicycle (Alt)":          "\\b",
    "Fire (Alt)":             "\\f",
    "Church (Alt)":           "\\c",
    "School (Alt)":           "\\k",
    "Hospital (Alt)":         "\\h",
    "Police (Alt)":           "\\p",
    "Marker (Alt)":           "\\.",
    "Circle (Alt)":           "\\o",
}
```

---

## Bug 3 — Stale symbol names in saved config crash silently

**Location:** `loadObject` function in `main.go`

**Problem:** Config JSON stores the human-readable symbol name (e.g. `"RV (Primary)"`). After the map keys are renamed in Bug 2, any saved config with old key names will fail the `aprsSymbols[obj.Symbol]` lookup silently — returning `""` — causing the object to be skipped entirely and nothing transmitted.

**Fix:** After reading `obj.Symbol` from config, check it exists in the map. If not, warn and default to `"House (Primary)"`:

```go
loadObject := func(idx int) {
    obj := cfg.Objects[idx]
    objNameEntry.SetText(obj.Name)
    latEntry.SetText(obj.Latitude)
    lonEntry.SetText(obj.Longitude)
    descEntry.SetText(obj.Description)

    // Validate saved symbol name still exists in map
    if _, ok := aprsSymbols[obj.Symbol]; !ok && obj.Symbol != "" {
        updateLog(fmt.Sprintf("WARNING: symbol %q not recognised, defaulting to House (Primary)", obj.Symbol))
        obj.Symbol = "House (Primary)"
    }

    symbolSelect.SetSelected("House (Primary)") // safe default
    for i, opt := range symbolKeys {
        if opt == obj.Symbol {
            symbolSelect.SetSelectedIndex(i)
            break
        }
    }
    updateLog(fmt.Sprintf("Loaded object %d: %s", idx+1, obj.Name))
}
```

---

## Verification

After applying all fixes, check the log for a correctly formed packet:

```
SEND: packet="VK5XXX>APRS,TCPIP*:;MYREPEATER*011423z3417.10S/13836.07Er/My repeater\r\n"
```

Confirm:
- Symbol table char appears **between** lat and lon ✅
- Symbol code char appears **immediately after** lon ✅
- Description text appears **after** symbol code ✅
- Packet ends with `\r\n` ✅

---

## Do Not Change

- `go.mod` / `go.sum`
- `.github/workflows/build.yml`
- `formatAPRSLat` / `formatAPRSLon`
- Login / connect / daemon logic
