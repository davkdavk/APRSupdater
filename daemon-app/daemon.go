package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sync"
    "time"
)

type ObjectConfig struct {
    Name        string `json:"name"`
    Symbol      string `json:"symbol"`
    Latitude    string `json:"latitude"`
    Longitude   string `json:"longitude"`
    Description string `json:"description"`
    Enabled     bool   `json:"enabled"`
}

type Config struct {
    Callsign string         `json:"callsign"`
    Passcode string         `json:"passcode"`
    Server   string         `json:"server"`
    Port     string         `json:"port"`
    Interval int            `json:"interval"`
    Objects  []ObjectConfig `json:"objects"`
}

var aprsSymbols = map[string]string{
    "House (Primary)":       "/-",
    "House w/ HF (Alt)":     "\\-",
    "Red Cross (Primary)":    "/&",
    "Red Cross (Alt)":       "\\+",
    "Helicopter (Primary)":   "/X",
    "Helicopter (Alt)":      "\\X",
    "Plane (Primary)":        "/'",
    "Plane (Alt)":           "\\'",
    "Car (Primary)":         "/>",
    "Car (Alt)":            "\\>",
    "Boat (Alt)":           "\\b",
    "Sailboat (Alt)":       "\\/",
    "Lighthouse (Alt)":     "\\L",
    "RV (Primary)":          "/y",
    "RV (Alt)":             "\\y",
    "Repeater Tower":        "/r",
    "Radio Tower (Primary)": "/R",
    "Shack w/ Antenna":     "\\Y",
    "Restaurant (Alt)":     "\\R",
    "Bicycle (Alt)":        "\\d",
    "Fire (Alt)":            "\\f",
    "Church (Alt)":          "\\c",
    "School (Alt)":          "\\s",
    "Hospital (Alt)":        "\\h",
    "Police (Alt)":          "\\p",
    "Marker (Alt)":          "\\.",
    "Circle (Alt)":          "\\o",
}

type APRSClient struct{ connected bool }
func (a *APRSClient) Connect(server, port, callsign, passcode string) error { a.connected = true; return nil }
func (a *APRSClient) SendObject(callsign, objName, lat, lon, symbol, desc string) error { if !a.connected { return fmt.Errorf("not connected") }; return nil }
func (a *APRSClient) Close() { a.connected = false }

func configPath() string {
    home, err := os.UserHomeDir()
    if err != nil { return ".aprsupdater.json" }
    return filepath.Join(home, ".aprsupdater.json")
}

func loadConfig() (Config, error) {
    p := configPath()
    b, err := os.ReadFile(p)
    if err != nil { return Config{}, err }
    var cfg Config
    if err := json.Unmarshal(b, &cfg); err != nil { return Config{}, err }
    if len(cfg.Objects) < 10 {
        for i := len(cfg.Objects); i < 10; i++ { cfg.Objects = append(cfg.Objects, ObjectConfig{}) }
    }
    return cfg, nil
}

func saveConfig(cfg Config) error {
    p := configPath()
    f, err := os.Create(p)
    if err != nil { return err }
    defer f.Close()
    enc := json.NewEncoder(f); enc.SetIndent("", "  ")
    return enc.Encode(cfg)
}

const (
    DefaultServer   = "rotate.aprs.net"
    DefaultPort     = "14580"
    DefaultInterval = 15
    MaxObjects      = 10
)

type Daemon struct {
    mu      sync.Mutex
    running bool
    stop    chan struct{}
}

func (d *Daemon) Start(cfg Config) error {
    d.mu.Lock()
    defer d.mu.Unlock()
    if d.running {
        return fmt.Errorf("daemon already running")
    }
    d.running = true
    d.stop = make(chan struct{})
    go func() {
        ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Minute)
        defer ticker.Stop()
        for {
            select {
            case <-d.stop:
                return
            case <-ticker.C:
                sendAllObjects(cfg)
            }
        }
    }()
    return nil
}

func (d *Daemon) Stop() {
    d.mu.Lock()
    defer d.mu.Unlock()
    if d.running {
        close(d.stop)
        d.running = false
    }
}

func (d *Daemon) IsRunning() bool {
    d.mu.Lock()
    defer d.mu.Unlock()
    return d.running
}

func sendAllObjects(cfg Config) {
    client := &APRSClient{}
    if err := client.Connect(cfg.Server, cfg.Port, cfg.Callsign, cfg.Passcode); err != nil {
        log.Printf("Connect failed: %v", err)
        return
    }
    defer client.Close()
    for _, obj := range cfg.Objects {
        if !obj.Enabled || obj.Name == "" {
            continue
        }
        symbol := aprsSymbols[obj.Symbol]
        if symbol == "" {
            log.Printf("Invalid symbol for %s, skipping", obj.Name)
            continue
        }
        if err := client.SendObject(cfg.Callsign, obj.Name, obj.Latitude, obj.Longitude, symbol, obj.Description); err != nil {
            log.Printf("Send failed for %s: %v", obj.Name, err)
            continue
        }
        time.Sleep(400 * time.Millisecond)
    }
}

func loadConfigSafe() *Config {
    cfg, err := loadConfig()
    if err != nil {
        return nil
    }
    return &cfg
}
