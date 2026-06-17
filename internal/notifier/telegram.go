package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type telegramMessage struct {
	ChatID              string `json:"chat_id"`
	Text                string `json:"text"`
	ParseMode           string `json:"parse_mode"`
	DisableNotification bool   `json:"disable_notification"`
}

func sendTelegramMessage(botToken, chatID, text string) error {
	body := telegramMessage{
		ChatID:              chatID,
		Text:                text,
		ParseMode:           "Markdown",
		DisableNotification: false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error encoding message: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("error sending telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		return fmt.Errorf("telegram API error (status %d): %v", resp.StatusCode, result)
	}

	return nil
}

func SendStartupNotification(botToken, chatID string, siteCount int) error {
	message := fmt.Sprintf(`🚀 *BlackBox Monitor*

✅ Monitor iniciado correctamente
📡 Monitoreando *%d sitios*
⏰ Intervalo configurado
🔔 Alertas activadas`,
		siteCount)

	return sendTelegramMessage(botToken, chatID, message)
}

func SendStateChangeAlert(botToken, chatID, siteName, oldStatus, newStatus, details string) error {
	emoji := "✅"
	if newStatus == "Offline" {
		emoji = "❌"
	}

	message := fmt.Sprintf(`%s *BlackBox Monitor*

*Site:* %s
*Status:* %s → %s
*Details:* %s
*Time:* %s`,
		emoji, siteName, oldStatus, newStatus, details, time.Now().Format("15:04:05"))

	return sendTelegramMessage(botToken, chatID, message)
}
