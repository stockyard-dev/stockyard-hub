package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/stockyard-dev/stockyard-hub/internal/tools"
)

type Server struct {
	mgr *tools.Manager
	mux *http.ServeMux
}

func New(mgr *tools.Manager) *Server {
	s := &Server{mgr: mgr, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /api/tools", s.listTools)
	s.mux.HandleFunc("POST /api/tools/{slug}/start", s.startTool)
	s.mux.HandleFunc("POST /api/tools/{slug}/stop", s.stopTool)
	s.mux.HandleFunc("POST /api/tools/{slug}/install", s.installTool)
	s.mux.HandleFunc("GET /api/tools/{slug}", s.getTool)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
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

func (s *Server) listTools(w http.ResponseWriter, r *http.Request) {
	statuses := s.mgr.Discover()
	// Filter by query params
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
	statuses := s.mgr.Discover()
	for _, st := range statuses {
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
	wj(w, 200, map[string]string{"status": "started", "slug": slug})
}

func (s *Server) stopTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := s.mgr.Stop(slug); err != nil {
		we(w, 400, err.Error())
		return
	}
	wj(w, 200, map[string]string{"status": "stopped", "slug": slug})
}

func (s *Server) installTool(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := s.mgr.Install(slug); err != nil {
		we(w, 500, err.Error())
		return
	}
	wj(w, 200, map[string]string{"status": "installed", "slug": slug})
}

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
		"total":        len(statuses),
		"installed":    installed,
		"running":      running,
		"healthy":      healthy,
		"by_category":  byCat,
	})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	statuses := s.mgr.Discover()
	installed := 0
	for _, st := range statuses {
		if st.Installed {
			installed++
		}
	}
	wj(w, 200, map[string]any{"status": "ok", "service": "hub", "tools_installed": installed})
}

func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }
