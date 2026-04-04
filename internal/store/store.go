package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

type HealthRecord struct {
	ID         string `json:"id"`
	Tool       string `json:"tool"`
	Status     string `json:"status"`
	ResponseMs int    `json:"response_ms"`
	CheckedAt  string `json:"checked_at"`
}

type ActivityRecord struct {
	ID        string `json:"id"`
	Tool      string `json:"tool"`
	Action    string `json:"action"`
	Detail    string `json:"detail,omitempty"`
	CreatedAt string `json:"created_at"`
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dsn := filepath.Join(dataDir, "hub.db") + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS health_log (
			id TEXT PRIMARY KEY,
			tool TEXT NOT NULL,
			status TEXT NOT NULL,
			response_ms INTEGER DEFAULT 0,
			checked_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_health_tool ON health_log(tool)`,
		`CREATE INDEX IF NOT EXISTS idx_health_time ON health_log(checked_at)`,
		`CREATE TABLE IF NOT EXISTS activity (
			id TEXT PRIMARY KEY,
			tool TEXT NOT NULL,
			action TEXT NOT NULL,
			detail TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_time ON activity(created_at)`,
		`CREATE TABLE IF NOT EXISTS tool_state (
			slug TEXT PRIMARY KEY,
			last_status TEXT DEFAULT '',
			last_checked TEXT DEFAULT ''
		)`,
	} {
		if _, err := db.Exec(q); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string   { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) RecordHealth(tool, status string, responseMs int) {
	d.db.Exec(`INSERT INTO health_log (id,tool,status,response_ms,checked_at) VALUES (?,?,?,?,?)`,
		genID(), tool, status, responseMs, now())
	var lastStatus string
	d.db.QueryRow(`SELECT last_status FROM tool_state WHERE slug=?`, tool).Scan(&lastStatus)
	if lastStatus != status {
		if lastStatus != "" {
			d.LogActivity(tool, "health_changed", fmt.Sprintf("%s -> %s", lastStatus, status))
		}
		d.db.Exec(`INSERT OR REPLACE INTO tool_state (slug,last_status,last_checked) VALUES (?,?,?)`,
			tool, status, now())
	} else {
		d.db.Exec(`UPDATE tool_state SET last_checked=? WHERE slug=?`, now(), tool)
	}
}

func (d *DB) HealthHistory(tool string, limit int) []HealthRecord {
	if limit <= 0 {
		limit = 100
	}
	q := `SELECT id,tool,status,response_ms,checked_at FROM health_log`
	var args []any
	if tool != "" {
		q += ` WHERE tool=?`
		args = append(args, tool)
	}
	q += ` ORDER BY checked_at DESC LIMIT ?`
	args = append(args, limit)
	rows, err := d.db.Query(q, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []HealthRecord
	for rows.Next() {
		var r HealthRecord
		rows.Scan(&r.ID, &r.Tool, &r.Status, &r.ResponseMs, &r.CheckedAt)
		out = append(out, r)
	}
	return out
}

func (d *DB) LogActivity(tool, action, detail string) {
	d.db.Exec(`INSERT INTO activity (id,tool,action,detail,created_at) VALUES (?,?,?,?,?)`,
		genID(), tool, action, detail, now())
}

func (d *DB) RecentActivity(limit int) []ActivityRecord {
	if limit <= 0 {
		limit = 50
	}
	rows, err := d.db.Query(`SELECT id,tool,action,detail,created_at FROM activity ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ActivityRecord
	for rows.Next() {
		var a ActivityRecord
		rows.Scan(&a.ID, &a.Tool, &a.Action, &a.Detail, &a.CreatedAt)
		out = append(out, a)
	}
	return out
}

func (d *DB) PruneHealth(keepDays int) {
	if keepDays <= 0 {
		keepDays = 7
	}
	d.db.Exec(`DELETE FROM health_log WHERE checked_at < datetime('now', ?)`,
		fmt.Sprintf("-%d days", keepDays))
}
