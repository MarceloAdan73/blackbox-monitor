package main

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIVersion(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFiles.ReadFile("web/static/index.html")
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html, got %s", ct)
	}
}

func TestStaticFS(t *testing.T) {
	staticFS, err := fs.Sub(staticFiles, "web/static")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := staticFS.Open("index.html"); err != nil {
		t.Errorf("index.html not found in embedded FS: %v", err)
	}
}

func TestAPIStatusEndpoint(t *testing.T) {
	state := NewDashboardState()
	state.Update([]SiteStatus{
		{Name: "test", URL: "http://test.com", Online: true, HTTPCode: 200, LatencyMs: 100, CheckedAt: "12:00:00"},
	}, 1, 1, 100, 5, 0)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		sites, online, total, avgLat, lastCheck, checks, failures := state.Get()
		if sites == nil {
			sites = []SiteStatus{}
		}

		percent := 0.0
		if total > 0 {
			percent = float64(online) / float64(total) * 100
		}

		resp := map[string]interface{}{
			"sites":          sites,
			"online":         online,
			"total":          total,
			"percent":        percent,
			"avg_latency":    avgLat,
			"last_check":     lastCheck,
			"total_checks":   checks,
			"total_failures": failures,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/status")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result["online"] != float64(1) {
		t.Errorf("expected online=1, got %v", result["online"])
	}
	if result["total"] != float64(1) {
		t.Errorf("expected total=1, got %v", result["total"])
	}
	if result["percent"] != float64(100) {
		t.Errorf("expected percent=100, got %v", result["percent"])
	}
	if result["total_failures"] != float64(0) {
		t.Errorf("expected total_failures=0, got %v", result["total_failures"])
	}

	sites := result["sites"].([]interface{})
	if len(sites) != 1 {
		t.Fatalf("expected 1 site, got %d", len(sites))
	}

	site := sites[0].(map[string]interface{})
	if site["name"] != "test" {
		t.Errorf("expected site name 'test', got %v", site["name"])
	}
	if site["online"] != true {
		t.Errorf("expected site online=true, got %v", site["online"])
	}
}

func TestAPIStatusEmptyState(t *testing.T) {
	state := NewDashboardState()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		sites, online, total, avgLat, lastCheck, checks, failures := state.Get()
		if sites == nil {
			sites = []SiteStatus{}
		}

		percent := 0.0
		if total > 0 {
			percent = float64(online) / float64(total) * 100
		}

		resp := map[string]interface{}{
			"sites":          sites,
			"online":         online,
			"total":          total,
			"percent":        percent,
			"avg_latency":    avgLat,
			"last_check":     lastCheck,
			"total_checks":   checks,
			"total_failures": failures,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/status")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result["total"] != float64(0) {
		t.Errorf("expected total=0 for empty state, got %v", result["total"])
	}
	if result["percent"] != float64(0) {
		t.Errorf("expected percent=0 for empty state, got %v", result["percent"])
	}

	sites := result["sites"].([]interface{})
	if len(sites) != 0 {
		t.Errorf("expected 0 sites for empty state, got %d", len(sites))
	}
}

func TestAPIStatusNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("ok"))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("expected 404 for unknown path, got %d", resp.StatusCode)
	}
}
