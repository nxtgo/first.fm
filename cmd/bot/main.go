package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"first.fm/internal/bot"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	lastfmKey := os.Getenv("LASTFM_API_KEY")

	if token == "" || lastfmKey == "" {
		panic("DISCORD_TOKEN and LASTFM_API_KEY must be set")
	}

	bot, err := bot.New(token, lastfmKey)
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err = bot.Run(ctx); err != nil {
		panic(err)
	}
}
