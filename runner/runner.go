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
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
)

type procEntry struct {
	cmd   *exec.Cmd
	timer *time.Timer
}

type Manager struct {
	mu      sync.Mutex
	procs   map[string]*procEntry
	logsDir string
	binPath string
}

func New(logsDir string) *Manager {
	return &Manager{
		procs:   make(map[string]*procEntry),
		logsDir: logsDir,
		binPath: findBin(),
	}
}

func findBin() string {
	candidates := []string{
		"/usr/local/bin/slipstream",
		"/usr/bin/slipstream",
		os.ExpandEnv("$HOME/slipstream"),
		"./slipstream",
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	if path, err := exec.LookPath("slipstream"); err == nil {
		return path
	}
	return ""
}

func (m *Manager) BinPath() string { return m.binPath }

func (m *Manager) Status(id string) Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.procs[id]
	if !ok {
		return StatusStopped
	}
	if entry.cmd.ProcessState != nil {
		return StatusStopped
	}
	if entry.cmd.Process == nil {
		return StatusStopped
	}
	// Check if process is still alive
	err := entry.cmd.Process.Signal(syscall.Signal(0))
	if err != nil {
		return StatusStopped
	}
	return StatusRunning
}

func (m *Manager) Start(inst store.Instance) error {
	if m.binPath == "" {
		return fmt.Errorf("slipstream binary not found")
	}

	m.mu.Lock()
	// Kill existing if any
	if entry, ok := m.procs[inst.ID]; ok {
		m.killEntry(entry)
		delete(m.procs, inst.ID)
	}
	m.mu.Unlock()

	args := []string{"-r", inst.Resolver, "-d", inst.Domain, "-l", fmt.Sprintf("%d", inst.SocksPort)}
	if strings.TrimSpace(inst.ExtraArgs) != "" {
		extra := strings.Fields(inst.ExtraArgs)
		args = append(args, extra...)
	}

	logFile := filepath.Join(m.logsDir, inst.ID+".log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}

	cmd := exec.Command(m.binPath, args...)
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	fmt.Fprintf(f, "\n[panel] started at %s\n[panel] cmd: %s %s\n",
		time.Now().Format("2006-01-02 15:04:05"), m.binPath, strings.Join(args, " "))

	if err := cmd.Start(); err != nil {
		f.Close()
		return fmt.Errorf("start process: %w", err)
	}

	// Wait in background so ProcessState gets set
	go func() {
		cmd.Wait()
		f.Close()
		log.Printf("[runner] instance %s exited", inst.ID)
	}()

	entry := &procEntry{cmd: cmd}
	m.scheduleRestart(entry, inst)

	m.mu.Lock()
	m.procs[inst.ID] = entry
	m.mu.Unlock()

	log.Printf("[runner] started instance %s (pid=%d) socks=:%d", inst.ID, cmd.Process.Pid, inst.SocksPort)
	return nil
}

func (m *Manager) Stop(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.procs[id]; ok {
		m.killEntry(entry)
		delete(m.procs, id)
	}
	log.Printf("[runner] stopped instance %s", id)
}

func (m *Manager) Restart(inst store.Instance) error {
	m.Stop(inst.ID)
	time.Sleep(300 * time.Millisecond)
	return m.Start(inst)
}

func (m *Manager) killEntry(entry *procEntry) {
	if entry.timer != nil {
		entry.timer.Stop()
	}
	if entry.cmd.Process != nil {
		syscall.Kill(-entry.cmd.Process.Pid, syscall.SIGTERM)
		time.AfterFunc(2*time.Second, func() {
			if entry.cmd.ProcessState == nil {
				entry.cmd.Process.Kill()
			}
		})
	}
}

func (m *Manager) scheduleRestart(entry *procEntry, inst store.Instance) {
	if inst.RestartMinutes <= 0 || !inst.AutoRestart {
		return
	}
	entry.timer = time.AfterFunc(time.Duration(inst.RestartMinutes)*time.Minute, func() {
		log.Printf("[runner] auto-restart instance %s", inst.ID)
		m.Restart(inst)
	})
}

func (m *Manager) LogPath(id string) string {
	return filepath.Join(m.logsDir, id+".log")
}

func (m *Manager) ClearLog(id string) error {
	return os.WriteFile(filepath.Join(m.logsDir, id+".log"), nil, 0644)
}
