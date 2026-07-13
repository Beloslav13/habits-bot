package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/beloslav13/habits-bot/internal/bot"
	"github.com/beloslav13/habits-bot/internal/config"
	"github.com/beloslav13/habits-bot/pkg/logger"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.Env)

	log.Info("starting bot", "env", cfg.Env)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b := bot.New(cfg.Bot.Token, log)
	if err := b.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("bot exited with error", "error", "env", cfg.Env)
		return
	}

	log.Info("shutting down", "env", cfg.Env)
}
