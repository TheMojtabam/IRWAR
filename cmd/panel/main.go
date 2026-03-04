package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yourusername/slipstream-panel/internal/dnstest"
	"github.com/yourusername/slipstream-panel/internal/runner"
	"github.com/yourusername/slipstream-panel/internal/store"
)

//go:embed static/index.html
var staticFiles embed.FS

var (
	st  *store.Store
	mgr *runner.Manager
)

func main() {
	baseDir := "/opt/slipstream-panel"
	if dir := os.Getenv("PANEL_DIR"); dir != "" {
		baseDir = dir
	}
	port := "9090"
	if p := os.Getenv("PANEL_PORT"); p != "" {
		port = p
	}

	os.MkdirAll(baseDir, 0755)
	logsDir := filepath.Join(baseDir, "logs")
	os.MkdirAll(logsDir, 0755)

	var err error
	st, err = store.New(filepath.Join(baseDir, "instances.json"))
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	mgr = runner.New(logsDir)
	log.Printf("[panel] binary: %s", mgr.BinPath())

	// Auto-start instances on boot
	for _, inst := range st.List() {
		if inst.AutoRestart {
			if err := mgr.Start(inst); err != nil {
				log.Printf("[panel] autostart %s failed: %v", inst.ID, err)
			}
		}
	}

	mux := http.NewServeMux()

	// ── Static files ──────────────────────────────────────
	sub, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/", http.FileServer(http.FS(sub)))

	// ── API ───────────────────────────────────────────────
	mux.HandleFunc("/api/status",              handleStatus)
	mux.HandleFunc("/api/instances",           handleInstances)
	mux.HandleFunc("/api/instances/",          handleInstance)
	mux.HandleFunc("/api/test",                handleTestRaw)

	log.Printf("[panel] listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// ── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

func decodeBody(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// ── /api/status ───────────────────────────────────────────────────────────────
func handleStatus(w http.ResponseWriter, r *http.Request) {
	bin := mgr.BinPath()
	if bin == "" {
		bin = "NOT FOUND"
	}
	writeJSON(w, map[string]string{"bin": bin})
}

// ── /api/instances ────────────────────────────────────────────────────────────
func handleInstances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list := st.List()
		type instView struct {
			store.Instance
			Status string `json:"status"`
		}
		out := make([]instView, len(list))
		for i, inst := range list {
			out[i] = instView{
				Instance: inst,
				Status:   string(mgr.Status(inst.ID)),
			}
		}
		writeJSON(w, out)

	case http.MethodPost:
		var body store.Instance
		if err := decodeBody(r, &body); err != nil {
			writeErr(w, 400, "invalid body")
			return
		}
		if body.Domain == "" {
			writeErr(w, 400, "domain required")
			return
		}
		if body.SocksPort == 0 {
			body.SocksPort = 1080
		}
		if body.Name == "" {
			body.Name = fmt.Sprintf("Instance %d", len(st.List())+1)
		}
		inst, err := st.Create(body)
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, map[string]string{"ok": "true", "id": inst.ID})

	default:
		w.WriteHeader(405)
	}
}

// ── /api/instances/{id}[/action] ─────────────────────────────────────────────
func handleInstance(w http.ResponseWriter, r *http.Request) {
	// Parse path: /api/instances/{id}  or  /api/instances/{id}/action
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/instances/"), "/")
	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	if id == "" {
		writeErr(w, 400, "missing id")
		return
	}

	switch action {
	case "start":
		inst, ok := st.Get(id)
		if !ok { writeErr(w, 404, "not found"); return }
		if err := mgr.Start(inst); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		writeJSON(w, map[string]bool{"ok": true})

	case "stop":
		mgr.Stop(id)
		writeJSON(w, map[string]bool{"ok": true})

	case "restart":
		inst, ok := st.Get(id)
		if !ok { writeErr(w, 404, "not found"); return }
		if err := mgr.Restart(inst); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		writeJSON(w, map[string]bool{"ok": true})

	case "logs":
		logPath := mgr.LogPath(id)
		lines := ""
		if f, err := os.Open(logPath); err == nil {
			defer f.Close()
			// Read last 200 lines
			b, _ := io.ReadAll(f)
			all := strings.Split(string(b), "\n")
			if len(all) > 200 {
				all = all[len(all)-200:]
			}
			lines = strings.Join(all, "\n")
		}
		writeJSON(w, map[string]string{"logs": lines})

	case "clear_logs":
		mgr.ClearLog(id)
		writeJSON(w, map[string]bool{"ok": true})

	case "test":
		inst, ok := st.Get(id)
		if !ok { writeErr(w, 404, "not found"); return }
		result := dnstest.Run(inst.Resolver, inst.Domain)
		writeJSON(w, result)

	case "": // /api/instances/{id}
		switch r.Method {
		case http.MethodPut:
			var body store.Instance
			if err := decodeBody(r, &body); err != nil {
				writeErr(w, 400, "invalid body")
				return
			}
			wasRunning := mgr.Status(id) == runner.StatusRunning
			if wasRunning {
				mgr.Stop(id)
			}
			if err := st.Update(id, body); err != nil {
				writeErr(w, 500, err.Error())
				return
			}
			if wasRunning {
				inst, _ := st.Get(id)
				mgr.Start(inst)
			}
			writeJSON(w, map[string]bool{"ok": true})

		case http.MethodDelete:
			mgr.Stop(id)
			st.Delete(id)
			writeJSON(w, map[string]bool{"ok": true})

		default:
			w.WriteHeader(405)
		}

	default:
		writeErr(w, 404, "unknown action: "+action)
	}
}

// ── /api/test (raw, no saved instance) ───────────────────────────────────────
func handleTestRaw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405); return
	}
	var body struct {
		Resolver string `json:"resolver"`
		Domain   string `json:"domain"`
	}
	if err := decodeBody(r, &body); err != nil || body.Domain == "" {
		writeErr(w, 400, "resolver and domain required")
		return
	}
	result := dnstest.Run(body.Resolver, body.Domain)
	writeJSON(w, result)
}

// port helper (unused but kept for reference)
var _ = strconv.Itoa
