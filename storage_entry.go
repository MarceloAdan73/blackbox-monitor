package main

import "time"

type LogEntry struct {
	Name      string
	URL       string
	Online    bool
	HTTPCode  int
	LatencyMs int64
	ErrorMsg  string
	CheckedAt time.Time
}

type Store interface {
	SaveLog(entry LogEntry) error
	GetRecentLogs(limit int) ([]LogEntry, error)
	Close() error
}
