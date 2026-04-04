package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/stockyard-dev/stockyard-hub/internal/store"
	"github.com/stockyard-dev/stockyard-hub/internal/tools"
)

type Poller struct {
	mgr      *tools.Manager
	db       *store.DB
	interval time.Duration
	stop     chan struct{}
}

func NewPoller(mgr *tools.Manager, db *store.DB, interval time.Duration) *Poller {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Poller{mgr: mgr, db: db, interval: interval, stop: make(chan struct{})}
}

func (p *Poller) Start() {
	go func() {
		p.poll()
		ticker := time.NewTicker(p.interval)
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
	log.Printf("[poller] health checks every %s", p.interval)
}

func (p *Poller) Stop() {
	close(p.stop)
}

func (p *Poller) poll() {
	statuses := p.mgr.Discover()
	for _, st := range statuses {
		if !st.Installed {
			continue
		}
		status := "stopped"
		responseMs := 0
		if st.Running {
			start := time.Now()
			client := &http.Client{Timeout: 3 * time.Second}
			resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", st.Port))
			responseMs = int(time.Since(start).Milliseconds())
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					status = "healthy"
				} else {
					status = "unhealthy"
				}
			} else {
				status = "unhealthy"
			}
		}
		p.db.RecordHealth(st.Slug, status, responseMs)
	}
	p.db.PruneHealth(7)
}
