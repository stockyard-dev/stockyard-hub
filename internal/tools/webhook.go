package tools

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// WebhookEvent represents a tool lifecycle event.
type WebhookEvent struct {
	Event     string `json:"event"` // tool.started, tool.stopped, tool.installed, tool.unhealthy
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Port      int    `json:"port"`
	Timestamp string `json:"timestamp"`
}

// FireWebhook sends a tool event to the configured webhook URL.
func FireWebhook(event, slug, name string, port int) {
	url := os.Getenv("HUB_WEBHOOK_URL")
	if url == "" {
		return // No webhook configured
	}

	payload := WebhookEvent{
		Event:     event,
		Slug:      slug,
		Name:      name,
		Port:      port,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	data, _ := json.Marshal(payload)

	go func() {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(url, "application/json", bytes.NewReader(data))
		if err != nil {
			log.Printf("[webhook] failed: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			log.Printf("[webhook] %s returned %d", url, resp.StatusCode)
		}
	}()
}
