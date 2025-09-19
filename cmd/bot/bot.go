package main

import (
	"context"
	"os"

	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"

	"go.fm/commands"
	"go.fm/events"
	"go.fm/zlog"
)

func main() {
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		zlog.Log.Fatal("missing DISCORD_TOKEN env")
	}

	s := state.New("Bot " + discordToken)
	r := cmdroute.NewRouter()
	commands.RegisterCommands(r, s)

	// command handlers
	if err := commands.Sync(s); err != nil {
		zlog.Log.Fatalf("failed syncing commands: %v", err)
	}
	s.AddInteractionHandler(r)

	// event handlers
	for _, event := range events.Events {
		s.AddHandler(event)
	}
	zlog.Log.Debugf("added %d event handlers", len(events.Events))

	// bot intents
	s.AddIntents(gateway.IntentGuildMembers)

	// open gateway
	if err := s.Open(context.Background()); err != nil {
		zlog.Log.Fatalf("failed to open gateway: %v", err)
	}
	defer s.Close()

	// set status
	err := s.Gateway().Send(
		context.Background(),
		&gateway.UpdatePresenceCommand{
			Since:  discord.UnixMsTimestamp(0),
			Status: discord.OnlineStatus,
			Activities: []discord.Activity{
				{
					Name: "your breath",
					Type: discord.ListeningActivity,
				},
			},
			AFK: false,
		},
	)
	if err != nil {
		zlog.Log.Warnf("failed to set status: %v", err)
	}

	select {}
}
