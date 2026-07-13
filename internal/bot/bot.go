package bot

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

const baseUrl = "https://api.telegram.org"

type Bot struct {
	token   string
	baseUrl string
	client  *http.Client
	log     *slog.Logger
}

func New(token string, log *slog.Logger) *Bot {
	return &Bot{
		token:   token,
		baseUrl: baseUrl,
		client:  &http.Client{},
		log:     log,
	}
}

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
			b.log.Error("get updates failed", "error", err)
			time.Sleep(time.Second)
			continue
		}

		for _, u := range upd {
			offset = u.UpdateID + 1
			b.handleUpdate(ctx, u)
		}
	}
}
