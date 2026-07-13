package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beloslav13/habits-bot/internal/domain"
	"github.com/beloslav13/habits-bot/internal/storage/postgres"
)

const (
	msgWelcome     = "Привет! Я трекер привычек. Список команд появится позже."
	msgUnknownCmd  = "Неизвестная команда."
	msgUseCommands = "Используйте команды, например /start"
)

// handleUpdate обрабатывает одно обновление от Telegram:
// регистрирует пользователя (если новый) и маршрутизирует сообщение.
func (b *Bot) handleUpdate(ctx context.Context, upd Update) {
	if upd.Message == nil || upd.Message.Text == "" {
		return
	}

	b.ensureUser(ctx, upd.Message.From)
	b.routeMessage(upd.Message.Chat.ID, upd.Message.Text)
}

// ensureUser проверяет существование пользователя в БД и создаёт при необходимости.
func (b *Bot) ensureUser(ctx context.Context, tgUser *User) {
	_, err := b.store.User.ByTelegramID(ctx, int64(tgUser.ID))
	if err == nil {
		return
	}
	if !errors.Is(err, postgres.ErrUserNotFound) {
		b.log.Error("ensureUser: lookup failed", "error", err)
		return
	}

	user := domain.User{
		TelegramID: int64(tgUser.ID),
		FirstName:  tgUser.FirstName,
		IsPremium:  tgUser.IsPremium,
	}
	if tgUser.Username != "" {
		user.Username = &tgUser.Username
	}
	if tgUser.LastName != "" {
		user.LastName = &tgUser.LastName
	}
	if tgUser.Language != "" {
		user.Language = &tgUser.Language
	}

	_, err = b.store.User.Create(ctx, &user)
	if err != nil {
		b.log.Error("ensureUser: create failed", "error", err)
	}
}

// routeMessage направляет сообщение нужному обработчику в зависимости от команды.
func (b *Bot) routeMessage(chatID int, text string) {
	if strings.HasPrefix(text, "/") {
		cmd := strings.Fields(text)[0]
		cmd = strings.TrimPrefix(cmd, "/")
		switch cmd {
		case "start":
			b.reply(chatID, msgWelcome)
		default:
			b.reply(chatID, msgUnknownCmd)
		}
		return
	}
	b.reply(chatID, msgUseCommands)
}

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
