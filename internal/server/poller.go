package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/stockyard-dev/stockyard-hub/internal/store"
	"github.com/stockyard-dev/stockyard-hub/internal/tools"
)

type HealthPoller struct {
	mgr  *tools.Manager
	db   *store.DB
	stop chan struct{}
}

func NewHealthPoller(mgr *tools.Manager, db *store.DB) *HealthPoller {
	return &HealthPoller{mgr: mgr, db: db, stop: make(chan struct{})}
}

func (p *HealthPoller) Start() {
	go func() {
		// Initial poll after 5 seconds
		time.Sleep(5 * time.Second)
		p.poll()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.poll()
			case <-p.stop:
				return
			}
		}
	}()
	log.Printf("[health] Poller started (every 30s)")
}

func (p *HealthPoller) Stop() {
	close(p.stop)
}

func (p *HealthPoller) poll() {
	statuses := p.mgr.Discover()
	for _, s := range statuses {
		if !s.Installed || !s.Running {
			continue
		}
		start := time.Now()
		status := checkToolHealth(s.Port)
		ms := int(time.Since(start).Milliseconds())
		p.db.RecordHealth(s.Slug, status, ms)
	}
}

func checkToolHealth(port int) string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://localhost:" + itoa(port) + "/api/health")
	if err != nil {
		// Try /health as fallback
		resp, err = client.Get("http://localhost:" + itoa(port) + "/health")
		if err != nil {
			return "unhealthy"
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return "healthy"
	}
	return "unhealthy"
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
