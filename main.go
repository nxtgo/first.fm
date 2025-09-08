package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	dgocache "github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nxtgo/env"
	"go.fm/cache"
	"go.fm/commands"
	"go.fm/lastfm"
	"go.fm/logger"
)

func init() {
	err := env.Load("")
	if err != nil {
		logger.Log.Fatalf("%v", err)
	}
}

var sharedCache *cache.CustomCaches

func main() {
	logger.Log.Debug("starting client...")
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logger.Log.Fatal("missing DISCORD_TOKEN variable")
	}

	sharedCache = cache.New()
	lfm := lastfm.New()
	ctx := commands.CommandContext{
		LastFM: lfm,
		Cache:  sharedCache,
	}
	commands.InitDependencies(ctx)

	cacheOptions := bot.WithCacheConfigOpts(
		dgocache.WithCaches(dgocache.FlagsNone),
	)
	options := bot.WithGatewayConfigOpts(
		gateway.WithIntents(
			gateway.IntentsNonPrivileged,
		),
	)
	listeners := bot.WithEventListeners(
		commands.Handler(),
		&events.ListenerAdapter{
			OnGuildReady: GuildReadyListener,
		},
	)
	client, err := disgo.New(token, options, listeners, cacheOptions)
	if err != nil {
		logger.Log.Fatalf("failed to instantiate client: %v", err)
	}
	defer client.Close(context.TODO())

	if err = client.OpenGateway(context.TODO()); err != nil {
		logger.Log.Fatalf("failed to open gateway: %v", err)
	}

	guildId := snowflake.GetEnv("GUILD_ID")
	_, err = client.Rest().
		SetGuildCommands(
			client.ApplicationID(),
			guildId,
			commands.All(),
		)
	if err != nil {
		logger.Log.Fatalf("failed registering commands: %v", err)
	}
	logger.Log.Info("registered slash commands")

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}
