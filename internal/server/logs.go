package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// readCLIConfig reads ~/.stockyard/config.json
func readCLIConfig() map[string]any {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(filepath.Join(home, ".stockyard", "config.json"))
	if err != nil {
		return map[string]any{}
	}
	var cfg map[string]any
	json.Unmarshal(data, &cfg)
	return cfg
}

// writeCLIConfig writes ~/.stockyard/config.json
func writeCLIConfig(cfg map[string]any) error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".stockyard")
	os.MkdirAll(dir, 0755)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0644)
}

// readToolLog reads the last N lines from a CLI-managed tool's log
func readToolLog(slug string, lines int) string {
	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".stockyard", "logs", slug+".log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return "(no logs available)"
	}
	allLines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if lines <= 0 {
		lines = 100
	}
	start := len(allLines) - lines
	if start < 0 {
		start = 0
	}
	return strings.Join(allLines[start:], "\n")
}
