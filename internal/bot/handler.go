package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	msgWelcome     = "Привет! Я трекер привычек. Список команд появится позже."
	msgUnknownCmd  = "Неизвестная команда."
	msgUseCommands = "Используйте команды, например /start"
)

func (b *Bot) handleUpdate(ctx context.Context, upd Update) {
	if upd.Message == nil || upd.Message.Text == "" {
		return
	}

	text := upd.Message.Text
	if strings.HasPrefix(text, "/") {
		cmd := strings.Fields(text)[0]
		cmd = strings.TrimPrefix(cmd, "/")
		switch cmd {
		case "start":
			b.reply(upd.Message.Chat.ID, msgWelcome)
		default:
			b.reply(upd.Message.Chat.ID, msgUnknownCmd)
		}
		return
	}

	b.reply(upd.Message.Chat.ID, msgUseCommands)

}

type MessageDTO struct {
	ChatID int    `json:"chat_id"`
	Text   string `json:"text"`
}

func (b *Bot) sendMessage(chatID int, text string) error {
	msg := MessageDTO{chatID, text}
	url := fmt.Sprintf("%s/bot%s/sendMessage", b.baseUrl, b.token)

	body, err := json.Marshal(msg)
	if err != nil {
		b.log.Error("failed marshalling message", "error", err)
		return err
	}

	r, err := b.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		b.log.Error("failed sending message", "error", err)
		return err
	}

	_, err = io.Copy(io.Discard, r.Body)
	if err != nil {
		b.log.Error("failed Copy message", "error", err)
		return err
	}
	defer r.Body.Close()

	return nil
}

func (b *Bot) reply(chatID int, text string) {
	if err := b.sendMessage(chatID, text); err != nil {
		b.log.Error("reply failed", "chat_id", chatID, "error", err)
	}
}
