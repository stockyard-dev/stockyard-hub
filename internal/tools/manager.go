package tools

import (
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

// Manager handles discovering, starting, stopping, and installing tools.
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

// Discover finds all installed stockyard-* binaries.
func (m *Manager) Discover() []Status {
	var results []Status
	for _, t := range Catalog() {
		s := Status{
			Slug:     t.Slug,
			Name:     t.Name,
			Tagline:  t.Tagline,
			Port:     t.Port,
			Category: t.Category,
		}
		bin := filepath.Join(m.binDir, "stockyard-"+t.Slug)
		if _, err := os.Stat(bin); err == nil {
			s.Installed = true
			s.Health = "stopped"
			// Check if running
			if pid := m.findProcess(t.Slug); pid > 0 {
				s.Running = true
				s.PID = pid
				s.Health = m.checkHealth(t.Port)
			}
		} else {
			// Also check current directory
			if _, err := os.Stat("stockyard-" + t.Slug); err == nil {
				s.Installed = true
				s.Health = "stopped"
			} else {
				s.Health = "not_installed"
			}
		}
		results = append(results, s)
	}
	return results
}

// Start launches a tool as a background process.
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

	// Wait briefly and check health
	go func() { cmd.Wait() }()
	return nil
}

// Stop kills a running tool process.
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

// Install downloads a tool binary from GitHub.
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
