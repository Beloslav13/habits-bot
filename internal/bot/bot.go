package bot

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/beloslav13/habits-bot/internal/storage"
)

const baseUrl = "https://api.telegram.org"

// Bot управляет жизненным циклом Telegram-бота: long polling, обработка сообщений.
type Bot struct {
	token   string
	baseUrl string
	client  *http.Client
	log     *slog.Logger
	store   *storage.Storage
}

// New создаёт экземпляр бота с заданным токеном, логгером и хранилищем.
func New(token string, log *slog.Logger, store *storage.Storage) *Bot {
	return &Bot{
		token:   token,
		baseUrl: baseUrl,
		client:  &http.Client{},
		log:     log,
		store:   store,
	}
}

// Run запускает цикл long polling. Блокирует выполнение до отмены контекста.
// Возвращает ctx.Err() при graceful shutdown или ошибку при критическом сбое.
func (b *Bot) Run(ctx context.Context) error {
	var offset int
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		upd, err := b.getUpdates(offset)
		if err != nil {
			b.log.Error("get updates failed", "offset", offset, "error", err)
			time.Sleep(time.Second)
			continue
		}

		for _, u := range upd {
			offset = u.UpdateID + 1
			b.handleUpdate(ctx, u)
		}
	}
}
