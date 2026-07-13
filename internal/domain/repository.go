package domain

import (
	"context"
)

// UserRepository определяет контракт для работы с хранилищем пользователей.
type UserRepository interface {
	Create(ctx context.Context, user *User) (int, error)
	ByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	Update(ctx context.Context, user *User) error
}

// HabitRepository определяет контракт для работы с хранилищем привычек.
type HabitRepository interface {
	Create(ctx context.Context, habit *Habit) (int, error)
	ByUserID(ctx context.Context, userID int) ([]*Habit, error)
	ByID(ctx context.Context, id int) (*Habit, error)
	Update(ctx context.Context, habit *Habit) error
	Delete(ctx context.Context, id int) error
}
