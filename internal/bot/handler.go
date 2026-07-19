package bot

import (
	"context"
	"errors"
	"strconv"
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
	if upd.CallbackQuery != nil {
		b.routeCallback(ctx, upd)
		return
	}

	if upd.Message == nil || upd.Message.Text == "" {
		return
	}

	chatID := upd.Message.Chat.ID
	text := upd.Message.Text
	userID := b.ensureUser(ctx, upd.Message.From)
	if userID == 0 {
		return
	}

	state, hasState := b.getState(userID)
	if hasState {
		if strings.HasPrefix(text, "/") {
			b.clearState(userID)
		} else {
			b.handleState(ctx, state, userID, chatID, text)
			return
		}
	}
	b.routeMessage(ctx, userID, chatID, text)
}

func (b *Bot) handleState(ctx context.Context, state userState, userID, chatID int, text string) {
	switch state.name {
	case "creating_habit":
		b.finishCreating(ctx, state, userID, chatID, text)
	case "editing_habit":
		b.finishEditing(ctx, state, userID, chatID, text)
	}
}

// ensureUser проверяет существование пользователя в БД и создаёт при необходимости.
func (b *Bot) ensureUser(ctx context.Context, tgUser *User) int {
	user, err := b.store.User.ByTelegramID(ctx, tgUser.ID)
	if err == nil {
		return user.ID
	}
	if !errors.Is(err, postgres.ErrUserNotFound) {
		b.log.Error("ensureUser: lookup failed", "error", err)
		return 0
	}

	u := domain.User{
		TelegramID: tgUser.ID,
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

func (b *Bot) routeCallback(ctx context.Context, upd Update) {
	parts := strings.Split(upd.CallbackQuery.Data, "_") // всегда ["habit", "delete", "5", "123"]
	entity := parts[0]                                  // habit / habits
	action := parts[1]                                  // view / edit / delete / list / new
	id, _ := strconv.Atoi(parts[2])
	dbUserID, _ := strconv.Atoi(parts[3])

	cb := upd.CallbackQuery
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID
	switch entity {
	case "habit":
		switch action {
		case "view":
			// показать клавиатуру действий
			m := habitActionKeyboard(id, dbUserID)
			err := b.editMessage(chatID, msgID, "Что делаем?", &m)
			if err != nil {
				b.log.Error("routeCallback: view", "error", err, "action", action, "habit_id", id, "user_id", dbUserID)
			}
		case "delete":
			// показать подтверждение
			h, err := b.store.Habit.ByID(ctx, id)
			if err != nil {
				if errors.Is(err, postgres.ErrHabitNotFound) {
					b.showHabitsList(ctx, chatID, msgID, dbUserID) // привычка уже удалена
					return
				}
				b.log.Error("routeCallback: delete", "error", err, "action", action, "habit_id", id, "user_id", dbUserID)
				return
			}

			m := confirmDeleteKeyboard(id, dbUserID)
			err = b.editMessage(chatID, msgID, "Точно удалить «"+h.Name+"»?", &m)
			if err != nil {
				b.log.Error("routeCallback: delete -> editMessage", "error", err, "action", action, "habit_id", id, "user_id", dbUserID)
			}
			b.log.Info("routeCallback: confirm delete ok", "action", action, "habit_id", id, "user_id", dbUserID)
		case "confirmdelete":
			// удалить → показать список
			err := b.store.Habit.Delete(ctx, id)
			if err != nil {
				if errors.Is(err, postgres.ErrHabitNotFound) {
					b.log.Error("routeCallback: confirm_delete -> habit not found", "action", action, "habit_id", id, "user_id", dbUserID)
					return
				}
				b.log.Error("routeCallback: confirm_delete", "error", err, "action", action, "habit_id", id, "user_id", dbUserID)
				return
			}
			b.log.Info("routeCallback: delete habit_db ok", "action", action, "habit_id", id, "user_id", dbUserID)
			b.showHabitsList(ctx, chatID, msgID, dbUserID)
		case "edit":
			b.setState("editing_habit", dbUserID, id)
			b.editReply(chatID, msgID, "Введи новое название:")
		case "new":
			b.setState("creating_habit", dbUserID, 0)
			b.editReply(chatID, msgID, "Введи название новой привычки:")
		}
	case "habits":
		switch action {
		case "list":
			// показать список
			b.showHabitsList(ctx, chatID, msgID, dbUserID)
		}
	}

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
		b.help(chatID)
	}
}

func (b *Bot) showHabitsList(ctx context.Context, chatID, msgID, userID int) {
	habits, _ := b.store.Habit.ByUserID(ctx, userID)
	m := habitListKeyboard(habits, userID)
	if err := b.editMessage(chatID, msgID, "Твои привычки:", &m); err != nil {
		b.log.Error("showHabitsList", "error", err, "user_id", userID)
	}
}
