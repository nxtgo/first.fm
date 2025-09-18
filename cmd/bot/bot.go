package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	dgocache "github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nxtgo/env"

	"go.fm/cache"
	"go.fm/commands"
	"go.fm/db"
	"go.fm/lfm"
	"go.fm/logger"
	gofmCtx "go.fm/pkg/ctx"
)

var (
	globalCmds     bool
	deleteCommands bool
	dbPath         string
)

func init() {
	debug.SetMemoryLimit(1 << 30)
	if err := env.Load(""); err != nil {
		logger.Log.Fatalf("failed loading environment: %v", err)
	}

	flag.BoolVar(&globalCmds, "global", true, "upload global commands to discord")
	flag.BoolVar(&deleteCommands, "delete", false, "delete commands on exit")
	flag.StringVar(&dbPath, "db", "database.db", "path to the SQLite database file")
	flag.Parse()
}

func main() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			if m.Alloc > 500*1024*1024 {
				logger.Log.Warnf("high memory usage: %d MB", m.Alloc/1024/1024)
				runtime.GC()
			}
		}
	}()

	ctx := context.Background()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logger.Log.Fatal("missing DISCORD_TOKEN environment variable")
	}

	lfmCache := cache.NewCache()
	lfm := lfm.New(os.Getenv("LASTFM_API_KEY"), lfmCache)
	defer lfmCache.Close()

	closeConnection, database := initDatabase(ctx, dbPath)
	defer closeConnection()

	cmdCtx := gofmCtx.CommandContext{
		LastFM:   lfm,
		Database: database,
		Cache:    lfmCache,
		Context:  ctx,
	}
	commands.InitDependencies(cmdCtx)

	client := initDiscordClient(token)
	defer client.Close(context.TODO())

	if err := client.OpenGateway(context.TODO()); err != nil {
		logger.Log.Fatalf("failed to open gateway: %v", err)
	}

	if globalCmds {
		uploadGlobalCommands(*client)
		if deleteCommands {
			defer deleteAllGlobalCommands(*client)
		}
	} else {
		uploadGuildCommands(*client)
		if deleteCommands {
			defer deleteAllGuildCommands(*client)
		}
	}

	waitForShutdown()
}

func initDatabase(ctx context.Context, path string) (func() error, *db.Queries) {
	dbConn, err := sql.Open("sqlite3", path)
	if err != nil {
		logger.Log.Fatalf("failed opening database: %v", err)
	}

	if _, err := dbConn.ExecContext(ctx, db.Schema); err != nil {
		dbConn.Close()
		logger.Log.Fatalf("failed executing schema: %v", err)
	}

	queries, err := db.Prepare(ctx, dbConn)
	if err != nil {
		dbConn.Close()
		logger.Log.Fatalf("failed preparing queries: %v", err)
	}

	return func() error {
		queries.Close()
		return dbConn.Close()
	}, queries
}

func initDiscordClient(token string) *bot.Client {
	cacheOptions := bot.WithCacheConfigOpts(
		dgocache.WithCaches(dgocache.FlagMembers, dgocache.FlagGuilds),
	)

	options := bot.WithGatewayConfigOpts(
		gateway.WithIntents(
			gateway.IntentsNonPrivileged,
			gateway.IntentGuildMembers,
			gateway.IntentsGuild,
		),
	)

	client, err := disgo.New(
		token,
		options,
		bot.WithEventListeners(
			commands.Handler(),
			bot.NewListenerFunc(func(r *events.Ready) {
				logger.Log.Info("client ready v/")
				if err := r.Client().SetPresence(context.TODO(),
					//gateway.WithListeningActivity("your scrobbles <3"),
					gateway.WithCustomActivity("xd"),
					gateway.WithOnlineStatus(discord.OnlineStatusOnline),
				); err != nil {
					logger.Log.Errorf("failed to set presence: %s", err)
				}
			}),
		),
		cacheOptions,
	)
	if err != nil {
		logger.Log.Fatalf("failed to instantiate Discord client: %v", err)
	}

	return client
}

func uploadGlobalCommands(client bot.Client) {
	_, err := client.Rest.SetGlobalCommands(client.ApplicationID, commands.All())
	if err != nil {
		logger.Log.Fatalf("failed registering global commands: %v", err)
	}
	logger.Log.Info("registered global slash commands")
}

func uploadGuildCommands(client bot.Client) {
	guildId := snowflake.GetEnv("GUILD_ID")
	_, err := client.Rest.SetGuildCommands(client.ApplicationID, guildId, commands.All())
	if err != nil {
		logger.Log.Fatalf("failed registering global commands: %v", err)
	}
	logger.Log.Infof("registered guild slash commands to guild '%s'", guildId.String())
}

func deleteAllGlobalCommands(client bot.Client) {
	commands, err := client.Rest.GetGlobalCommands(client.ApplicationID, false)
	if err != nil {
		logger.Log.Fatalf("failed fetching global commands: %v", err)
	}

	for _, cmd := range commands {
		if err := client.Rest.DeleteGlobalCommand(client.ApplicationID, cmd.ID()); err != nil {
			logger.Log.Errorf("failed deleting global command '%s': %v", cmd.Name(), err)
		} else {
			logger.Log.Infof("deleted global command '%s'", cmd.Name())
		}
	}
}

func deleteAllGuildCommands(client bot.Client) {
	guildId := snowflake.GetEnv("GUILD_ID")

	commands, err := client.Rest.GetGuildCommands(client.ApplicationID, guildId, false)
	if err != nil {
		logger.Log.Fatalf("failed fetching guild commands: %v", err)
	}

	for _, cmd := range commands {
		if err := client.Rest.DeleteGuildCommand(client.ApplicationID, guildId, cmd.ID()); err != nil {
			logger.Log.Errorf("failed deleting guild command '%s': %v", cmd.Name(), err)
		} else {
			logger.Log.Infof("deleted guild command '%s'", cmd.Name())
		}
	}
}

func waitForShutdown() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	logger.Log.Info("goodbye :)")
}
