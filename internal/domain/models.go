package domain

import "time"

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

type Habit struct {
	ID          int
	UserID      int
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
