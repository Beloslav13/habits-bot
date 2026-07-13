package main

import (
	"context"
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
	err := b.Run(ctx)
	if err != nil {
		log.Error("bot exited with error", "error", err.Error())
		return
	}

	log.Info("shutting down")
}
