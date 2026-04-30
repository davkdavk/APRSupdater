package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "sync"
    "time"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
)

const (
    DefaultServer   = "rotate.aprs2.net"
    DefaultPort     = "14580"
    DefaultInterval = 25
    MaxObjects      = 10
    APRSFiAPI       = "https://api.aprs.f.i/api/get"
)

// APRS Symbol Tables (from APRSpedia)
// Primary Table (/): /L = Logged-ON user (PC), /l = Laptop
// Alternate Table (\): \L = Lighthouse, \l = Areas
var aprsSymbols = map[string]string{
    "House (Primary)":       "/-",   // Primary = House
    "House w/ HF (Alt)":     "\\-",  // Alternate = House w/ HF
    "Red Cross (Primary)":    "/&",  // Primary = Red Cross
    "Red Cross (Alt)":       "\\+",  // Alternate = Red Cross
    "Helicopter (Primary)":   "/X",  // Primary = Helicopter
    "Helicopter (Alt)":      "\\X",  // Alternate = Helicopter
    "Plane (Primary)":        "/'",  // Primary = Small Aircraft
    "Plane (Alt)":           "\\'",  // Alternate = Small Aircraft
    "Car (Primary)":         "/>",  // Primary = Car
    "Car (Alt)":            "\\>",  // Alternate = Car
    "Boat (Alt)":           "\\b",  // Alternate = Power Boat (top view)
    "Sailboat (Alt)":       "\\/",  // Alternate = Yacht/Sailboat
    "Lighthouse (Alt)":     "\\L",  // Alternate = Lighthouse (UPPERCASE L!)
    "RV (Primary)":          "/y",  // Primary = Yagi @ QTH (closest to RV)
    "RV (Alt)":             "\\y",  // Alternate = Rec Vehicle
    "Ambulance (Alt)":      "\\a",  // Alternate = Ambulance
    "Antenna (Alt)":        "\\#",  // Alternate = HF Antenna
    "Repeater Tower":        "/r",  // Primary = Repeater Tower (lowercase r!)
    "Radio Tower (Primary)": "/R",  // Primary = Radio Tower (UPPERCASE R!)
    "Shack w/ Antenna":     "\\Y",  // Alternate = House @ QTH (Shack)
    "Restaurant (Alt)":     "\\R",  // Alternate = Restaurant
    "Bicycle (Alt)":        "\\d",  // Alternate = Bicycle
    "Fire (Alt)":            "\\f",  // Alternate = Fire
    "Church (Alt)":          "\\c",  // Alternate = Church
    "School (Alt)":          "\\s",  // Alternate = School (top view)
    "Hospital (Alt)":        "\\h",  // Alternate = Hospital
    "Police (Alt)":          "\\p",  // Alternate = Police
    "Marker (Alt)":          "\\.",  // Alternate = DOT
    "Circle (Alt)":          "\\o",  // Alternate = Circle
}

type ObjectConfig struct {
    Name        string `json:"name"`
    Symbol      string `json:"symbol"`
    Latitude    string `json:"latitude"`
    Longitude   string `json:"longitude"`
    Description string `json:"description"`
    Enabled     bool   `json:"enabled"`
}

type Config struct {
    Callsign     string         `json:"callsign"`
    Passcode     string         `json:"passcode"`
    Server       string         `json:"server"`
    Port         string         `json:"port"`
    Interval     int            `json:"interval"`
    APRSFiAPIKey string         `json:"aprsfi_api_key"`
    Objects      []ObjectConfig `json:"objects"`
}

type APRSClient struct {
    conn net.Conn
}

func (c *APRSClient) Connect(server, port, callsign, passcode string) error {
    addr := net.JoinHostPort(server, port)
    conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
    if err != nil {
        return fmt.Errorf("connection failed: %v", err)
    }
    c.conn = conn
    login := fmt.Sprintf("user %s pass %s vers APRSUpdater 1.0\n", callsign, passcode)
    _, err = fmt.Fprint(conn, login)
    if err != nil {
        return fmt.Errorf("login failed: %v", err)
    }
    conn.SetReadDeadline(time.Now().Add(5 * time.Second))
    scanner := bufio.NewScanner(conn)
    if scanner.Scan() {
        resp := scanner.Text()
        if strings.Contains(resp, "unverified") || strings.Contains(resp, "invalid") {
            return fmt.Errorf("login rejected: %s", resp)
        }
    }
    conn.SetReadDeadline(time.Time{})
    return nil
}

func (c *APRSClient) SendObject(callsign, objName, lat, lon, symbol, desc string) error {
    if c.conn == nil {
        return fmt.Errorf("not connected")
    }
    now := time.Now().UTC()
    ts := fmt.Sprintf("%02d%02d%02dz", now.Day(), now.Hour(), now.Minute())
    latF := formatAPRSLat(lat)
    lonF := formatAPRSLon(lon)
    objNamePadded := fmt.Sprintf("%-9s", objName)
    if len(objNamePadded) > 9 {
        objNamePadded = objNamePadded[:9]
    }
    // Build packet body: ;OBJNAME*TS LAT table LON code desc
    body := ";" + objNamePadded + "*" + ts + latF + symbol[:1] + lonF + symbol[1:] + desc
    packet := fmt.Sprintf("%s>APRS,TCPIP*:%s\r\n", callsign, body)
    log.Printf("SEND: name=%s symbol=%q (table=%q code=%q)", objName, symbol, symbol[:1], symbol[1:])
    log.Printf("SEND: packet=%q", packet)
    _, err := fmt.Fprint(c.conn, packet)
    return err
}

func (c *APRSClient) Close() {
    if c.conn != nil {
        c.conn.Close()
    }
}

func verifyOnAPRSFi(apiKey, name string) bool {
    if apiKey == "" {
        return false
    }
    url := fmt.Sprintf("%s?name=%s&what=objects&apikey=%s&format=json",
        APRSFiAPI, name, apiKey)
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    var result struct {
        Result string `json:"result"`
        Found  []struct {
            Name string `json:"name"`
        } `json:"found"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return false
    }
    if result.Result != "ok" {
        return false
    }
    for _, f := range result.Found {
        if strings.Contains(f.Name, name) {
            return true
        }
    }
    return false
}

func formatAPRSLat(latStr string) string {
    lat, _ := strconv.ParseFloat(latStr, 64)
    hemi := "N"
    if lat < 0 {
        lat = -lat
        hemi = "S"
    }
    deg := int(lat)
    min := (lat - float64(deg)) * 60
    return fmt.Sprintf("%02d%05.2f%s", deg, min, hemi)
}

func formatAPRSLon(lonStr string) string {
    lon, _ := strconv.ParseFloat(lonStr, 64)
    hemi := "E"
    if lon < 0 {
        lon = -lon
        hemi = "W"
    }
    deg := int(lon)
    min := (lon - float64(deg)) * 60
    return fmt.Sprintf("%03d%05.2f%s", deg, min, hemi)
}

type Daemon struct {
    mu       sync.Mutex
    running  bool
    stopChan chan struct{}
    wg       sync.WaitGroup
}

func (d *Daemon) Start(interval int, sendFunc func() error, statusFunc func(string)) {
    d.mu.Lock()
    defer d.mu.Unlock()
    if d.running {
        return
    }
    d.running = true
    d.stopChan = make(chan struct{})
    d.wg.Add(1)
    go func() {
        defer d.wg.Done()
        ticker := time.NewTicker(time.Duration(interval) * time.Minute)
        defer ticker.Stop()
        for {
            select {
            case <-d.stopChan:
                return
            case <-ticker.C:
                if err := sendFunc(); err != nil {
                    statusFunc(fmt.Sprintf("Daemon error: %v", err))
                } else {
                    statusFunc(fmt.Sprintf("Updated at %s", time.Now().Format("15:04:05")))
                }
            }
        }
    }()
}

func (d *Daemon) Stop(statusFunc func(string)) {
    d.mu.Lock()
    defer d.mu.Unlock()
    if !d.running {
        return
    }
    close(d.stopChan)
    d.wg.Wait()
    d.running = false
    statusFunc("Daemon stopped")
}

func (d *Daemon) IsRunning() bool {
    d.mu.Lock()
    defer d.mu.Unlock()
    return d.running
}

func getConfigPath() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return "aprsupdater.json"
    }
    return filepath.Join(home, ".aprsupdater.json")
}

func loadConfig() (*Config, error) {
    path := getConfigPath()
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg Config
    err = json.Unmarshal(data, &cfg)
    return &cfg, err
}

func saveConfig(cfg *Config) error {
    path := getConfigPath()
    data, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}

func main() {
    myApp := app.New()
    myApp.SetIcon(nil)
    w := myApp.NewWindow("APRS.f.i Object Updater")
    w.Resize(fyne.NewSize(600, 850))

    cfg := &Config{
        Callsign:     "VK5LEX",
        Passcode:     "21949",
        Server:       DefaultServer,
        Port:         DefaultPort,
        Interval:     DefaultInterval,
        Objects:      make([]ObjectConfig, MaxObjects),
    }
    if loaded, err := loadConfig(); err == nil {
        cfg = loaded
        if len(cfg.Objects) < MaxObjects {
            objs := make([]ObjectConfig, MaxObjects)
            copy(objs, cfg.Objects)
            cfg.Objects = objs
        }
    }

    currentObj := 0
    var daemon *Daemon
    var daemonBtn *widget.Button
    var statusLabel *widget.Label
    var logEntry *widget.Entry

    logEntry = widget.NewMultiLineEntry()
    logEntry.SetText("Ready\n")
    logEntry.Disable()

    statusLabel = widget.NewLabel("Ready")

    callsignEntry := widget.NewEntry()
    callsignEntry.SetText(cfg.Callsign)

    passcodeEntry := widget.NewEntry()
    passcodeEntry.SetText(cfg.Passcode)
    passcodeEntry.Password = true

    serverEntry := widget.NewEntry()
    serverEntry.SetText(cfg.Server)

    portEntry := widget.NewEntry()
    portEntry.SetText(cfg.Port)

    apiKeyEntry := widget.NewEntry()
    apiKeyEntry.SetText(cfg.APRSFiAPIKey)
    apiKeyEntry.SetPlaceHolder("aprs.f.i API key (for verification)")

    objNameEntry := widget.NewEntry()
    objNameEntry.SetPlaceHolder("Object name (max 9 chars)")

    descEntry := widget.NewMultiLineEntry()
    descEntry.SetPlaceHolder("Object description")

    enabledCheck := widget.NewCheck("Enabled (send with 'Send All Objects')", nil)
    enabledCheck.Checked = true

    latEntry := widget.NewEntry()
    latEntry.SetPlaceHolder("Latitude (e.g. -34.9285)")

    lonEntry := widget.NewEntry()
    lonEntry.SetPlaceHolder("Longitude (e.g. 138.6007)")

    symbolKeys := make([]string, 0, len(aprsSymbols))
    for k := range aprsSymbols {
        symbolKeys = append(symbolKeys, k)
    }
    sort.Strings(symbolKeys)
    symbolSelect := widget.NewSelect(symbolKeys, nil)
    symbolSelect.Selected = "House (Primary)"

    intervalEntry := widget.NewEntry()
    intervalEntry.SetText(fmt.Sprintf("%d", cfg.Interval))

    objectSelect := widget.NewSelect(nil, nil)
    for i := 0; i < MaxObjects; i++ {
        label := fmt.Sprintf("Object %d", i+1)
        if cfg.Objects[i].Name != "" {
            label = fmt.Sprintf("Object %d: %s", i+1, cfg.Objects[i].Name)
            if !cfg.Objects[i].Enabled {
                label += " [Disabled]"
            }
        }
        objectSelect.Options = append(objectSelect.Options, label)
    }
    objectSelect.Selected = objectSelect.Options[0]

    updateLog := func(msg string) {
        logEntry.Enable()
        logEntry.SetText(logEntry.Text + msg + "\n")
        logEntry.Disable()
        statusLabel.SetText(msg)
    }

    loadObject := func(idx int) {
        obj := cfg.Objects[idx]
        objNameEntry.SetText(obj.Name)
        latEntry.SetText(obj.Latitude)
        lonEntry.SetText(obj.Longitude)
        descEntry.SetText(obj.Description)
        enabledCheck.Checked = obj.Enabled
        enabledCheck.Refresh()

        // Bug 3 fix: migrate old symbol names
        if _, ok := aprsSymbols[obj.Symbol]; !ok && obj.Symbol != "" {
            // Handle renamed symbols
            switch obj.Symbol {
            case "Radio Tower (Primary)":
                obj.Symbol = "Repeater Tower"
            case "Lighthouse", "Lighthouse (Primary)":
                obj.Symbol = "Lighthouse (Alt)"
            default:
                updateLog(fmt.Sprintf("WARNING: symbol %q not recognised, defaulting to House (Primary)", obj.Symbol))
                obj.Symbol = "House (Primary)"
            }
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

    saveObject := func(idx int) {
        cfg.Objects[idx] = ObjectConfig{
            Name:        objNameEntry.Text,
            Symbol:      symbolSelect.Selected,
            Latitude:    latEntry.Text,
            Longitude:   lonEntry.Text,
            Description: descEntry.Text,
            Enabled:     enabledCheck.Checked,
        }
        label := fmt.Sprintf("Object %d", idx+1)
        if objNameEntry.Text != "" {
            label = fmt.Sprintf("Object %d: %s", idx+1, objNameEntry.Text)
            if !enabledCheck.Checked {
                label += " [Disabled]"
            }
        }
        objectSelect.Options[idx] = label
        objectSelect.Selected = label
        objectSelect.Refresh()
        updateLog(fmt.Sprintf("Saved object %d", idx+1))
    }

    loadObject(0)

    objectSelect.OnChanged = func(sel string) {
        saveObject(currentObj)
        for i := range objectSelect.Options {
            if objectSelect.Options[i] == sel {
                currentObj = i
                loadObject(i)
                objectSelect.Selected = sel
                objectSelect.Refresh()
                return
            }
        }
    }

    saveObjBtn := widget.NewButton("Save This Object", func() {
        saveObject(currentObj)
    })

    removeObjBtn := widget.NewButton("Remove Object", func() {
        cfg.Objects[currentObj] = ObjectConfig{}
        objectSelect.Options[currentObj] = fmt.Sprintf("Object %d", currentObj+1)
        objectSelect.Selected = objectSelect.Options[currentObj]
        objectSelect.Refresh()
        loadObject(currentObj)
        if err := saveConfig(cfg); err != nil {
            updateLog(fmt.Sprintf("Save error: %v", err))
        } else {
            updateLog(fmt.Sprintf("Removed object %d", currentObj+1))
        }
    })

    saveConfigBtn := widget.NewButton("Save Config", func() {
        saveObject(currentObj)
        cfg.Callsign = callsignEntry.Text
        cfg.Passcode = passcodeEntry.Text
        cfg.Server = serverEntry.Text
        cfg.Port = portEntry.Text
        cfg.APRSFiAPIKey = apiKeyEntry.Text
        cfg.Interval, _ = strconv.Atoi(intervalEntry.Text)
        // Clean empty objects
        clean := []ObjectConfig{}
        for _, o := range cfg.Objects {
            if o.Name != "" {
                clean = append(clean, o)
            }
        }
        for len(clean) < MaxObjects {
            clean = append(clean, ObjectConfig{})
        }
        cfg.Objects = clean
        if err := saveConfig(cfg); err != nil {
            updateLog(fmt.Sprintf("Save error: %v", err))
        } else {
            updateLog("Config saved!")
        }
    })

    var sendOnce func() error
    var client *APRSClient

    sendOnce = func() error {
        saveObject(currentObj)
        if err := saveConfig(cfg); err != nil {
            updateLog(fmt.Sprintf("Save error: %v", err))
        }

        if client != nil {
            client.Close()
        }
        client = &APRSClient{}
        err := client.Connect(serverEntry.Text, portEntry.Text, callsignEntry.Text, passcodeEntry.Text)
        if err != nil {
            updateLog(fmt.Sprintf("Connect failed: %v", err))
            return err
        }

        var lastErr error
        for i, obj := range cfg.Objects {
            if obj.Name == "" || !obj.Enabled {
                continue
            }
            // Migrate old symbol names
            if _, ok := aprsSymbols[obj.Symbol]; !ok && obj.Symbol != "" {
                switch obj.Symbol {
                case "Radio Tower (Primary)":
                    obj.Symbol = "Repeater Tower"
                case "Lighthouse", "Lighthouse (Primary)":
                    obj.Symbol = "Lighthouse (Alt)"
                default:
                    lastErr = fmt.Errorf("invalid symbol for object %d: %s", i+1, obj.Symbol)
                    continue
                }
            }
            symbol := aprsSymbols[obj.Symbol]
            if len(symbol) < 2 {
                lastErr = fmt.Errorf("invalid symbol for object %d", i+1)
                continue
            }
            updateLog(fmt.Sprintf("Sending object: %s...", obj.Name))
            err = client.SendObject(callsignEntry.Text, obj.Name, obj.Latitude, obj.Longitude, symbol, obj.Description)
            if err != nil {
                lastErr = fmt.Errorf("send failed for %s: %v", obj.Name, err)
                updateLog(fmt.Sprintf("Send error: %v", lastErr))
            }
            time.Sleep(1 * time.Second)
        }

        client.Close()
        client = nil

        if cfg.APRSFiAPIKey != "" {
            updateLog("Verifying on aprs.f.i...")
            for _, obj := range cfg.Objects {
                if obj.Name == "" {
                    continue
                }
                updateLog(fmt.Sprintf("Checking %s on aprs.f.i...", obj.Name))
                found := false
                for i := 0; i < 5; i++ {
                    if verifyOnAPRSFi(cfg.APRSFiAPIKey, obj.Name) {
                        updateLog(fmt.Sprintf("Verified: %s on aprs.f.i!", obj.Name))
                        found = true
                        break
                    }
                    time.Sleep(2 * time.Second)
                }
                if !found {
                    updateLog(fmt.Sprintf("WARNING: %s not found on aprs.f.i", obj.Name))
                }
            }
        }
        return lastErr
    }

    sendBtn := widget.NewButton("Send All Objects", func() {
        go func() {
            updateLog("Starting send...")
            if err := sendOnce(); err != nil {
                fyne.CurrentApp().SendNotification(&fyne.Notification{
                    Title:   "APRS Updater",
                    Content: fmt.Sprintf("Error: %v", err),
                })
            } else {
                fyne.CurrentApp().SendNotification(&fyne.Notification{
                    Title:   "APRS Updater",
                    Content: "All objects sent and verified!",
                })
            }
        }()
    })

    daemon = &Daemon{}
    daemonBtn = widget.NewButton("Start Daemon", func() {
        if daemon.IsRunning() {
            daemon.Stop(func(msg string) {
                updateLog(msg)
                daemonBtn.SetText("Start Daemon")
                daemonBtn.Refresh()
            })
            return
        }
        interval, err := strconv.Atoi(intervalEntry.Text)
        if err != nil || interval < 1 {
            updateLog("Invalid interval (must be >= 1 minute)")
            return
        }
        updateLog(fmt.Sprintf("Daemon starting (every %d min)...", interval))
        daemonBtn.SetText("Stop Daemon")
        daemonBtn.Refresh()
        daemon.Start(interval, func() error {
            return sendOnce()
        }, func(msg string) {
            updateLog(msg)
        })
    })

    connForm := container.NewVBox(
        widget.NewLabel("APRS-IS Connection"),
        container.NewGridWithColumns(2,
            widget.NewLabel("Callsign:"), callsignEntry,
            widget.NewLabel("Passcode:"), passcodeEntry,
            widget.NewLabel("Server:"), serverEntry,
            widget.NewLabel("Port:"), portEntry,
        ),
        widget.NewLabel("aprs.f.i API Key (for verification):"),
        apiKeyEntry,
    )

    objForm := container.NewVBox(
        container.NewGridWithColumns(2,
            widget.NewLabel("Object Slot:"), objectSelect,
            widget.NewLabel("Name:"), objNameEntry,
            widget.NewLabel("Icon:"), symbolSelect,
            widget.NewLabel("Latitude:"), latEntry,
            widget.NewLabel("Longitude:"), lonEntry,
        ),
        enabledCheck,
        widget.NewLabel("Description:"),
        descEntry,
        container.NewHBox(saveObjBtn, removeObjBtn, saveConfigBtn),
    )

    daemonForm := container.NewVBox(
        widget.NewLabel("Daemon Mode"),
        container.NewGridWithColumns(2,
            widget.NewLabel("Interval (min):"), intervalEntry,
        ),
    )

    w.SetContent(container.NewVBox(
        connForm,
        widget.NewSeparator(),
        objForm,
        widget.NewSeparator(),
        daemonForm,
        widget.NewSeparator(),
        container.NewHBox(sendBtn, daemonBtn),
        statusLabel,
        widget.NewSeparator(),
        widget.NewLabel("Log:"),
        container.NewScroll(logEntry),
    ))

    w.SetCloseIntercept(func() {
        dialog.ShowConfirm("Minimize", "Hide window but keep daemon running?", func(ok bool) {
            if ok {
                saveObject(currentObj)
                saveConfig(cfg)
                w.Hide()
            } else {
                saveObject(currentObj)
                saveConfig(cfg)
                daemon.Stop(func(msg string) {})
                myApp.Quit()
            }
        }, w)
    })

    w.ShowAndRun()
}
