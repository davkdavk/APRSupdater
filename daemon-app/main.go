package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"embed"
)

var daemon = &Daemon{}
var configMu sync.Mutex

//go:embed web.html
var webFS embed.FS

func main() {
	// Load config or use defaults
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("Using defaults: %v", err)
		cfg = Config{
			Callsign: "VK5LEX",
			Passcode: "21949",
			Server:   DefaultServer,
			Port:     DefaultPort,
			Interval: DefaultInterval,
			Objects:  make([]ObjectConfig, MaxObjects),
		}
	}
	_ = cfg // Will be used by API handlers via loadConfig()

	// Serve web UI from embedded filesystem
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		data, err := fs.ReadFile(webFS, "web.html")
		if err != nil {
			http.Error(w, "web.html not found", 404)
			return
		}
		w.Write(data)
	})

	// API: Get/Update config
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			configMu.Lock()
			cfg, _ := loadConfig()
			configMu.Unlock()
			json.NewEncoder(w).Encode(cfg)
			return
		}
		if r.Method == http.MethodPost {
			var c Config
			if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if len(c.Objects) < MaxObjects {
				for i := len(c.Objects); i < MaxObjects; i++ {
					c.Objects = append(c.Objects, ObjectConfig{})
				}
			}
			if err := saveConfig(c); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		http.Error(w, "Method not allowed", 405)
	})

	// API: Manage individual objects
	http.HandleFunc("/api/objects/", func(w http.ResponseWriter, r *http.Request) {
		// Extract index from path: /api/objects/0
		idx := r.URL.Path[len("/api/objects/"):]
		if idx == "" {
			http.Error(w, "Missing object index", 400)
			return
		}
		// Parse index
		var i int
		if _, err := fmt.Sscanf(idx, "%d", &i); err != nil || i < 0 || i >= MaxObjects {
			http.Error(w, "Invalid object index", 400)
			return
		}

		if r.Method == http.MethodGet {
			cfg, _ := loadConfig()
			if i >= len(cfg.Objects) {
				http.Error(w, "Object not found", 404)
				return
			}
			json.NewEncoder(w).Encode(cfg.Objects[i])
			return
		}
		if r.Method == http.MethodPost {
			var obj ObjectConfig
			if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			cfg, _ := loadConfig()
			for len(cfg.Objects) <= i {
				cfg.Objects = append(cfg.Objects, ObjectConfig{})
			}
			cfg.Objects[i] = obj
			saveConfig(cfg)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		if r.Method == http.MethodDelete {
			cfg, _ := loadConfig()
			if i < len(cfg.Objects) {
				cfg.Objects[i] = ObjectConfig{}
				saveConfig(cfg)
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		http.Error(w, "Method not allowed", 405)
	})

	// API: Send all enabled objects
	http.HandleFunc("/api/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", 405)
			return
		}
		cfg, err := loadConfig()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		go sendAllObjects(cfg)
		json.NewEncoder(w).Encode(map[string]string{"status": "sending"})
	})

	// API: Start daemon
	http.HandleFunc("/api/daemon/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", 405)
			return
		}
		cfg, err := loadConfig()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if err := daemon.Start(cfg); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "started"})
	})

	// API: Stop daemon
	http.HandleFunc("/api/daemon/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", 405)
			return
		}
		daemon.Stop()
		json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
	})

	// API: Status
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"daemon_running": daemon.IsRunning(),
		})
	})

	log.Println("Daemon HTTP server starting on :8080 (LAN accessible)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
