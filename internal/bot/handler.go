package bot

import (
	"context"
	"errors"
	"strings"

	"github.com/beloslav13/habits-bot/internal/domain"
	"github.com/beloslav13/habits-bot/internal/storage/postgres"
)

// parseCommands разбирает текст сообщения на команду и аргументы.
// Возвращает пустую строку, если сообщение не является командой.
func parseCommands(text string) (string, []string) {
	if !strings.HasPrefix(text, "/") {
		return "", []string{}
	}
	args := strings.Fields(text)
	cmd := strings.TrimPrefix(args[0], "/")
	return cmd, args[1:]
}

// handleUpdate обрабатывает одно обновление от Telegram:
// регистрирует пользователя (если новый) и маршрутизирует сообщение.
func (b *Bot) handleUpdate(ctx context.Context, upd Update) {
	if upd.Message == nil || upd.Message.Text == "" {
		return
	}

	userID := b.ensureUser(ctx, upd.Message.From)
	if userID == 0 {
		return
	}
	b.routeMessage(ctx, userID, upd.Message.Chat.ID, upd.Message.Text)
}

// ensureUser проверяет существование пользователя в БД и создаёт при необходимости.
func (b *Bot) ensureUser(ctx context.Context, tgUser *User) int {
	user, err := b.store.User.ByTelegramID(ctx, int64(tgUser.ID))
	if err == nil {
		return user.ID
	}
	if !errors.Is(err, postgres.ErrUserNotFound) {
		b.log.Error("ensureUser: lookup failed", "error", err)
		return 0
	}

	u := domain.User{
		TelegramID: int64(tgUser.ID),
		FirstName:  tgUser.FirstName,
		IsPremium:  tgUser.IsPremium,
	}
	if tgUser.Username != "" {
		u.Username = &tgUser.Username
	}
	if tgUser.LastName != "" {
		u.LastName = &tgUser.LastName
	}
	if tgUser.Language != "" {
		u.Language = &tgUser.Language
	}

	id, err := b.store.User.Create(ctx, &u)
	if err != nil {
		b.log.Error("ensureUser: create failed", "error", err)
		return 0
	}
	return id
}

// routeMessage направляет сообщение нужному обработчику в зависимости от команды.
func (b *Bot) routeMessage(ctx context.Context, userID, chatID int, text string) {
	cmd, args := parseCommands(text)
	switch cmd {
	case "start":
		b.reply(chatID, msgWelcome)
		b.help(chatID)
	case "help":
		b.help(chatID)
	case "habits":
		b.habitList(ctx, userID, chatID)
	case "newhabit":
		b.habitCreate(ctx, userID, chatID, args)
	case "edithabit":
		b.habitUpdate(ctx, userID, chatID, args)
	case "deletehabit":
		b.habitDelete(ctx, userID, chatID, args)
	default:
		b.reply(chatID, msgUnknownCommand)
	}
}
