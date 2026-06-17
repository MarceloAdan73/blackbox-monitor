package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	Purple  = lipgloss.Color("#7C3AED")
	Green   = lipgloss.Color("#10B981")
	Red     = lipgloss.Color("#EF4444")
	Yellow  = lipgloss.Color("#F59E0B")
	Gray    = lipgloss.Color("#6B7280")
	White   = lipgloss.Color("#FFFFFF")
	Cyan    = lipgloss.Color("#06B6D4")
	Orange  = lipgloss.Color("#F97316")
	DarkBg  = lipgloss.Color("#1F2937")
	DarkBg2 = lipgloss.Color("#111827")
)

func RenderTitle() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(Purple).
		Render("◆ BlackBox Monitor v1.0")

	subtitle := lipgloss.NewStyle().
		Foreground(Cyan).
		Italic(true).
		Render("Web Health Dashboard")

	block := lipgloss.JoinVertical(lipgloss.Center, title, subtitle)

	box := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(Purple).
		Align(lipgloss.Center).
		Width(48)

	return box.Render(block)
}

func RenderSiteBox(name, url string, online bool, httpCode int, latencyMs int64, checkedAt time.Time, errMsg string, certExpiryDays int) string {
	borderColor := Green
	if !online {
		borderColor = Red
	}
	if errMsg != "" {
		borderColor = Orange
	}

	statusColor := Green
	statusText := "● ONLINE"
	statusIcon := "✅"
	if !online && errMsg == "" {
		statusColor = Red
		statusText = "● OFFLINE"
		statusIcon = "❌"
	} else if errMsg != "" {
		statusColor = Orange
		statusText = "● ERROR"
		statusIcon = "⚠️"
	}

	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(White)
	urlStyle := lipgloss.NewStyle().Foreground(Gray).Italic(true)
	statusStyle := lipgloss.NewStyle().Bold(true).Foreground(statusColor)

	latencyColor := Green
	switch {
	case latencyMs >= 1000:
		latencyColor = Red
	case latencyMs >= 300:
		latencyColor = Yellow
	}
	latencyStyle := lipgloss.NewStyle().Bold(true).Foreground(latencyColor)

	lines := []string{
		nameStyle.Render("📡 " + name),
		urlStyle.Render(url),
		"",
		statusStyle.Render(statusIcon+" "+statusText) + "  " + lipgloss.NewStyle().Foreground(White).Render(fmt.Sprintf("Código: %d", httpCode)),
		latencyStyle.Render(fmt.Sprintf("⚡ %dms", latencyMs)) + "  " + lipgloss.NewStyle().Foreground(Gray).Italic(true).Render(checkedAt.Format("15:04:05")),
	}
	if errMsg != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(Orange).Render("✖ "+errMsg))
	}
	if certExpiryDays > 0 && certExpiryDays <= 30 {
		warnStyle := lipgloss.NewStyle().Foreground(Yellow).Bold(true)
		lines = append(lines, warnStyle.Render(fmt.Sprintf("🔒 SSL expira en %d días", certExpiryDays)))
	} else if certExpiryDays == 0 {
		warnStyle := lipgloss.NewStyle().Foreground(Red).Bold(true)
		lines = append(lines, warnStyle.Render("🔒 SSL expira HOY"))
	} else if certExpiryDays < 0 {
		warnStyle := lipgloss.NewStyle().Foreground(Red).Bold(true)
		lines = append(lines, warnStyle.Render("🔒 SSL expirado"))
	}

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(52)

	return box.Render(content)
}

func RenderSummaryBox(onlineCount, total int, avgLatency int64, totalOk, totalFail int) string {
	barWidth := 20
	onlineBars := int(float64(onlineCount) / float64(total) * float64(barWidth))
	if onlineBars > barWidth {
		onlineBars = barWidth
	}
	offlineBars := barWidth - onlineBars

	bar := lipgloss.NewStyle().Foreground(Green).Render(strings.Repeat("█", onlineBars)) +
		lipgloss.NewStyle().Foreground(Red).Render(strings.Repeat("█", offlineBars))

	percent := float64(onlineCount) / float64(total) * 100

	avgColor := Green
	switch {
	case avgLatency >= 1000:
		avgColor = Red
	case avgLatency >= 300:
		avgColor = Yellow
	}

	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(Purple).Render("📊 DASHBOARD"),
		bar,
		"",
		lipgloss.NewStyle().Foreground(Gray).Render("Operativos:") + " " + lipgloss.NewStyle().Bold(true).Foreground(Green).Render(fmt.Sprintf("%d/%d (%.0f%%)", onlineCount, total, percent)),
		lipgloss.NewStyle().Foreground(Gray).Render("Promedio:") + " " + lipgloss.NewStyle().Bold(true).Foreground(avgColor).Render(fmt.Sprintf("%dms", avgLatency)),
		lipgloss.NewStyle().Foreground(Gray).Render("Chequeos:") + " " + lipgloss.NewStyle().Bold(true).Foreground(White).Render(fmt.Sprintf("%d exitosos, %d fallidos", totalOk, totalFail)),
	}

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Purple).
		Padding(0, 2).
		Width(48).
		Align(lipgloss.Center)

	return lipgloss.NewStyle().Foreground(Gray).Render(box.Render(content))
}

func RenderFooter(nextCheck time.Time, interval int) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(Gray).
		Border(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(DarkBg2).
		Padding(0, 1).
		Width(48).
		Align(lipgloss.Center)

	content := fmt.Sprintf("⏰ Próximo: %s  |  Intervalo: %ds  |  %s",
		lipgloss.NewStyle().Foreground(Cyan).Render(nextCheck.Format("15:04:05")),
		interval,
		lipgloss.NewStyle().Foreground(Gray).Italic(true).Render("Ctrl+C para salir"),
	)

	return footerStyle.Render(content)
}

func RenderShutdown() string {
	return lipgloss.NewStyle().
		Foreground(Purple).
		Bold(true).
		Render("\n  ◆ Monitoreo detenido. ¡Hasta luego!")
}

func RenderError(msg string) string {
	box := lipgloss.NewStyle().
		Foreground(Red).
		Bold(true).
		Border(lipgloss.NormalBorder()).
		BorderForeground(Red).
		Padding(0, 1).
		Render("✖ Error: " + msg)
	return box
}

func RenderSeparator() string {
	return lipgloss.NewStyle().
		Foreground(DarkBg2).
		Render(strings.Repeat("─", 52))
}
