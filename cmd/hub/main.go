package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-hub/internal/server"
	"github.com/stockyard-dev/stockyard-hub/internal/store"
	"github.com/stockyard-dev/stockyard-hub/internal/tools"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9800"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = os.Getenv("HOME") + "/.stockyard-hub"
	}
	binDir := os.Getenv("BIN_DIR")
	if binDir == "" {
		binDir = os.Getenv("HOME") + "/.stockyard/bin"
	}
	os.MkdirAll(dataDir, 0755)

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("hub: open database: %v", err)
	}
	defer db.Close()

	mgr := tools.NewManager(binDir, dataDir)
	srv := server.New(mgr, db)

	// Start health poller
	poller := server.NewHealthPoller(mgr, db)
	poller.Start()
	defer poller.Stop()

	fmt.Printf("\n  Stockyard Hub — Tool Management Dashboard\n")
	fmt.Printf("  ─────────────────────────────────────────\n")
	fmt.Printf("  Dashboard:  http://localhost:%s/ui\n", port)
	fmt.Printf("  API:        http://localhost:%s/api\n", port)
	fmt.Printf("  Data:       %s\n", dataDir)
	fmt.Printf("  Bin dir:    %s\n", binDir)
	fmt.Printf("  ─────────────────────────────────────────\n\n")

	log.Printf("hub: listening on :%s", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("hub: %v", err)
	}
}
