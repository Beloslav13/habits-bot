package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/beloslav13/habits-bot/internal/domain"
)

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

type HabitRepo struct {
	db *sql.DB
}

func NewHabitRepo(db *sql.DB) *HabitRepo {
	return &HabitRepo{db: db}
}

func (r *HabitRepo) Create(ctx context.Context, habit *domain.Habit) (int, error) {
	row := r.db.QueryRowContext(ctx, createHabitSQL, habit.UserID, habit.Name, habit.Description)
	err := row.Scan(&habit.ID, &habit.CreatedAt, &habit.UpdatedAt)
	if err != nil {
		return 0, err
	}
	return habit.ID, nil
}

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
