package bot

const (
	// Общие
	msgWelcome        = "Привет! Я трекер привычек. Список команд появится позже."
	msgUnknownCommand = "Неизвестная команда."
	msgUseCommands    = "Используйте команды, например /start"

	// Привычки: успех
	msgHabitCreated   = "Круто! Привычка создана."
	msgHabitDeleted   = "Привычка удалена."
	msgHabitEmptyList = "У тебя пока нет привычек, создай первую: /newhabit"
	msgHabitUpdated   = "Привычка изменена."

	// Привычки: ошибки
	msgErrHabitNotCreated  = "Упс, что-то пошло не так... Привычка не создана."
	msgErrHabitNotDeleted  = "Упс, что-то пошло не так... Привычка не удалена."
	msgErrHabitNotUpdated  = "Упс, что-то пошло не так... Привычка не изменена."
	msgErrHabitNotReceived = "Упс, что-то пошло не так... Привычки не получены."
	msgErrHabitIDNotNumber = "ID должен быть числом!"
	msgErrHabitNotFound    = "Привычка не найдена."
	msgErrHabitNotYours    = "Упс :(, это не твоя привычка."
	msgErrHabitNoID        = "Укажи какую привычку удалить, например: /deletehabit 42"
	msgErrHabitNoName      = "Укажи название привычки, например: /newhabit спорт"
	msgErrHabitEmptyName   = "Название привычки не может быть пустым."
	msgErrHabitNameTooLong = "Название привычки не должно превышать 256 символов."
	msgErrHabitEditNotArgs = "Укажите какую привычку и как ее переменовать: /edithabit 42 спорт"
)
