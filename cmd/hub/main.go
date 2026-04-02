package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-hub/internal/server"
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
		binDir = "/usr/local/bin"
	}
	os.MkdirAll(dataDir, 0755)

	mgr := tools.NewManager(binDir, dataDir)
	srv := server.New(mgr)

	fmt.Printf("Stockyard Hub running on :%s\n", port)
	fmt.Printf("Dashboard: http://localhost:%s/ui\n", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
