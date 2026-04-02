package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
}

type InstalledTool struct {
	Slug        string
	BinaryPath  string
	DataDir     string
	PID         int
	InstalledAt time.Time
	LastStarted time.Time
	AutoStart   bool
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS installed_tools (
		slug TEXT PRIMARY KEY,
		binary_path TEXT NOT NULL,
		data_dir TEXT NOT NULL,
		pid INTEGER DEFAULT 0,
		installed_at TEXT DEFAULT (datetime('now')),
		last_started TEXT DEFAULT '',
		auto_start INTEGER DEFAULT 0
	)`)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	)`)
	if err != nil {
		return nil, fmt.Errorf("create config table: %w", err)
	}

	return &DB{db: db}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) MarkInstalled(slug, binaryPath, dataDir string) error {
	_, err := d.db.Exec(`INSERT OR REPLACE INTO installed_tools (slug, binary_path, data_dir, installed_at)
		VALUES (?, ?, ?, datetime('now'))`, slug, binaryPath, dataDir)
	return err
}

func (d *DB) MarkUninstalled(slug string) error {
	_, err := d.db.Exec(`DELETE FROM installed_tools WHERE slug = ?`, slug)
	return err
}

func (d *DB) SetPID(slug string, pid int) error {
	_, err := d.db.Exec(`UPDATE installed_tools SET pid = ?, last_started = datetime('now') WHERE slug = ?`, pid, slug)
	return err
}

func (d *DB) ClearPID(slug string) error {
	_, err := d.db.Exec(`UPDATE installed_tools SET pid = 0 WHERE slug = ?`, slug)
	return err
}

func (d *DB) SetAutoStart(slug string, auto bool) error {
	v := 0
	if auto {
		v = 1
	}
	_, err := d.db.Exec(`UPDATE installed_tools SET auto_start = ? WHERE slug = ?`, v, slug)
	return err
}

func (d *DB) GetInstalled(slug string) *InstalledTool {
	row := d.db.QueryRow(`SELECT slug, binary_path, data_dir, pid, installed_at, last_started, auto_start
		FROM installed_tools WHERE slug = ?`, slug)
	var t InstalledTool
	var installedAt, lastStarted string
	var autoStart int
	if err := row.Scan(&t.Slug, &t.BinaryPath, &t.DataDir, &t.PID, &installedAt, &lastStarted, &autoStart); err != nil {
		return nil
	}
	t.InstalledAt, _ = time.Parse("2006-01-02 15:04:05", installedAt)
	t.LastStarted, _ = time.Parse("2006-01-02 15:04:05", lastStarted)
	t.AutoStart = autoStart == 1
	return &t
}

func (d *DB) ListInstalled() []InstalledTool {
	rows, err := d.db.Query(`SELECT slug, binary_path, data_dir, pid, installed_at, last_started, auto_start
		FROM installed_tools ORDER BY slug`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []InstalledTool
	for rows.Next() {
		var t InstalledTool
		var installedAt, lastStarted string
		var autoStart int
		if err := rows.Scan(&t.Slug, &t.BinaryPath, &t.DataDir, &t.PID, &installedAt, &lastStarted, &autoStart); err != nil {
			continue
		}
		t.InstalledAt, _ = time.Parse("2006-01-02 15:04:05", installedAt)
		t.LastStarted, _ = time.Parse("2006-01-02 15:04:05", lastStarted)
		t.AutoStart = autoStart == 1
		result = append(result, t)
	}
	return result
}

func (d *DB) SetConfig(key, value string) error {
	_, err := d.db.Exec(`INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`, key, value)
	return err
}

func (d *DB) GetConfig(key string) string {
	var value string
	d.db.QueryRow(`SELECT value FROM config WHERE key = ?`, key).Scan(&value)
	return value
}
