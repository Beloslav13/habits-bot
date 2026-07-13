package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Response — ответ Telegram API на getUpdates.
type Response struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

// Update — одно обновление (сообщение или callback).
type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

// Message — сообщение от пользователя.
type Message struct {
	Text string `json:"text"`
	Chat *Chat  `json:"chat"`
	From *User  `json:"from"`
}

// Chat — информация о чате.
type Chat struct {
	ID int `json:"id"`
}

// User — информация о пользователе Telegram.
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Language  string `json:"language_code"`
	IsPremium bool   `json:"is_premium"`
}

// CallbackQuery — данные нажатия на inline-кнопку.
type CallbackQuery struct {
	ID   string `json:"id"`
	User *User  `json:"from"`
	Data string `json:"data"`
}

// getUpdates получает непрочитанные обновления от Telegram API (long polling).
func (b *Bot) getUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("%s/bot%s/getUpdates?offset=%d&timeout=10", b.baseUrl, b.token, offset)
	resp, err := b.client.Get(url)
	if err != nil {
		b.log.Error("getUpdates: network error")
		return nil, fmt.Errorf("getUpdates: request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b.log.Error("getUpdates: bad status", "status", resp.StatusCode)
		return nil, fmt.Errorf("getUpdates: status %d", resp.StatusCode)
	}

	r, err := io.ReadAll(resp.Body)
	if err != nil {
		b.log.Error("getUpdates: read body", "error", err)
		return nil, fmt.Errorf("getUpdates: read body failed")
	}

	var response Response
	if err := json.Unmarshal(r, &response); err != nil {
		b.log.Error("getUpdates: unmarshal", "error", err)
		return nil, fmt.Errorf("getUpdates: unmarshal failed")
	}
	return response.Result, nil
}
