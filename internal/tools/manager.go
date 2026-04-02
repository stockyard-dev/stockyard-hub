package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ServiceInfo matches the discovery package format.
type ServiceInfo struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Port      int    `json:"port"`
	PID       int    `json:"pid"`
	Health    string `json:"health_url"`
	Dashboard string `json:"dashboard_url"`
	StartedAt string `json:"started_at"`
	Version   string `json:"version,omitempty"`
}

type Manager struct {
	binDir  string
	dataDir string
	procs   map[string]*os.Process
}

func NewManager(binDir, dataDir string) *Manager {
	if binDir == "" {
		binDir = "/usr/local/bin"
	}
	return &Manager{binDir: binDir, dataDir: dataDir, procs: make(map[string]*os.Process)}
}

// discoveryDir returns ~/.stockyard/discovery/
func discoveryDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "/tmp"
	}
	dir := filepath.Join(home, ".stockyard", "discovery")
	os.MkdirAll(dir, 0755)
	return dir
}

// discoverRegistered finds tools that registered via the discovery package.
func discoverRegistered() map[string]ServiceInfo {
	result := make(map[string]ServiceInfo)
	dir := discoveryDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var info ServiceInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}
		// Verify process is still running
		p, err := os.FindProcess(info.PID)
		if err != nil {
			os.Remove(filepath.Join(dir, entry.Name()))
			continue
		}
		if err := p.Signal(syscall.Signal(0)); err != nil {
			os.Remove(filepath.Join(dir, entry.Name()))
			continue
		}
		result[info.Slug] = info
	}
	return result
}

func (m *Manager) Discover() []Status {
	// First, check file-based discovery
	registered := discoverRegistered()

	var results []Status
	for _, t := range Catalog() {
		s := Status{
			Slug:     t.Slug,
			Name:     t.Name,
			Tagline:  t.Tagline,
			Port:     t.Port,
			Category: t.Category,
		}

		// Check discovery registry first
		if info, ok := registered[t.Slug]; ok {
			s.Installed = true
			s.Running = true
			s.PID = info.PID
			s.Port = info.Port // Use actual port from discovery
			s.Health = m.checkHealth(info.Port)
			results = append(results, s)
			continue
		}

		// Fall back to binary detection
		bin := filepath.Join(m.binDir, "stockyard-"+t.Slug)
		if _, err := os.Stat(bin); err == nil {
			s.Installed = true
			s.Health = "stopped"
			if pid := m.findProcess(t.Slug); pid > 0 {
				s.Running = true
				s.PID = pid
				s.Health = m.checkHealth(t.Port)
			}
		} else if _, err := os.Stat("stockyard-" + t.Slug); err == nil {
			s.Installed = true
			s.Health = "stopped"
		} else {
			s.Health = "not_installed"
		}
		results = append(results, s)
	}
	return results
}

func (m *Manager) Start(slug string) error {
	t := FindTool(slug)
	if t == nil {
		return fmt.Errorf("unknown tool: %s", slug)
	}
	bin := filepath.Join(m.binDir, "stockyard-"+slug)
	if _, err := os.Stat(bin); err != nil {
		bin = "stockyard-" + slug
		if _, err := os.Stat(bin); err != nil {
			return fmt.Errorf("stockyard-%s not installed", slug)
		}
	}
	dataDir := filepath.Join(m.dataDir, slug+"-data")
	os.MkdirAll(dataDir, 0755)
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", t.Port),
		fmt.Sprintf("DATA_DIR=%s", dataDir),
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	m.procs[slug] = cmd.Process
	go func() { cmd.Wait() }()
	return nil
}

func (m *Manager) Stop(slug string) error {
	if p, ok := m.procs[slug]; ok {
		p.Signal(syscall.SIGTERM)
		delete(m.procs, slug)
		return nil
	}
	if pid := m.findProcess(slug); pid > 0 {
		p, err := os.FindProcess(pid)
		if err == nil {
			p.Signal(syscall.SIGTERM)
		}
		return nil
	}
	return fmt.Errorf("stockyard-%s not running", slug)
}

func (m *Manager) Install(slug string) error {
	t := FindTool(slug)
	if t == nil {
		return fmt.Errorf("unknown tool: %s", slug)
	}
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("curl -fsSL https://stockyard.dev/%s/install.sh | INSTALL_DIR=%s sh", slug, m.binDir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Manager) findProcess(slug string) int {
	out, err := exec.Command("pgrep", "-f", "stockyard-"+slug).Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 {
		pid, _ := strconv.Atoi(lines[0])
		return pid
	}
	return 0
}

func (m *Manager) checkHealth(port int) string {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/api/health", port))
	if err != nil {
		return "unhealthy"
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return "healthy"
	}
	return "unhealthy"
}

func FindTool(slug string) *Tool {
	for _, t := range Catalog() {
		if t.Slug == slug {
			t2 := t
			return &t2
		}
	}
	return nil
}
