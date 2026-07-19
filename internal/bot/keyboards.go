package bot

import (
	"strconv"

	"github.com/beloslav13/habits-bot/internal/domain"
)

const (
	cbActionHabit  = "habit"
	cbActionHabits = "habits"

	cbEntityView          = "view"
	cbEntityEdit          = "edit"
	cbEntityDelete        = "delete"
	cbEntityConfirmDelete = "confirmdelete"
	cbEntityList          = "list"
	cbEntityNew           = "new"
)

func cbData(action, entity string, id, userID int) string {
	return action + "_" + entity + "_" + strconv.Itoa(id) + "_" + strconv.Itoa(userID)
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// habitListKeyboard — список привычек (каждая кнопкой), снизу [+ Создать].
func habitListKeyboard(habits []*domain.Habit, userID int) InlineKeyboardMarkup {
	rows := make([][]InlineKeyboardButton, 0, len(habits)+1)
	for _, h := range habits {
		rows = append(rows, []InlineKeyboardButton{{
			Text:         h.Name,
			CallbackData: cbData(cbActionHabit, cbEntityView, h.ID, userID),
		}})
	}
	rows = append(rows, []InlineKeyboardButton{{
		Text:         "+ Создать",
		CallbackData: cbData(cbActionHabit, cbEntityNew, 0, userID),
	}})
	return InlineKeyboardMarkup{InlineKeyboard: rows}
}

// habitActionKeyboard — действия над конкретной привычкой после выбора.
func habitActionKeyboard(habitID, userID int) InlineKeyboardMarkup {
	actions := [][]InlineKeyboardButton{
		{
			{Text: "✏️ Изменить", CallbackData: cbData(cbActionHabit, cbEntityEdit, habitID, userID)},
			{Text: "🗑 Удалить", CallbackData: cbData(cbActionHabit, cbEntityDelete, habitID, userID)},
		},
		{{Text: "⬅️ Назад", CallbackData: cbData(cbActionHabits, cbEntityList, 0, userID)}},
	}
	return InlineKeyboardMarkup{InlineKeyboard: actions}
}

// confirmDeleteKeyboard — подтверждение удаления.
func confirmDeleteKeyboard(habitID, userID int) InlineKeyboardMarkup {
	confirm := [][]InlineKeyboardButton{
		{{Text: "✅ Да, удалить", CallbackData: cbData(cbActionHabit, cbEntityConfirmDelete, habitID, userID)}},
		{{Text: "❌ Нет", CallbackData: cbData(cbActionHabits, cbEntityList, 0, userID)}},
	}
	return InlineKeyboardMarkup{InlineKeyboard: confirm}
}
