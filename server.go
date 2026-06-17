package main

import (
	"embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"time"
)

//go:embed web/static
var staticFiles embed.FS

var startTime = time.Now()

func startServer(addr string, state *DashboardState, store Store) {
	mux := http.NewServeMux()

	staticFS, _ := fs.Sub(staticFiles, "web/static")
	fileServer := http.FileServer(http.FS(staticFS))

	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, _ := staticFiles.ReadFile("web/static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _, total, _, _, _, _ := state.Get()
		uptime := time.Since(startTime).Round(time.Second).String()

		resp := map[string]interface{}{
			"status":          "ok",
			"uptime":          uptime,
			"sites_monitored": total,
			"version":         "1.3",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/export", func(w http.ResponseWriter, r *http.Request) {
		logs, err := store.GetRecentLogs(1000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=blackbox-export.csv")

		writer := csv.NewWriter(w)
		writer.Write([]string{"name", "url", "online", "http_code", "latency_ms", "error_msg", "checked_at"})

		for _, log := range logs {
			online := "false"
			if log.Online {
				online = "true"
			}
			writer.Write([]string{
				log.Name,
				log.URL,
				online,
				fmt.Sprintf("%d", log.HTTPCode),
				fmt.Sprintf("%d", log.LatencyMs),
				log.ErrorMsg,
				log.CheckedAt.Format(time.RFC3339),
			})
		}
		writer.Flush()
	})

	fmt.Printf("\n  ◆ Dashboard web disponible en: http://localhost%s\n\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("Error al iniciar servidor web: %v\n", err)
	}
}
