//go:build sqlite

package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type LogEntry struct {
	ID        int64
	Name      string
	URL       string
	Online    bool
	HTTPCode  int
	LatencyMs int64
	ErrorMsg  string
	CheckedAt time.Time
}

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (s *Store) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS check_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		online INTEGER NOT NULL,
		http_code INTEGER,
		latency_ms INTEGER,
		error_msg TEXT,
		checked_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := s.db.Exec(query)
	return err
}

func (s *Store) SaveLog(entry LogEntry) error {
	query := `INSERT INTO check_logs (name, url, online, http_code, latency_ms, error_msg, checked_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	var errMsg string
	if entry.ErrorMsg != "" {
		errMsg = entry.ErrorMsg
	}
	_, err := s.db.Exec(query, entry.Name, entry.URL, entry.Online, entry.HTTPCode, entry.LatencyMs, errMsg, entry.CheckedAt)
	return err
}

func (s *Store) GetRecentLogs(limit int) ([]LogEntry, error) {
	query := `SELECT id, name, url, online, http_code, latency_ms, error_msg, checked_at 
	          FROM check_logs ORDER BY checked_at DESC LIMIT ?`
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		err := rows.Scan(&entry.ID, &entry.Name, &entry.URL, &entry.Online,
			&entry.HTTPCode, &entry.LatencyMs, &entry.ErrorMsg, &entry.CheckedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, entry)
	}
	return logs, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
