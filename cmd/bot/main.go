package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/beloslav13/habits-bot/internal/bot"
	"github.com/beloslav13/habits-bot/internal/config"
	"github.com/beloslav13/habits-bot/internal/storage"
	"github.com/beloslav13/habits-bot/pkg/logger"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.Env)

	log.Info("starting bot", "env", cfg.Env)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	store, err := storage.New(cfg.DataBase.DSN())
	if err != nil {
		log.Error("storage init failed", "error", err)
		return
	}
	defer store.Close()

	b := bot.New(cfg.Bot.Token, log)
	err = b.Run(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Error("bot exited with error", "error", err, "env", cfg.Env)
		return
	}

	log.Info("shutting down", "env", cfg.Env)
}
