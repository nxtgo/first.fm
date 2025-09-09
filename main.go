package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	dgobot "github.com/disgoorg/disgo/bot"
	dgocache "github.com/disgoorg/disgo/cache"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/gateway"
	"github.com/nxtgo/env"

	"go.fm/commands"
	"go.fm/db"
	"go.fm/lastfm"
	"go.fm/logger"

	_ "embed"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed db/sql/schema.sql
var ddl string

func init() {
	if err := env.Load(""); err != nil {
		logger.Log.Fatalf("%v", err)
	}
}

func main() {
	dbCtx := context.Background()

	logger.Log.Debug("starting client...")
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logger.Log.Fatal("missing DISCORD_TOKEN variable")
	}

	lfm := lastfm.New()
	dbConn, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		logger.Log.Fatalf("failed opening database: %v", err)
	}
	defer dbConn.Close()

	if _, err := dbConn.ExecContext(dbCtx, ddl); err != nil {
		logger.Log.Fatalf("failed executing schema: %v", err)
	}

	if _, err := db.Prepare(dbCtx, dbConn); err != nil {
		logger.Log.Fatalf("failed preparing queries: %v", err)
	}

	database := db.New(dbConn)
	ctx := commands.CommandContext{
		LastFM:   lfm,
		Database: database,
	}
	commands.InitDependencies(ctx)

	cacheOptions := dgobot.WithCacheConfigOpts(
		dgocache.WithCaches(dgocache.FlagMembers),
	)
	options := dgobot.WithGatewayConfigOpts(
		gateway.WithIntents(
			gateway.IntentsNonPrivileged,
			gateway.IntentGuildMembers,
			gateway.IntentsGuild,
		),
	)

	client, err := disgo.New(
		token,
		options,
		dgobot.WithEventListeners(
			commands.Handler(),
		),
		cacheOptions,
	)
	if err != nil {
		logger.Log.Fatalf("failed to instantiate client: %v", err)
	}
	defer client.Close(context.TODO())

	if err = client.OpenGateway(context.TODO()); err != nil {
		logger.Log.Fatalf("failed to open gateway: %v", err)
	}

	if _, err = client.Rest().SetGlobalCommands(client.ApplicationID(), commands.All()); err != nil {
		logger.Log.Fatalf("failed registering commands: %v", err)
	}
	logger.Log.Info("registered slash commands")

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}
