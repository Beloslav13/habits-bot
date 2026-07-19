package bot

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
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

	states   map[int]userState
	statesMu sync.RWMutex
}

type userState struct {
	name    string
	habitID int
}

func (b *Bot) setState(name string, userID, habitID int) {
	b.statesMu.Lock()
	defer b.statesMu.Unlock()
	b.states[userID] = userState{name, habitID}
}

func (b *Bot) getState(userID int) (userState, bool) {
	b.statesMu.RLock()
	defer b.statesMu.RUnlock()
	state, ok := b.states[userID]
	return state, ok
}

func (b *Bot) clearState(userID int) {
	b.statesMu.Lock()
	defer b.statesMu.Unlock()
	delete(b.states, userID)
}

// New создаёт экземпляр бота.
func New(token string, log *slog.Logger, store *storage.Storage) *Bot {
	return &Bot{
		token:   token,
		baseUrl: baseUrl,
		client:  &http.Client{},
		log:     log,
		store:   store,
		states:  make(map[int]userState),
	}
}

// Run запускает цикл long polling. Блокирует выполнение до отмены контекста.
// Возвращает ctx.Err() при graceful shutdown или ошибку при критическом сбое.
func (b *Bot) Run(ctx context.Context) error {
	if err := b.setMyCommands(); err != nil {
		b.log.Warn("setMyCommands failed", "error", err)
	} else {
		b.log.Info("commands registered")
	}

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
