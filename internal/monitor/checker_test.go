package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
}

func TestCheckSite_Online(t *testing.T) {
	statusCodes := []int{200, 201, 204, 301, 302, 399}
	for _, code := range statusCodes {
		ts := newTestServer(code)
		defer ts.Close()

		result := CheckSite("test", ts.URL, 5000)
		if !result.Online {
			t.Errorf("expected Online=true for status %d, got Online=%v, HTTPCode=%d", code, result.Online, result.HTTPCode)
		}
		if result.HTTPCode != code {
			t.Errorf("expected HTTPCode=%d, got %d", code, result.HTTPCode)
		}
	}
}

func TestCheckSite_Offline(t *testing.T) {
	statusCodes := []int{400, 403, 404, 500, 502, 503}
	for _, code := range statusCodes {
		ts := newTestServer(code)
		defer ts.Close()

		result := CheckSite("test", ts.URL, 5000)
		if result.Online {
			t.Errorf("expected Online=false for status %d, got Online=%v", code, result.Online)
		}
		if result.HTTPCode != code {
			t.Errorf("expected HTTPCode=%d, got %d", code, result.HTTPCode)
		}
	}
}

func TestCheckSite_Error(t *testing.T) {
	result := CheckSite("test", "http://127.0.0.1:1", 100)
	if result.Online {
		t.Error("expected Online=false for invalid address")
	}
	if result.Error == nil {
		t.Error("expected non-nil error for invalid address")
	}
}

func TestCheckSite_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	result := CheckSite("test", ts.URL, 50)
	if result.Online {
		t.Error("expected Online=false for timeout")
	}
	if result.Error == nil {
		t.Error("expected non-nil error for timeout")
	}
}

func TestCheckSite_NameAndURL(t *testing.T) {
	ts := newTestServer(200)
	defer ts.Close()

	result := CheckSite("Mi Sitio", ts.URL, 5000)
	if result.Name != "Mi Sitio" {
		t.Errorf("expected Name='Mi Sitio', got '%s'", result.Name)
	}
	if result.URL != ts.URL {
		t.Errorf("expected URL='%s', got '%s'", ts.URL, result.URL)
	}
}

func TestCheckSite_Latency(t *testing.T) {
	ts := newTestServer(200)
	defer ts.Close()

	result := CheckSite("test", ts.URL, 5000)
	if result.LatencyMs < 0 {
		t.Error("expected non-negative latency")
	}
}
