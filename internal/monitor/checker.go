package monitor

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

type SiteResult struct {
	Name           string
	URL            string
	Online         bool
	HTTPCode       int
	LatencyMs      int64
	Error          error
	CheckedAt      time.Time
	CertExpiry     time.Time
	CertExpiryDays int
}

func CheckSite(name, url string, timeoutMs int) SiteResult {
	result := SiteResult{
		Name:      name,
		URL:       url,
		CheckedAt: time.Now(),
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	start := time.Now()
	resp, err := client.Get(url)
	latency := time.Since(start)
	result.LatencyMs = latency.Milliseconds()

	if err != nil {
		result.Error = err
		result.Online = false
		return result
	}
	defer resp.Body.Close()

	result.HTTPCode = resp.StatusCode
	result.Online = resp.StatusCode >= 200 && resp.StatusCode < 400

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		result.CertExpiry = cert.NotAfter
		result.CertExpiryDays = int(time.Until(cert.NotAfter).Hours() / 24)
	} else if result.Online {
		parts := extractHost(url)
		if parts != "" {
			result.CertExpiryDays = -1
			result.Error = fmt.Errorf("no se pudo obtener certificado TLS para %s", parts)
		}
	}

	return result
}

func extractHost(rawURL string) string {
	// Simple host extraction without importing net/url to keep it lightweight
	if len(rawURL) == 0 {
		return ""
	}
	start := 0
	if rawURL[:8] == "https://" {
		start = 8
	} else if rawURL[:7] == "http://" {
		start = 7
	}
	for i := start; i < len(rawURL); i++ {
		if rawURL[i] == '/' || rawURL[i] == ':' {
			return rawURL[start:i]
		}
	}
	return rawURL[start:]
}
