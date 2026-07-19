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
	ChatID      int                   `json:"chat_id"`
	MessageID   int                   `json:"message_id,omitempty"`
	Text        string                `json:"text"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type sendMessageResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
	} `json:"result"`
}

// sendMessage отправляет текстовое сообщение через Telegram API.
func (b *Bot) sendMessage(chatID int, text string, markup *InlineKeyboardMarkup) (int, error) {
	msg := MessageDTO{ChatID: chatID, MessageID: 0, Text: text, ReplyMarkup: markup}
	url := fmt.Sprintf("%s/bot%s/sendMessage", b.baseUrl, b.token)

	body, err := json.Marshal(msg)
	if err != nil {
		b.log.Error("sendMessage: marshal", "error", err)
		return 0, fmt.Errorf("sendMessage: marshal failed")
	}

	r, err := b.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		b.log.Error("sendMessage: network error")
		return 0, fmt.Errorf("sendMessage: request failed")
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		b.log.Error("sendMessage: bad status", "status", r.StatusCode)
		return 0, fmt.Errorf("sendMessage: status %d", r.StatusCode)
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		b.log.Error("sendMessage: read body", "error", err)
		return 0, fmt.Errorf("sendMessage: read body failed")
	}

	var resp sendMessageResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		b.log.Error("sendMessage: unmarshal", "error", err)
		return 0, fmt.Errorf("sendMessage: unmarshal failed")
	}
	if !resp.Ok {
		b.log.Error("sendMessage: bad response", "status", resp.Result.MessageID)
		return 0, fmt.Errorf("sendMessage: bad response")
	}
	return resp.Result.MessageID, nil
}

func (b *Bot) editMessage(chatID, messageID int, text string, markup *InlineKeyboardMarkup) error {
	msg := MessageDTO{ChatID: chatID, MessageID: messageID, Text: text, ReplyMarkup: markup}
	url := fmt.Sprintf("%s/bot%s/editMessageText", b.baseUrl, b.token)

	body, err := json.Marshal(msg)
	if err != nil {
		b.log.Error("editMessage: marshal", "error", err)
		return fmt.Errorf("editMessage: marshal failed")
	}

	r, err := b.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		b.log.Error("editMessage: network error")
		return fmt.Errorf("editMessage: request failed")
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		b.log.Error("editMessage: bad status", "status", r.StatusCode)
		return fmt.Errorf("editMessage: status %d", r.StatusCode)
	}

	_, err = io.Copy(io.Discard, r.Body)
	if err != nil {
		b.log.Error("editMessage: drain body", "error", err)
		return fmt.Errorf("sendMessage: drain body failed")
	}

	return nil
}

// reply — обёртка над sendMessage с логированием ошибок.
func (b *Bot) reply(chatID int, text string) int {
	msgID, err := b.sendMessage(chatID, text, nil)
	if err != nil {
		b.log.Error("reply failed", "chat_id", chatID, "error", err)
	}
	return msgID
}

func (b *Bot) replyWithKeyboard(chatID int, text string, markup *InlineKeyboardMarkup) int {
	msgID, err := b.sendMessage(chatID, text, markup)
	if err != nil {
		b.log.Error("replyWithKeyboard failed", "chat_id", chatID, "error", err)
	}
	return msgID
}

func (b *Bot) editReply(chatID, messageID int, text string) {
	if err := b.editMessage(chatID, messageID, text, nil); err != nil {
		b.log.Error("editReply failed", "chat_id", chatID, "message_id", messageID, "error", err)
	}
}
