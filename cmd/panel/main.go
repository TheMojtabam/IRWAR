package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/slipstream-panel/internal/dnstest"
	"github.com/yourusername/slipstream-panel/internal/runner"
	"github.com/yourusername/slipstream-panel/internal/store"
)

var (
	db  *store.Store
	mgr *runner.Manager
)

func main() {
	dir := "/opt/slipstream-panel"
	if v := os.Getenv("PANEL_DIR"); v != "" {
		dir = v
	}
	port := "9090"
	if v := os.Getenv("PANEL_PORT"); v != "" {
		port = v
	}

	os.MkdirAll(dir, 0755)
	os.MkdirAll(filepath.Join(dir, "logs"), 0755)

	var err error
	db, err = store.New(filepath.Join(dir, "instances.json"))
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	mgr = runner.New(filepath.Join(dir, "logs"))
	log.Printf("binary: %s", mgr.Bin())

	for _, inst := range db.List() {
		if inst.AutoRestart {
			if err := mgr.Start(inst); err != nil {
				log.Printf("autostart %s: %v", inst.ID, err)
			}
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(page)
	})
	mux.HandleFunc("/api/status",     apiStatus)
	mux.HandleFunc("/api/instances",  apiInstances)
	mux.HandleFunc("/api/instances/", apiInstance)
	mux.HandleFunc("/api/test",       apiTestRaw)

	log.Printf("panel on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func jw(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func je(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

func apiStatus(w http.ResponseWriter, r *http.Request) {
	b := mgr.Bin()
	if b == "" {
		b = "NOT FOUND"
	}
	jw(w, map[string]string{"bin": b})
}

func apiInstances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		type row struct {
			store.Instance
			Status string `json:"status"`
		}
		list := db.List()
		out := make([]row, len(list))
		for i, v := range list {
			out[i] = row{v, string(mgr.Status(v.ID))}
		}
		jw(w, out)
	case "POST":
		var body store.Instance
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			je(w, 400, "bad json")
			return
		}
		if body.Domain == "" {
			je(w, 400, "domain required")
			return
		}
		if body.SocksPort == 0 {
			body.SocksPort = 1080
		}
		if body.Name == "" {
			body.Name = fmt.Sprintf("Instance %d", len(db.List())+1)
		}
		inst, err := db.Create(body)
		if err != nil {
			je(w, 400, err.Error())
			return
		}
		jw(w, map[string]string{"ok": "true", "id": inst.ID})
	default:
		w.WriteHeader(405)
	}
}

func apiInstance(w http.ResponseWriter, r *http.Request) {
	tail := strings.TrimPrefix(r.URL.Path, "/api/instances/")
	parts := strings.SplitN(tail, "/", 2)
	id := parts[0]
	action := ""
	if len(parts) == 2 {
		action = parts[1]
	}
	if id == "" {
		je(w, 400, "missing id")
		return
	}

	switch action {
	case "start":
		inst, ok := db.Get(id)
		if !ok {
			je(w, 404, "not found")
			return
		}
		if err := mgr.Start(inst); err != nil {
			je(w, 500, err.Error())
			return
		}
		jw(w, map[string]bool{"ok": true})
	case "stop":
		mgr.Stop(id)
		jw(w, map[string]bool{"ok": true})
	case "restart":
		inst, ok := db.Get(id)
		if !ok {
			je(w, 404, "not found")
			return
		}
		if err := mgr.Restart(inst); err != nil {
			je(w, 500, err.Error())
			return
		}
		jw(w, map[string]bool{"ok": true})
	case "logs":
		b, _ := os.ReadFile(mgr.LogPath(id))
		lines := strings.Split(string(b), "\n")
		if len(lines) > 200 {
			lines = lines[len(lines)-200:]
		}
		jw(w, map[string]string{"logs": strings.Join(lines, "\n")})
	case "clear_logs":
		mgr.ClearLog(id)
		jw(w, map[string]bool{"ok": true})
	case "test":
		inst, ok := db.Get(id)
		if !ok {
			je(w, 404, "not found")
			return
		}
		jw(w, dnstest.Run(inst.Resolver, inst.Domain))
	case "":
		switch r.Method {
		case "PUT":
			var body store.Instance
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				je(w, 400, "bad json")
				return
			}
			was := mgr.Status(id) == runner.Running
			if was {
				mgr.Stop(id)
			}
			if err := db.Update(id, body); err != nil {
				je(w, 500, err.Error())
				return
			}
			if was {
				if inst, ok := db.Get(id); ok {
					mgr.Start(inst)
				}
			}
			jw(w, map[string]bool{"ok": true})
		case "DELETE":
			mgr.Stop(id)
			db.Delete(id)
			jw(w, map[string]bool{"ok": true})
		default:
			w.WriteHeader(405)
		}
	default:
		je(w, 404, "unknown: "+action)
	}
}

func apiTestRaw(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	var body struct {
		Resolver string `json:"resolver"`
		Domain   string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Domain == "" {
		je(w, 400, "resolver and domain required")
		return
	}
	jw(w, dnstest.Run(body.Resolver, body.Domain))
}
