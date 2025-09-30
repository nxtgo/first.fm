package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"first.fm/internal/bot"
	"first.fm/internal/logger"
	"first.fm/internal/persistence/sqlc"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	lastfmKey := os.Getenv("LASTFM_API_KEY")

	if token == "" || lastfmKey == "" {
		panic("DISCORD_TOKEN and LASTFM_API_KEY must be set")
	}

	q, db, err := sqlc.Start(context.Background(), "database.db")
	if err != nil {
		logger.Fatalf("%v", err)
	}
	defer db.Close()

	bot, err := bot.New(token, lastfmKey, q)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err = bot.Run(ctx); err != nil {
		logger.Fatalf("%v", err)
	}
}
