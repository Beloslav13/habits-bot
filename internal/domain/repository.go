package domain

import (
	"context"
)

// UserRepository определяет контракт для работы с хранилищем пользователей.
type UserRepository interface {
	Create(ctx context.Context, user *User) (int, error)
	ByTelegramID(ctx context.Context, telegramID int64) (*User, error)
}

// HabitRepository определяет контракт для работы с хранилищем привычек.
type HabitRepository interface {
	Create(ctx context.Context, habit *Habit) (int, error)
	ByUserID(ctx context.Context, userID int) ([]*Habit, error)
}
