package tools

// Tool represents a Stockyard tool in the catalog.
type Tool struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Tagline  string `json:"tagline"`
	Port     int    `json:"port"`
	Category string `json:"category"`
	ProPrice string `json:"pro_price"`
}

// Status represents the runtime state of a tool.
type Status struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Tagline   string `json:"tagline"`
	Port      int    `json:"port"`
	Category  string `json:"category"`
	Installed bool   `json:"installed"`
	Running   bool   `json:"running"`
	PID       int    `json:"pid,omitempty"`
	Health    string `json:"health"` // "healthy", "unhealthy", "stopped", "not_installed"
}
