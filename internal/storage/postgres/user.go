package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/beloslav13/habits-bot/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

const getUserByTelegramIDSQL = `
SELECT id, telegram_id, first_name, last_name, username, language_code, is_premium, created_at, updated_at
FROM users
WHERE telegram_id = $1
`

const createUserSQL = `
INSERT INTO users (telegram_id, first_name, last_name, username, language_code, is_premium)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at
`

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) (int, error) {
	row := r.db.QueryRowContext(ctx, createUserSQL,
		user.TelegramID, user.FirstName, user.LastName,
		user.Username, user.Language, user.IsPremium,
	)
	err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (r *UserRepo) ByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, getUserByTelegramIDSQL, telegramID)
	user := &domain.User{}
	err := row.Scan(
		&user.ID, &user.TelegramID, &user.FirstName, &user.LastName, &user.Username,
		&user.Language, &user.IsPremium, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
