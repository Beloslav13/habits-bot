package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/beloslav13/habits-bot/internal/domain"
)

// ErrUserNotFound возвращается, когда пользователь не найден в БД.
var ErrUserNotFound = errors.New("user not found")

// TODO: этап 8 — заменить SQL-запрос на каждом сообщении на Redis-кеш (sync.Map как временный вариант)

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

const updateUserSQL = `
UPDATE users                                                                                                                                                       
SET first_name = $1, last_name = $2, username = $3, language_code = $4, is_premium = $5                                                                            
WHERE id = $6                                                                                                                                                      
RETURNING updated_at
`

// UserRepo — реализация domain.UserRepository для PostgreSQL.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo создаёт репозиторий пользователей.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create добавляет нового пользователя в БД и возвращает его ID.
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

// ByTelegramID ищет пользователя по Telegram ID.
// Если пользователь не найден, возвращает ErrUserNotFound.
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

// Update обновляет пользователя со всеми полями кроме telegram_id
func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	row := r.db.QueryRowContext(ctx, updateUserSQL, user.FirstName, user.LastName,
		user.Username, user.Language, user.IsPremium, user.ID,
	)
	err := row.Scan(&user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}
