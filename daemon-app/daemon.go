package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	APIKey   string         `json:"api_key,omitempty"`
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

type APRSClient struct {
	conn net.Conn
}

func (a *APRSClient) Connect(server, port, callsign, passcode string) error {
	addr := server + ":" + port
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect failed: %v", err)
	}
	a.conn = conn

	// Read welcome message
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, _ := conn.Read(buf)

	// Login to APRS-IS
	login := fmt.Sprintf("user %s pass %s vers APRSupdater 1.0\r\n", callsign, passcode)
	conn.Write([]byte(login))

	// Wait for login response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err = conn.Read(buf)
	if err != nil {
		a.Close()
		return fmt.Errorf("login failed: %v", err)
	}
	resp := string(buf[:n])
	if strings.Contains(resp, "unverified") || strings.Contains(resp, "invalid") {
		a.Close()
		return fmt.Errorf("authentication failed")
	}
	log.Printf("APRS-IS connected: %s", server)
	return nil
}

func (a *APRSClient) SendObject(callsign, objName, lat, lon, symbol, desc string) error {
	if a.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Format: ;OBJNAME*tsLAT[table]LON[code][desc]\r\n
	// Example: ;VK5ARC*011423z3517.10S/13828.39EeShack\r\n

	// Pad object name to 9 chars
	objNamePadded := fmt.Sprintf("%-9s", objName)[:9]

	// Get current timestamp
	now := time.Now().UTC()
	ts := now.Format("021504") + "z" // DDMMSSz

	// Format lat/lon
	latF, err := formatAPRSLat(lat)
	if err != nil {
		return fmt.Errorf("invalid lat: %v", err)
	}
	lonF, err := formatAPRSLon(lon)
	if err != nil {
		return fmt.Errorf("invalid lon: %v", err)
	}

	// Build packet
	body := fmt.Sprintf(";%s*%s%s%s%s%s%s",
		objNamePadded, ts, latF, symbol[:1], lonF, symbol[1:], desc)

	// Full packet with source and path
	packet := fmt.Sprintf("%s>APRS,TCPIP*:%s\r\n", callsign, body)

	log.Printf("SEND: %s", packet)
	_, err = a.conn.Write([]byte(packet))
	return err
}

func (a *APRSClient) Close() {
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
	}
}

func formatAPRSLat(latStr string) (string, error) {
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return "", err
	}
	hemi := "N"
	if lat < 0 {
		hemi = "S"
		lat = -lat
	}
	deg := int(lat)
	min := (lat - float64(deg)) * 60
	return fmt.Sprintf("%02d%05.2f%s", deg, min, hemi), nil
}

func formatAPRSLon(lonStr string) (string, error) {
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return "", err
	}
	hemi := "E"
	if lon < 0 {
		hemi = "W"
		lon = -lon
	}
	deg := int(lon)
	min := (lon - float64(deg)) * 60
	return fmt.Sprintf("%03d%05.2f%s", deg, min, hemi), nil
}

func configPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "aprsupdater.json"
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, "aprsupdater.json")
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
