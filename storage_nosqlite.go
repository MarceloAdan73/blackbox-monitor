//go:build !sqlite

package main

type noopStore struct{}

func openStore(_ string) (Store, error) {
	return &noopStore{}, nil
}

func (n *noopStore) SaveLog(_ LogEntry) error                { return nil }
func (n *noopStore) GetRecentLogs(_ int) ([]LogEntry, error) { return nil, nil }
func (n *noopStore) Close() error                            { return nil }
