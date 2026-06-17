package main

import (
	"sync"
	"time"
)

type SiteStatus struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	Online         bool   `json:"online"`
	HTTPCode       int    `json:"http_code"`
	LatencyMs      int64  `json:"latency_ms"`
	Error          string `json:"error,omitempty"`
	CheckedAt      string `json:"checked_at"`
	CertExpiry     string `json:"cert_expiry"`
	CertExpiryDays int    `json:"cert_expiry_days"`
}

type DashboardState struct {
	mu            sync.RWMutex
	sites         []SiteStatus
	totalOnline   int
	totalSites    int
	avgLatency    int64
	lastCheck     string
	totalChecks   int
	totalFailures int
}

func NewDashboardState() *DashboardState {
	return &DashboardState{}
}

func (d *DashboardState) Update(sites []SiteStatus, online, total int, avgLat int64, checks, failures int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.sites = sites
	d.totalOnline = online
	d.totalSites = total
	d.avgLatency = avgLat
	d.lastCheck = time.Now().Format("2006-01-02 15:04:05")
	d.totalChecks = checks
	d.totalFailures = failures
}

func (d *DashboardState) Get() (sites []SiteStatus, online, total int, avgLat int64, lastCheck string, checks, failures int) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.sites, d.totalOnline, d.totalSites, d.avgLatency, d.lastCheck, d.totalChecks, d.totalFailures
}
