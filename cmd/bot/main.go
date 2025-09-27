package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.fm/internal/bot/bot"
	"go.fm/internal/bot/logging"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	lastfmKey := os.Getenv("LASTFM_API_KEY")
	dbPath := getEnv("DATABASE_PATH", "file:database.db?_foreign_keys=on")

	if token == "" || lastfmKey == "" {
		logging.Fatal("DISCORD_TOKEN and LASTFM_API_KEY must be set")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	b, err := bot.NewBot(ctx, token, dbPath, lastfmKey)
	if err != nil {
		logging.Fatalf("failed to create bot: %v", err)
	}

	if err := b.Run(ctx); err != nil {
		logging.Fatalf("bot stopped with error: %v", err)
	}
}
