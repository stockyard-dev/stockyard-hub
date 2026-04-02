package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/stockyard-dev/stockyard-hub/internal/store"
	"github.com/stockyard-dev/stockyard-hub/internal/tools"
)

type Server struct {
	db         *store.DB
	mux        *http.ServeMux
	dataDir    string
	binDir     string
	licenseKey string
}

func New(db *store.DB, dataDir, licenseKey string) *Server {
	binDir := filepath.Join(dataDir, "bin")
	os.MkdirAll(binDir, 0755)

	s := &Server{
		db:         db,
		mux:        http.NewServeMux(),
		dataDir:    dataDir,
		binDir:     binDir,
		licenseKey: licenseKey,
	}
	s.mux.HandleFunc("GET /api/tools", s.handleListTools)
	s.mux.HandleFunc("GET /api/tools/{slug}", s.handleGetTool)
	s.mux.HandleFunc("POST /api/tools/{slug}/install", s.handleInstall)
	s.mux.HandleFunc("POST /api/tools/{slug}/uninstall", s.handleUninstall)
	s.mux.HandleFunc("POST /api/tools/{slug}/start", s.handleStart)
	s.mux.HandleFunc("POST /api/tools/{slug}/stop", s.handleStop)
	s.mux.HandleFunc("POST /api/tools/{slug}/restart", s.handleRestart)
	s.mux.HandleFunc("POST /api/config/license", s.handleSetLicense)
	s.mux.HandleFunc("GET /api/config", s.handleGetConfig)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /ui", s.handleDashboard)
	s.mux.HandleFunc("GET /ui/", s.handleDashboard)
	s.mux.HandleFunc("GET /", s.handleRoot)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		http.Redirect(w, r, "/ui", http.StatusFound)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	installed := s.db.ListInstalled()
	running := 0
	for _, t := range installed {
		if t.PID > 0 && isProcessRunning(t.PID) {
			running++
		}
	}
	writeJSON(w, map[string]any{
		"status":    "ok",
		"installed": len(installed),
		"running":   running,
		"catalog":   len(tools.Catalog()),
	})
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"license_key_set": s.licenseKey != "",
		"data_dir":        s.dataDir,
		"bin_dir":         s.binDir,
	})
}

func (s *Server) handleSetLicense(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "invalid JSON")
		return
	}
	s.licenseKey = req.Key
	s.db.SetConfig("license_key", req.Key)
	writeJSON(w, map[string]string{"status": "ok"})
}

// handleListTools returns all 150 tools with install/run status.
func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	catalog := tools.Catalog()
	category := r.URL.Query().Get("category")
	search := strings.ToLower(r.URL.Query().Get("q"))

	var result []tools.Status
	for _, t := range catalog {
		if category != "" && t.Category != category {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(t.Name+t.Tagline+t.Slug), search) {
			continue
		}

		st := tools.Status{
			Slug:     t.Slug,
			Name:     t.Name,
			Tagline:  t.Tagline,
			Port:     t.Port,
			Category: t.Category,
			Health:   "not_installed",
		}

		installed := s.db.GetInstalled(t.Slug)
		if installed != nil {
			st.Installed = true
			st.Health = "stopped"
			st.PID = installed.PID
			if installed.PID > 0 && isProcessRunning(installed.PID) {
				st.Running = true
				if checkHealth(t.Port) {
					st.Health = "healthy"
				} else {
					st.Health = "unhealthy"
				}
			} else if installed.PID > 0 {
				// PID recorded but process not running — clean up
				s.db.ClearPID(t.Slug)
				st.PID = 0
			}
		}
		result = append(result, st)
	}
	writeJSON(w, map[string]any{"tools": result, "count": len(result)})
}

func (s *Server) handleGetTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var tool *tools.Tool
	for _, t := range tools.Catalog() {
		if t.Slug == slug {
			tool = &t
			break
		}
	}
	if tool == nil {
		writeErr(w, 404, "tool not found")
		return
	}

	st := tools.Status{
		Slug:     tool.Slug,
		Name:     tool.Name,
		Tagline:  tool.Tagline,
		Port:     tool.Port,
		Category: tool.Category,
		Health:   "not_installed",
	}

	installed := s.db.GetInstalled(slug)
	if installed != nil {
		st.Installed = true
		st.Health = "stopped"
		st.PID = installed.PID
		if installed.PID > 0 && isProcessRunning(installed.PID) {
			st.Running = true
			if checkHealth(tool.Port) {
				st.Health = "healthy"
			} else {
				st.Health = "unhealthy"
			}
		}
	}
	writeJSON(w, st)
}

func (s *Server) handleInstall(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var tool *tools.Tool
	for _, t := range tools.Catalog() {
		if t.Slug == slug {
			tool = &t
			break
		}
	}
	if tool == nil {
		writeErr(w, 404, "tool not found")
		return
	}

	// Download binary
	binaryName := "stockyard-" + slug
	arch := runtime.GOARCH
	goos := runtime.GOOS
	url := fmt.Sprintf("https://github.com/stockyard-dev/stockyard-%s/releases/latest/download/%s-%s-%s",
		slug, binaryName, goos, arch)

	log.Printf("hub: installing %s from %s", slug, url)

	resp, err := http.Get(url)
	if err != nil {
		// Fallback: try the install script pattern
		url = fmt.Sprintf("https://stockyard.dev/%s/install.sh", slug)
		writeErr(w, 500, fmt.Sprintf("download failed: %v — try manually: curl -fsSL %s | sh", err, url))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Binary not available as GitHub release — mark as installed locally
		// (user may have built from source or installed via curl)
		binaryPath := slug
		// Check if binary exists on PATH
		if path, err := exec.LookPath(slug); err == nil {
			binaryPath = path
		} else if path, err := exec.LookPath("stockyard-" + slug); err == nil {
			binaryPath = path
		}

		toolDataDir := filepath.Join(s.dataDir, "tools", slug)
		os.MkdirAll(toolDataDir, 0755)
		s.db.MarkInstalled(slug, binaryPath, toolDataDir)
		writeJSON(w, map[string]string{
			"status": "registered",
			"note":   fmt.Sprintf("Binary not found in releases. Registered as '%s'. Install manually: curl -fsSL https://stockyard.dev/%s/install.sh | sh", binaryPath, slug),
		})
		return
	}

	binPath := filepath.Join(s.binDir, binaryName)
	f, err := os.Create(binPath)
	if err != nil {
		writeErr(w, 500, fmt.Sprintf("create file: %v", err))
		return
	}
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		writeErr(w, 500, fmt.Sprintf("download: %v", err))
		return
	}
	os.Chmod(binPath, 0755)

	toolDataDir := filepath.Join(s.dataDir, "tools", slug)
	os.MkdirAll(toolDataDir, 0755)
	s.db.MarkInstalled(slug, binPath, toolDataDir)

	log.Printf("hub: installed %s at %s", slug, binPath)
	writeJSON(w, map[string]string{"status": "installed", "binary": binPath, "data_dir": toolDataDir})
}

func (s *Server) handleUninstall(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	installed := s.db.GetInstalled(slug)
	if installed == nil {
		writeErr(w, 404, "tool not installed")
		return
	}

	// Stop if running
	if installed.PID > 0 && isProcessRunning(installed.PID) {
		syscall.Kill(installed.PID, syscall.SIGTERM)
		time.Sleep(500 * time.Millisecond)
	}

	// Remove binary if in our bin dir
	if strings.HasPrefix(installed.BinaryPath, s.binDir) {
		os.Remove(installed.BinaryPath)
	}

	s.db.MarkUninstalled(slug)
	log.Printf("hub: uninstalled %s", slug)
	writeJSON(w, map[string]string{"status": "uninstalled"})
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	installed := s.db.GetInstalled(slug)
	if installed == nil {
		writeErr(w, 404, "tool not installed — install it first")
		return
	}

	if installed.PID > 0 && isProcessRunning(installed.PID) {
		writeJSON(w, map[string]string{"status": "already_running", "pid": strconv.Itoa(installed.PID)})
		return
	}

	// Find the binary
	binary := installed.BinaryPath
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		// Try PATH
		if path, err := exec.LookPath(slug); err == nil {
			binary = path
		} else if path, err := exec.LookPath("stockyard-" + slug); err == nil {
			binary = path
		} else {
			writeErr(w, 500, fmt.Sprintf("binary not found at %s or on PATH", installed.BinaryPath))
			return
		}
	}

	// Build environment
	env := os.Environ()
	env = append(env, fmt.Sprintf("DATA_DIR=%s", installed.DataDir))
	env = append(env, fmt.Sprintf("PORT=%d", portForSlug(slug)))

	// Set license key
	licKey := s.licenseKey
	if licKey != "" {
		envVar := strings.ToUpper(slug) + "_LICENSE_KEY"
		env = append(env, fmt.Sprintf("%s=%s", envVar, licKey))
	}

	// Start the process
	cmd := exec.Command(binary)
	cmd.Env = env
	cmd.Dir = installed.DataDir

	logFile := filepath.Join(installed.DataDir, "output.log")
	lf, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		cmd.Stdout = lf
		cmd.Stderr = lf
	}

	if err := cmd.Start(); err != nil {
		writeErr(w, 500, fmt.Sprintf("start failed: %v", err))
		return
	}

	pid := cmd.Process.Pid
	s.db.SetPID(slug, pid)

	// Detach — don't wait for the process
	go func() {
		cmd.Wait()
		s.db.ClearPID(slug)
		if lf != nil {
			lf.Close()
		}
	}()

	log.Printf("hub: started %s (PID %d, port %d)", slug, pid, portForSlug(slug))
	writeJSON(w, map[string]any{"status": "started", "pid": pid, "port": portForSlug(slug)})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	installed := s.db.GetInstalled(slug)
	if installed == nil {
		writeErr(w, 404, "tool not installed")
		return
	}
	if installed.PID == 0 || !isProcessRunning(installed.PID) {
		s.db.ClearPID(slug)
		writeJSON(w, map[string]string{"status": "not_running"})
		return
	}

	syscall.Kill(installed.PID, syscall.SIGTERM)
	// Wait briefly for graceful shutdown
	for i := 0; i < 10; i++ {
		time.Sleep(200 * time.Millisecond)
		if !isProcessRunning(installed.PID) {
			break
		}
	}
	// Force kill if still running
	if isProcessRunning(installed.PID) {
		syscall.Kill(installed.PID, syscall.SIGKILL)
	}
	s.db.ClearPID(slug)

	log.Printf("hub: stopped %s (was PID %d)", slug, installed.PID)
	writeJSON(w, map[string]string{"status": "stopped"})
}

func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	installed := s.db.GetInstalled(slug)
	if installed == nil {
		writeErr(w, 404, "tool not installed")
		return
	}
	// Stop
	if installed.PID > 0 && isProcessRunning(installed.PID) {
		syscall.Kill(installed.PID, syscall.SIGTERM)
		time.Sleep(time.Second)
		if isProcessRunning(installed.PID) {
			syscall.Kill(installed.PID, syscall.SIGKILL)
		}
		s.db.ClearPID(slug)
	}
	// Start via the handler
	s.handleStart(w, r)
}

func portForSlug(slug string) int {
	for _, t := range tools.Catalog() {
		if t.Slug == slug {
			return t.Port
		}
	}
	return 0
}

func isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func checkHealth(port int) bool {
	if port == 0 {
		return false
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
