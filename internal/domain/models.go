package domain

import "time"

// User представляет пользователя Telegram, зарегистрированного в боте.
// Поля с *string являются опциональными (NULL в БД).
type User struct {
	ID         int
	TelegramID int64
	FirstName  string
	LastName   *string
	Username   *string
	Language   *string
	IsPremium  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Habit представляет привычку, которую отслеживает пользователь.
// Description — опциональное поле (NULL в БД).
type Habit struct {
	ID          int
	UserID      int
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
