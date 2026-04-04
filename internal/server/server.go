package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

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
	s.mux.HandleFunc("GET /api/tools/{slug}", s.getTool)

	// Stats & Health
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/health/history", s.healthHistory)

	// Activity
	s.mux.HandleFunc("GET /api/activity", s.activity)

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
	s.db.LogActivity(slug, "started", "Tool started")
	tools.FireWebhook("tool.started", slug, slug, 0)
	wj(w, 200, map[string]string{"status": "started", "slug": slug})
}

func (s *Server) stopTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := s.mgr.Stop(slug); err != nil {
		we(w, 400, err.Error())
		return
	}
	s.db.LogActivity(slug, "stopped", "Tool stopped")
	tools.FireWebhook("tool.stopped", slug, slug, 0)
	wj(w, 200, map[string]string{"status": "stopped", "slug": slug})
}

func (s *Server) installTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := s.mgr.Install(slug); err != nil {
		we(w, 500, err.Error())
		return
	}
	s.db.LogActivity(slug, "installed", "Tool installed")
	tools.FireWebhook("tool.installed", slug, slug, 0)
	wj(w, 200, map[string]string{"status": "installed", "slug": slug})
}

// ── Stats & Health ──

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
		"total":       len(statuses),
		"installed":   installed,
		"running":     running,
		"healthy":     healthy,
		"by_category": byCat,
	})
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

func (s *Server) healthHistory(w http.ResponseWriter, r *http.Request) {
	tool := r.URL.Query().Get("tool")
	records := s.db.HealthHistory(tool, 100)
	if records == nil {
		records = []store.HealthRecord{}
	}
	wj(w, 200, map[string]any{"records": records})
}

// ── Activity ──

func (s *Server) activity(w http.ResponseWriter, r *http.Request) {
	records := s.db.RecentActivity(50)
	if records == nil {
		records = []store.ActivityRecord{}
	}
	wj(w, 200, map[string]any{"activity": records})
}

// ── License ──

func (s *Server) getLicense(w http.ResponseWriter, r *http.Request) {
	key := s.db.GetConfig("license_key")
	masked := ""
	if key != "" && len(key) > 6 {
		masked = key[:6] + strings.Repeat("*", len(key)-6)
	} else if key != "" {
		masked = "***"
	}
	wj(w, 200, map[string]any{"license_key": masked, "tier": s.limits.Tier})
}

func (s *Server) setLicense(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		we(w, 400, "invalid json")
		return
	}
	s.db.SetConfig("license_key", body.Key)
	s.db.LogActivity("hub", "license_updated", "License key updated")
	wj(w, 200, map[string]string{"status": "saved"})
}

func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }
