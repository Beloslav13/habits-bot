package storage

import (
	"database/sql"
	"embed"

	_ "github.com/lib/pq"

	"github.com/beloslav13/habits-bot/internal/domain"
	"github.com/beloslav13/habits-bot/internal/storage/postgres"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func New(dsn string) (*Storage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(migrations)
	err = goose.Up(db, "migrations")
	if err != nil {
		return nil, err
	}

	s := &Storage{
		db:    db,
		User:  postgres.NewUserRepo(db),
		Habit: postgres.NewHabitRepo(db),
	}
	return s, nil
}

type Storage struct {
	db    *sql.DB
	User  domain.UserRepository
	Habit domain.HabitRepository
}

func (s *Storage) Close() error {
	return s.db.Close()
}
