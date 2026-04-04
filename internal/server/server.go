package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stockyard-dev/stockyard-hub/internal/store"
	"github.com/stockyard-dev/stockyard-hub/internal/tools"
)

type Server struct {
	mgr    *tools.Manager
	db     *store.DB
	mux    *http.ServeMux
	limits Limits
}

func New(mgr *tools.Manager, db *store.DB) *Server {
	limits := DefaultLimits()
	s := &Server{mgr: mgr, db: db, mux: http.NewServeMux(), limits: limits}

	// Tools
	s.mux.HandleFunc("GET /api/tools", s.listTools)
	s.mux.HandleFunc("POST /api/tools/{slug}/start", s.startTool)
	s.mux.HandleFunc("POST /api/tools/{slug}/stop", s.stopTool)
	s.mux.HandleFunc("POST /api/tools/{slug}/install", s.installTool)
	s.mux.HandleFunc("POST /api/tools/{slug}/restart", s.restartTool)
	s.mux.HandleFunc("GET /api/tools/{slug}", s.getTool)
	s.mux.HandleFunc("GET /api/tools/{slug}/logs", s.toolLogs)
	s.mux.HandleFunc("GET /api/tools/{slug}/health-history", s.toolHealthHistory)

	// Global
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/activity", s.activity)
	s.mux.HandleFunc("POST /api/activity", s.seedActivity)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/health-history", s.globalHealthHistory)

	// License
	s.mux.HandleFunc("GET /api/license", s.getLicense)
	s.mux.HandleFunc("POST /api/license", s.setLicense)

	// Tier
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{"tier": s.limits.Tier, "upgrade_url": "https://stockyard.dev/hub/"})
	})

	// Dashboard
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)

	// Tool proxy (serves tool dashboards through Hub)
	s.mux.HandleFunc("/tool/{slug}/{path...}", s.proxyTool)

	s.mux.HandleFunc("GET /", s.root)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func wj(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
func we(w http.ResponseWriter, code int, msg string) {
	wj(w, code, map[string]string{"error": msg})
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", 302)
}

// ── Tools ──

func (s *Server) listTools(w http.ResponseWriter, r *http.Request) {
	statuses := s.mgr.Discover()
	q := strings.ToLower(r.URL.Query().Get("q"))
	cat := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
	var filtered []tools.Status
	for _, st := range statuses {
		if q != "" && !strings.Contains(strings.ToLower(st.Name), q) && !strings.Contains(strings.ToLower(st.Tagline), q) && !strings.Contains(strings.ToLower(st.Slug), q) {
			continue
		}
		if cat != "" && st.Category != cat {
			continue
		}
		if status == "installed" && !st.Installed {
			continue
		}
		if status == "running" && !st.Running {
			continue
		}
		if status == "stopped" && (!st.Installed || st.Running) {
			continue
		}
		filtered = append(filtered, st)
	}
	if filtered == nil {
		filtered = []tools.Status{}
	}
	wj(w, 200, map[string]any{"tools": filtered})
}

func (s *Server) getTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	for _, st := range s.mgr.Discover() {
		if st.Slug == slug {
			wj(w, 200, st)
			return
		}
	}
	we(w, 404, "tool not found")
}

func (s *Server) startTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := s.mgr.Start(slug); err != nil {
		we(w, 400, err.Error())
		return
	}
	s.db.LogActivity(slug, "started", "Started via Hub")
	tools.FireWebhook("tool.started", slug, slug, 0)
	wj(w, 200, map[string]string{"status": "started", "slug": slug})
}

func (s *Server) stopTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := s.mgr.Stop(slug); err != nil {
		we(w, 400, err.Error())
		return
	}
	s.db.LogActivity(slug, "stopped", "Stopped via Hub")
	tools.FireWebhook("tool.stopped", slug, slug, 0)
	wj(w, 200, map[string]string{"status": "stopped", "slug": slug})
}

func (s *Server) restartTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	s.mgr.Stop(slug)
	if err := s.mgr.Start(slug); err != nil {
		we(w, 400, err.Error())
		return
	}
	s.db.LogActivity(slug, "restarted", "Restarted via Hub")
	wj(w, 200, map[string]string{"status": "restarted", "slug": slug})
}

func (s *Server) installTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := s.mgr.Install(slug); err != nil {
		we(w, 500, err.Error())
		return
	}
	s.db.LogActivity(slug, "installed", "Installed via Hub")
	tools.FireWebhook("tool.installed", slug, slug, 0)
	wj(w, 200, map[string]string{"status": "installed", "slug": slug})
}

func (s *Server) toolLogs(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	lines, _ := strconv.Atoi(r.URL.Query().Get("lines"))
	if lines <= 0 {
		lines = 100
	}
	content := readToolLog(slug, lines)
	wj(w, 200, map[string]string{"slug": slug, "logs": content})
}

func (s *Server) toolHealthHistory(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	records := s.db.HealthHistory(slug, limit)
	if records == nil {
		records = []store.HealthRecord{}
	}
	wj(w, 200, map[string]any{"slug": slug, "history": records})
}

// ── Global ──

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	statuses := s.mgr.Discover()
	installed, running, healthy := 0, 0, 0
	byCat := map[string]int{}
	for _, st := range statuses {
		if st.Installed {
			installed++
		}
		if st.Running {
			running++
		}
		if st.Health == "healthy" {
			healthy++
		}
		if st.Installed {
			byCat[st.Category]++
		}
	}
	wj(w, 200, map[string]any{
		"total": len(statuses), "installed": installed,
		"running": running, "healthy": healthy, "by_category": byCat,
	})
}

func (s *Server) activity(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	records := s.db.RecentActivity(limit)
	if records == nil {
		records = []store.ActivityRecord{}
	}
	wj(w, 200, map[string]any{"activity": records})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	statuses := s.mgr.Discover()
	installed, running := 0, 0
	for _, st := range statuses {
		if st.Installed {
			installed++
		}
		if st.Running {
			running++
		}
	}
	wj(w, 200, map[string]any{"status": "ok", "service": "hub", "tools_installed": installed, "tools_running": running})
}

func (s *Server) globalHealthHistory(w http.ResponseWriter, r *http.Request) {
	tool := r.URL.Query().Get("tool")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	records := s.db.HealthHistory(tool, limit)
	if records == nil {
		records = []store.HealthRecord{}
	}
	wj(w, 200, map[string]any{"records": records})
}

// ── License ──

func (s *Server) getLicense(w http.ResponseWriter, r *http.Request) {
	cfg := readCLIConfig()
	key, _ := cfg["license_key"].(string)
	masked := ""
	if len(key) > 6 {
		masked = key[:6] + strings.Repeat("*", len(key)-6)
	}
	wj(w, 200, map[string]any{"license_key": masked, "tier": s.limits.Tier})
}

func (s *Server) setLicense(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
		we(w, 400, "key required")
		return
	}
	cfg := readCLIConfig()
	cfg["license_key"] = body.Key
	if err := writeCLIConfig(cfg); err != nil {
		we(w, 500, err.Error())
		return
	}
	s.db.LogActivity("hub", "license_set", "License key updated")
	wj(w, 200, map[string]string{"status": "saved"})
}

func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

// Seed activity (for demo)
func (s *Server) seedActivity(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Tool   string `json:"tool"`
		Action string `json:"action"`
		Detail string `json:"detail"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		we(w, 400, "invalid json")
		return
	}
	s.db.LogActivity(body.Tool, body.Action, body.Detail)
	wj(w, 200, map[string]string{"status": "ok"})
}

// ── Tool Proxy (for demo) ──

func (s *Server) proxyTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	t := tools.FindTool(slug)
	if t == nil {
		we(w, 404, "tool not found")
		return
	}

	// Build target URL
	rest := r.PathValue("path")
	if rest == "" {
		rest = "/"
	}
	target := fmt.Sprintf("http://localhost:%d/%s", t.Port, rest)
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	client := &http.Client{Timeout: 10 * time.Second}
	proxyReq, err := http.NewRequest(r.Method, target, r.Body)
	if err != nil {
		we(w, 502, err.Error())
		return
	}
	proxyReq.Header = r.Header.Clone()

	resp, err := client.Do(proxyReq)
	if err != nil {
		we(w, 502, "tool not reachable: "+err.Error())
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
