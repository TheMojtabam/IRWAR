package runner

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/yourusername/slipstream-panel/internal/store"
)

type Status string

const (
	Running Status = "running"
	Stopped Status = "stopped"
)

type entry struct {
	cmd     *exec.Cmd
	logFile *os.File
	stop    chan struct{} // closes to cancel restart timer
}

type Manager struct {
	mu      sync.Mutex
	procs   map[string]*entry
	logsDir string
	bin     string
}

func New(logsDir string) *Manager {
	return &Manager{
		procs:   make(map[string]*entry),
		logsDir: logsDir,
		bin:     findBin(),
	}
}

func findBin() string {
	if v := os.Getenv("SLIPSTREAM_BIN"); v != "" {
		return v
	}
	names := []string{"slp", "slipstream", "slepstream"}
	dirs := []string{"/usr/local/bin", "/usr/bin", os.ExpandEnv("$HOME")}
	for _, d := range dirs {
		for _, n := range names {
			p := filepath.Join(d, n)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
	}
	for _, n := range names {
		if p, err := exec.LookPath(n); err == nil {
			return p
		}
	}
	return ""
}

func (m *Manager) Bin() string { return m.bin }

func (m *Manager) Status(id string) Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.procs[id]
	if !ok {
		return Stopped
	}
	// signal 0 = check if alive
	if err := e.cmd.Process.Signal(syscall.Signal(0)); err != nil {
		delete(m.procs, id)
		return Stopped
	}
	return Running
}

func (m *Manager) Start(inst store.Instance) error {
	if m.bin == "" {
		return fmt.Errorf("slipstream binary not found — set SLIPSTREAM_BIN env var")
	}

	// stop existing first
	m.stopLocked(inst.ID)

	logPath := filepath.Join(m.logsDir, inst.ID+".log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}

	args := []string{"-r", inst.Resolver, "-d", inst.Domain, "-l", fmt.Sprintf("%d", inst.SocksPort)}
	if s := strings.TrimSpace(inst.ExtraArgs); s != "" {
		args = append(args, strings.Fields(s)...)
	}

	cmd := exec.Command(m.bin, args...)
	cmd.Stdout = f
	cmd.Stderr = f
	// New process group so we can kill the whole tree
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	fmt.Fprintf(f, "\n--- started %s  cmd: %s %s\n",
		time.Now().Format("2006-01-02 15:04:05"), m.bin, strings.Join(args, " "))

	if err := cmd.Start(); err != nil {
		f.Close()
		return fmt.Errorf("exec: %w", err)
	}

	stopCh := make(chan struct{})
	e := &entry{cmd: cmd, logFile: f, stop: stopCh}

	m.mu.Lock()
	m.procs[inst.ID] = e
	m.mu.Unlock()

	// Reap process in background — no goroutine leak: exits when process exits
	go func() {
		cmd.Wait()
		f.Close()
		log.Printf("[runner] %s (pid %d) exited", inst.ID, cmd.Process.Pid)
		m.mu.Lock()
		if m.procs[inst.ID] == e {
			delete(m.procs, inst.ID)
		}
		m.mu.Unlock()
	}()

	// Auto-restart timer — single goroutine, exits on stop signal or process death
	if inst.RestartMinutes > 0 && inst.AutoRestart {
		go func() {
			select {
			case <-stopCh:
				return
			case <-time.After(time.Duration(inst.RestartMinutes) * time.Minute):
			}
			log.Printf("[runner] auto-restart %s", inst.ID)
			m.Start(inst) // re-uses same inst snapshot
		}()
	}

	log.Printf("[runner] started %s pid=%d socks=:%d", inst.ID, cmd.Process.Pid, inst.SocksPort)
	return nil
}

func (m *Manager) Stop(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopLocked(id)
}

// must be called with m.mu held
func (m *Manager) stopLocked(id string) {
	e, ok := m.procs[id]
	if !ok {
		return
	}
	delete(m.procs, id)

	// Cancel the restart timer goroutine
	close(e.stop)

	if e.cmd.Process == nil {
		return
	}
	// Kill the whole process group
	pgid, err := syscall.Getpgid(e.cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		e.cmd.Process.Kill()
	}
	// Give it 2s then force kill
	go func() {
		time.Sleep(2 * time.Second)
		if e.cmd.ProcessState == nil {
			if pgid, err := syscall.Getpgid(e.cmd.Process.Pid); err == nil {
				syscall.Kill(-pgid, syscall.SIGKILL)
			}
		}
	}()
}

func (m *Manager) Restart(inst store.Instance) error {
	m.Stop(inst.ID)
	time.Sleep(300 * time.Millisecond)
	return m.Start(inst)
}

func (m *Manager) LogPath(id string) string {
	return filepath.Join(m.logsDir, id+".log")
}

func (m *Manager) ClearLog(id string) {
	os.WriteFile(filepath.Join(m.logsDir, id+".log"), nil, 0644)
}
