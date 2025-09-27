package bot

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	"github.com/nxtgo/arikawa/v3/api/cmdroute"
	"github.com/nxtgo/arikawa/v3/gateway"
	"github.com/nxtgo/arikawa/v3/state"
	"go.fm/internal/bot/bot/events"
	"go.fm/internal/bot/commands"
	"go.fm/internal/bot/lastfm"
	"go.fm/internal/bot/logging"
	"go.fm/internal/bot/persistence/sqlc"
)

// package for bot stuff idk.
// functions that open connections should also return
// a function to close that connection :pray:

type Bot struct {
	State    *state.State
	Query    *sqlc.Queries
	Database *sql.DB
	LastFM   *lastfm.Services
	Cache    *lastfm.Cache
	Registry *events.Registry
}

func NewBot(ctx context.Context, token, dbPath, lastfmKey string) (*Bot, error) {
	logging.Info("starting bot...")

	// database
	q, db, err := sqlc.Start(ctx, dbPath)
	if err != nil {
		return nil, err
	}

	// lastfm
	cache := lastfm.NewCache()
	lastfm := lastfm.NewServices(lastfmKey, cache)

	// bot state
	st := state.New("Bot " + token)
	st.AddIntents(gateway.IntentGuilds)
	st.AddIntents(gateway.IntentGuildMessages)

	// events
	reg := events.NewRegistry()
	events.RegisterDefaultEvents(reg)

	// commands
	r := cmdroute.NewRouter()
	commands.RegisterCommands(r, st, q, cache)
	if err := commands.Sync(st); err != nil {
		logging.Fatalf("failed syncing commands: %v", err)
	}
	st.AddInteractionHandler(r)

	return &Bot{
		State:    st,
		Query:    q,
		Database: db,
		LastFM:   lastfm,
		Cache:    cache,
		Registry: reg,
	}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	// automatize ts pls
	b.State.AddHandler(func(e *gateway.ReadyEvent) {
		b.Registry.Dispatch(events.TypeName(e), e)
	})

	if err := b.State.Open(ctx); err != nil {
		return err
	}
	defer b.State.Close()
	defer b.Database.Close()
	defer b.Cache.Close()

	logging.Info("bot is running")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-stop:
	}

	logging.Info("shutting down")
	return nil
}
