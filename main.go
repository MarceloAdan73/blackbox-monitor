package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marcelo-adan/blackbox-monitor/internal/monitor"
	"github.com/marcelo-adan/blackbox-monitor/internal/notifier"
	"github.com/marcelo-adan/blackbox-monitor/internal/ui"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Interval int `yaml:"interval"`
	Telegram struct {
		Enabled  bool   `yaml:"enabled"`
		BotToken string `yaml:"bot_token"`
		ChatID   string `yaml:"chat_id"`
	} `yaml:"telegram"`
	Sites []struct {
		Name    string `yaml:"name"`
		URL     string `yaml:"url"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"sites"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error al parsear %s: %w", path, err)
	}

	if len(cfg.Sites) == 0 {
		return nil, fmt.Errorf("el archivo %s no contiene sitios configurados", path)
	}

	for _, site := range cfg.Sites {
		if site.URL == "" {
			return nil, fmt.Errorf("una de las URLs en %s está vacía", path)
		}
		if site.Timeout <= 0 {
			return nil, fmt.Errorf("timeout inválido para %s", site.Name)
		}
	}

	return &cfg, nil
}

func runChecks(cfg *Config, store Store, state *DashboardState, prevStates map[string]bool) (onlineCount int, totalOk, totalFail int) {
	var totalLatency int64
	onlineCount = 0
	totalOk = 0
	totalFail = 0

	siteStatuses := make([]SiteStatus, 0, len(cfg.Sites))

	for _, site := range cfg.Sites {
		result := monitor.CheckSite(site.Name, site.URL, site.Timeout)

		errMsg := ""
		if result.Error != nil {
			errMsg = result.Error.Error()
			totalFail++
		} else {
			totalOk++
			if result.Online {
				onlineCount++
			}
			totalLatency += result.LatencyMs
		}

		fmt.Println(ui.RenderSiteBox(site.Name, site.URL, result.Online, result.HTTPCode, result.LatencyMs, result.CheckedAt, errMsg, result.CertExpiryDays))
		fmt.Println()

		store.SaveLog(LogEntry{
			Name:      site.Name,
			URL:       site.URL,
			Online:    result.Online,
			HTTPCode:  result.HTTPCode,
			LatencyMs: result.LatencyMs,
			ErrorMsg:  errMsg,
			CheckedAt: result.CheckedAt,
		})

		currentOnline := result.Online && errMsg == ""
		if prevStates != nil {
			if prev, seen := prevStates[site.Name]; seen {
				if prev != currentOnline && cfg.Telegram.Enabled && cfg.Telegram.BotToken != "" && cfg.Telegram.ChatID != "" {
					oldStatus := "Offline"
					if prev {
						oldStatus = "Online"
					}
					newStatus := "Offline"
					if currentOnline {
						newStatus = "Online"
					}
					details := fmt.Sprintf("HTTP %d | %dms", result.HTTPCode, result.LatencyMs)
					if errMsg != "" {
						details = errMsg
					}
					go notifier.SendStateChangeAlert(cfg.Telegram.BotToken, cfg.Telegram.ChatID, site.Name, oldStatus, newStatus, details)
				}
			}
			prevStates[site.Name] = currentOnline
		}

		certExpiry := ""
		if !result.CertExpiry.IsZero() {
			certExpiry = result.CertExpiry.Format("2006-01-02")
		}
		siteStatuses = append(siteStatuses, SiteStatus{
			Name:           site.Name,
			URL:            site.URL,
			Online:         result.Online,
			HTTPCode:       result.HTTPCode,
			LatencyMs:      result.LatencyMs,
			Error:          errMsg,
			CheckedAt:      result.CheckedAt.Format("15:04:05"),
			CertExpiry:     certExpiry,
			CertExpiryDays: result.CertExpiryDays,
		})
	}

	avgLatency := int64(0)
	if len(cfg.Sites) > 0 {
		avgLatency = totalLatency / int64(len(cfg.Sites))
	}

	if state != nil {
		state.Update(siteStatuses, onlineCount, len(cfg.Sites), avgLatency, totalOk+totalFail, totalFail)
	}

	fmt.Println(ui.RenderSummaryBox(onlineCount, len(cfg.Sites), avgLatency, totalOk, totalFail))
	fmt.Println()

	return
}

func main() {
	configPath := flag.String("config", "config.yaml", "ruta del archivo de configuración")
	intervalFlag := flag.Int("interval", 0, "intervalo en segundos entre chequeos (0 = usar config)")
	portFlag := flag.String("port", ":8080", "puerto del dashboard web (vacío = desactivado)")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Println(ui.RenderError(err.Error()))
		os.Exit(1)
	}

	store, err := openStore("blackbox.db")
	if err != nil {
		fmt.Println(ui.RenderError("error al iniciar almacenamiento: " + err.Error()))
		os.Exit(1)
	}
	defer store.Close()

	state := NewDashboardState()

	if cfg.Telegram.Enabled && cfg.Telegram.BotToken != "" && cfg.Telegram.ChatID != "" {
		go notifier.SendStartupNotification(cfg.Telegram.BotToken, cfg.Telegram.ChatID, len(cfg.Sites))
	}

	if *portFlag != "" {
		go startServer(*portFlag, state, store)
	}

	prevStates := make(map[string]bool)

	interval := cfg.Interval
	if *intervalFlag > 0 {
		interval = *intervalFlag
	}
	if interval <= 0 {
		interval = 60
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Print("\033[2J\033[H")
	fmt.Println(ui.RenderTitle())
	fmt.Println()

	runChecks(cfg, store, state, prevStates)

	for {
		nextCheck := time.Now().Add(time.Duration(interval) * time.Second)
		fmt.Println(ui.RenderFooter(nextCheck, interval))

		select {
		case <-ticker.C:
			fmt.Print("\033[2J\033[H")
			fmt.Println(ui.RenderTitle())
			fmt.Println()
			runChecks(cfg, store, state, prevStates)
		case <-sigChan:
			fmt.Println()
			fmt.Println(ui.RenderShutdown())
			return
		}
	}
}
