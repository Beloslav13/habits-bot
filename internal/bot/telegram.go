package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Response struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	Text string `json:"text"`
	Chat *Chat  `json:"chat"`
	From *User  `json:"from"`
}
type Chat struct {
	ID int `json:"id"`
}
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type CallbackQuery struct {
	ID   string `json:"id"`
	User *User  `json:"from"`
	Data string `json:"data"`
}

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
