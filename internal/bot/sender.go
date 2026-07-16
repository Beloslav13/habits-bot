package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MessageDTO — тело запроса к Telegram API sendMessage.
type MessageDTO struct {
	ChatID int    `json:"chat_id"`
	Text   string `json:"text"`
}

// sendMessage отправляет текстовое сообщение через Telegram API.
func (b *Bot) sendMessage(chatID int, text string) error {
	msg := MessageDTO{chatID, text}
	url := fmt.Sprintf("%s/bot%s/sendMessage", b.baseUrl, b.token)

	body, err := json.Marshal(msg)
	if err != nil {
		b.log.Error("sendMessage: marshal", "error", err)
		return fmt.Errorf("sendMessage: marshal failed")
	}

	r, err := b.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		b.log.Error("sendMessage: network error")
		return fmt.Errorf("sendMessage: request failed")
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		b.log.Error("sendMessage: bad status", "status", r.StatusCode)
		return fmt.Errorf("sendMessage: status %d", r.StatusCode)
	}

	_, err = io.Copy(io.Discard, r.Body)
	if err != nil {
		b.log.Error("sendMessage: drain body", "error", err)
		return fmt.Errorf("sendMessage: drain body failed")
	}

	return nil
}

// reply — обёртка над sendMessage с логированием ошибок.
func (b *Bot) reply(chatID int, text string) {
	if err := b.sendMessage(chatID, text); err != nil {
		b.log.Error("reply failed", "chat_id", chatID, "error", err)
	}
}
