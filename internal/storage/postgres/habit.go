package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/beloslav13/habits-bot/internal/domain"
)

// ErrHabitNotFound возвращается, когда привычка не найдена в БД.
var ErrHabitNotFound = errors.New("habit not found")

const getHabitByUserIDSQL = `
SELECT id, user_id, name, description, created_at, updated_at
FROM habits
WHERE user_id = $1
`

const createHabitSQL = `
INSERT INTO habits (user_id, name, description)
VALUES ($1, $2, $3)
RETURNING id, created_at, updated_at
`

const getHabitByIDSQL = `
SELECT id, user_id, name, description, created_at, updated_at
FROM habits
WHERE id = $1
`

const updateHabitSQL = `
UPDATE habits SET name = $2, description = $3 WHERE id = $1                                                                                    
RETURNING updated_at
`

const deleteHabitSQL = `
DELETE FROM habits
WHERE id = $1
`

// HabitRepo — реализация domain.HabitRepository для PostgreSQL.
type HabitRepo struct {
	db *sql.DB
}

// NewHabitRepo создаёт репозиторий привычек.
func NewHabitRepo(db *sql.DB) *HabitRepo {
	return &HabitRepo{db: db}
}

// Create добавляет новую привычку в БД и возвращает её ID.
func (r *HabitRepo) Create(ctx context.Context, habit *domain.Habit) (int, error) {
	row := r.db.QueryRowContext(ctx, createHabitSQL, habit.UserID, habit.Name, habit.Description)
	err := row.Scan(&habit.ID, &habit.CreatedAt, &habit.UpdatedAt)
	if err != nil {
		return 0, err
	}
	return habit.ID, nil
}

// ByUserID возвращает все привычки пользователя.
// Если привычек нет, возвращает пустой слайс (не ошибку).
func (r *HabitRepo) ByUserID(ctx context.Context, userID int) ([]*domain.Habit, error) {
	rows, err := r.db.QueryContext(ctx, getHabitByUserIDSQL, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []*domain.Habit
	for rows.Next() {
		habit := &domain.Habit{}
		if err := rows.Scan(
			&habit.ID, &habit.UserID, &habit.Name,
			&habit.Description, &habit.CreatedAt, &habit.UpdatedAt,
		); err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return habits, nil
}

// ByID возвращает конкретную привычку
// Если привычки нет, возвращает ошибку ErrHabitNotFound
func (r *HabitRepo) ByID(ctx context.Context, id int) (*domain.Habit, error) {
	row := r.db.QueryRowContext(ctx, getHabitByIDSQL, id)
	habit := &domain.Habit{}
	err := row.Scan(
		&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.CreatedAt, &habit.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHabitNotFound
		}
		return nil, err
	}
	return habit, nil
}

// Update обновляет название и описание привычки.
// Возвращает ErrHabitNotFound, если привычка не найдена.
func (r *HabitRepo) Update(ctx context.Context, habit *domain.Habit) error {
	row := r.db.QueryRowContext(ctx, updateHabitSQL, habit.ID, habit.Name, habit.Description)
	err := row.Scan(&habit.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrHabitNotFound
		}
		return err
	}
	return nil
}

// Delete удаляет привычку
// Если привычки нет, возвращает ошибку ErrHabitNotFound
func (r *HabitRepo) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, deleteHabitSQL, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrHabitNotFound
	}
	return nil
}
