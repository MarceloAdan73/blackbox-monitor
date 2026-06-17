//go:build !sqlite

package main

type noopStore struct {
	logs []LogEntry
}

func openStore(_ string) (Store, error) {
	return &noopStore{}, nil
}

func (n *noopStore) SaveLog(entry LogEntry) error {
	n.logs = append(n.logs, entry)
	return nil
}

func (n *noopStore) GetRecentLogs(limit int) ([]LogEntry, error) {
	if len(n.logs) == 0 {
		return nil, nil
	}
	start := len(n.logs) - limit
	if start < 0 {
		start = 0
	}
	return n.logs[start:], nil
}

func (n *noopStore) Close() error { return nil }
