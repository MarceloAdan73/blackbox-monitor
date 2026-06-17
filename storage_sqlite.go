//go:build sqlite

package main

import (
	"fmt"

	"github.com/marcelo-adan/blackbox-monitor/internal/storage"
)

type sqliteStore struct {
	dbPath string
	store  *storage.Store
}

func openStore(dbPath string) (Store, error) {
	return &sqliteStore{dbPath: dbPath}, nil
}

func (s *sqliteStore) SaveLog(entry LogEntry) error {
	if s.store == nil {
		var err error
		s.store, err = storage.New(s.dbPath)
		if err != nil {
			return fmt.Errorf("error al abrir base de datos: %w", err)
		}
	}
	return s.store.SaveLog(storage.LogEntry{
		Name:      entry.Name,
		URL:       entry.URL,
		Online:    entry.Online,
		HTTPCode:  entry.HTTPCode,
		LatencyMs: entry.LatencyMs,
		ErrorMsg:  entry.ErrorMsg,
		CheckedAt: entry.CheckedAt,
	})
}

func (s *sqliteStore) GetRecentLogs(limit int) ([]LogEntry, error) {
	if s.store == nil {
		return nil, nil
	}
	entries, err := s.store.GetRecentLogs(limit)
	if err != nil {
		return nil, err
	}
	logs := make([]LogEntry, len(entries))
	for i, e := range entries {
		logs[i] = LogEntry{
			Name:      e.Name,
			URL:       e.URL,
			Online:    e.Online,
			HTTPCode:  e.HTTPCode,
			LatencyMs: e.LatencyMs,
			ErrorMsg:  e.ErrorMsg,
			CheckedAt: e.CheckedAt,
		}
	}
	return logs, nil
}

func (s *sqliteStore) Close() error {
	if s.store != nil {
		return s.store.Close()
	}
	return nil
}
