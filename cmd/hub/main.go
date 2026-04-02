package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-hub/internal/server"
	"github.com/stockyard-dev/stockyard-hub/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8600"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./hub-data"
	}
	licenseKey := os.Getenv("STOCKYARD_LICENSE_KEY")

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("hub: open database: %v", err)
	}
	defer db.Close()

	// Restore license key from DB if not in env
	if licenseKey == "" {
		if saved := db.GetConfig("license_key"); saved != "" {
			licenseKey = saved
		}
	}

	srv := server.New(db, dataDir, licenseKey)

	fmt.Printf("\n  Stockyard Hub\n")
	fmt.Printf("  ─────────────────────────────────\n")
	fmt.Printf("  Dashboard:  http://localhost:%s/ui\n", port)
	fmt.Printf("  API:        http://localhost:%s/api\n", port)
	fmt.Printf("  Data:       %s\n", dataDir)
	if licenseKey != "" {
		fmt.Printf("  License:    active\n")
	} else {
		fmt.Printf("  License:    not set (paste in dashboard)\n")
	}
	fmt.Printf("  ─────────────────────────────────\n\n")

	log.Printf("hub: listening on :%s", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("hub: %v", err)
	}
}
