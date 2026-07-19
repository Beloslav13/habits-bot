package bot

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/beloslav13/habits-bot/internal/domain"
	"github.com/beloslav13/habits-bot/internal/storage/postgres"
)

// validateName проверяет название привычки на пустоту и длину.
// Возвращает сообщение об ошибке или пустую строку, если название валидно.
func validateName(name string) string {
	if strings.TrimSpace(name) == "" {
		return msgErrHabitEmptyName
	}
	if len(name) > 256 {
		return msgErrHabitNameTooLong
	}
	return ""

}

// habitCreate создаёт новую привычку для пользователя.
func (b *Bot) habitCreate(ctx context.Context, userID, chatID int, args []string) {
	if len(args) == 0 {
		b.reply(chatID, msgErrHabitNoName)
		return
	}
	name := strings.Join(args, " ")
	msgErr := validateName(name)
	if msgErr != "" {
		b.reply(chatID, msgErr)
		return
	}
	h := &domain.Habit{
		UserID: userID,
		Name:   name,
	}
	id, err := b.store.Habit.Create(ctx, h)
	if err != nil {
		b.log.Error("habitCreate: create habit failed", "error", err, "user_id", userID, "chat_id", chatID)
		b.reply(chatID, msgErrHabitNotCreated)
		return
	}
	b.log.Info("habitCreate: create habit", "habit_id", id, "user_id", userID, "chat_id", chatID)
	b.reply(chatID, msgHabitCreated)

}

// habitUpdate обновляет название привычки по ID.
// Проверяет, что привычка существует и принадлежит пользователю.
func (b *Bot) habitUpdate(ctx context.Context, userID, chatID int, args []string) {
	if len(args) < 2 {
		b.reply(chatID, msgErrHabitEditNotArgs)
		return
	}
	habitID, err := strconv.Atoi(args[0])
	if err != nil {
		b.reply(chatID, msgErrHabitIDNotNumber)
		return
	}

	h, err := b.store.Habit.ByID(ctx, habitID)
	if err != nil {
		if errors.Is(err, postgres.ErrHabitNotFound) {
			b.log.Error("habitUpdate: habit not found", "user_id", userID, "chat_id", chatID, "habit_id", habitID)
			b.reply(chatID, msgErrHabitNotFound)
			return
		}
		b.log.Error("habitUpdate: get habit by id failed", "error", err, "user_id", userID, "chat_id", chatID, "habit_id", habitID)
		b.reply(chatID, msgErrHabitNotUpdated)
		return
	}
	if userID != h.UserID {
		b.log.Warn("habitUpdate: habit is anothers", "habit_id", h.ID, "user_id", userID)
		b.reply(chatID, msgErrHabitNotYours)
		return
	}

	name := strings.Join(args[1:], " ")
	msgErr := validateName(name)
	if msgErr != "" {
		b.reply(chatID, msgErr)
		return
	}

	h.Name = name
	err = b.store.Habit.Update(ctx, h)
	if err != nil {
		if errors.Is(err, postgres.ErrHabitNotFound) {
			b.log.Error("habitUpdate: habit not found", "user_id", userID, "chat_id", chatID, "habit_id", habitID)
			b.reply(chatID, msgErrHabitNotFound)
			return
		}
		b.log.Error("habitUpdate: update habit failed", "error", err, "user_id", userID, "chat_id", chatID, "habit_id", habitID)
		b.reply(chatID, msgErrHabitNotUpdated)
	}
	b.log.Info("habitUpdate: ok", "user_id", userID, "chat_id", chatID, "habit_id", habitID)
	b.reply(chatID, msgHabitUpdated)

}

// habitList выводит список всех привычек пользователя.
func (b *Bot) habitList(ctx context.Context, userID, chatID int) {
	habits, err := b.store.Habit.ByUserID(ctx, userID)
	if err != nil {
		b.log.Error("habitList: get habit by user id failed", "user_id", userID, "chat_id", chatID)
		b.reply(chatID, msgErrHabitNotReceived)
		return
	}

	if len(habits) == 0 {
		b.log.Info("habitList: no habits found", "user_id", userID, "chat_id", chatID)
		b.reply(chatID, msgHabitEmptyList)
		return
	}

	m := habitListKeyboard(habits, userID)
	b.replyWithKeyboard(chatID, "Твои привычки:", &m)

}

// habitDelete удаляет привычку пользователя по ID.
// Проверяет, что привычка существует и принадлежит пользователю.
func (b *Bot) habitDelete(ctx context.Context, userID, chatID int, args []string) {
	if len(args) == 0 {
		b.reply(chatID, msgErrHabitNoID)
		return
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		b.reply(chatID, msgErrHabitIDNotNumber)
		return
	}
	h, err := b.store.Habit.ByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrHabitNotFound) {
			b.log.Error("habitDelete: habit not found", "user_id", userID, "chat_id", chatID, "habit_id", id)
			b.reply(chatID, msgErrHabitNotFound)
			return
		}
		b.log.Error("habitDelete: get habit by id failed", "error", err, "user_id", userID, "chat_id", chatID, "habit_id", id)
		b.reply(chatID, msgErrHabitNotDeleted)
		return
	}
	if userID != h.UserID {
		b.log.Warn("habitDelete: habit is anothers", "habit_id", h.ID, "user_id", userID)
		b.reply(chatID, msgErrHabitNotYours)
		return
	}
	err = b.store.Habit.Delete(ctx, id)
	if err != nil {
		b.log.Error("habitDelete: delete habit failed", "error", err, "user_id", userID, "chat_id", chatID, "habit_id", id)
		b.reply(chatID, msgErrHabitNotDeleted)
		return
	}
	b.log.Info("habitDelete: delete habit", "habit_id", id, "user_id", userID)
	b.reply(chatID, msgHabitDeleted)

}

func (b *Bot) help(chatID int) {
	b.reply(chatID,
		"Доступные команды:\n"+
			"/habits — список привычек\n"+
			"/newhabit название — создать\n"+
			"/edithabit ID название — изменить\n"+
			"/deletehabit ID — удалить",
	)
}
