package commands

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go.fm/bot/cache"
	"go.fm/db"
	"go.fm/lastfm"
)

type CommandContext struct {
	LastFM   *lastfm.Client
	Cache    *cache.CustomCaches
	Database *db.Queries
}

type Command interface {
	Data() discord.ApplicationCommandCreate
	Handle(e *events.ApplicationCommandInteractionCreate, ctx CommandContext)
}

var registry = map[string]Command{}

func Register(cmd Command) {
	registry[cmd.Data().CommandName()] = cmd
}

func All() []discord.ApplicationCommandCreate {
	cmds := make([]discord.ApplicationCommandCreate, 0, len(registry))
	for _, cmd := range registry {
		cmds = append(cmds, cmd.Data())
	}

	return cmds
}

var sharedCtx CommandContext

func InitDependencies(ctx CommandContext) {
	sharedCtx = ctx
}

func Handler() bot.EventListener {
	return &events.ListenerAdapter{
		OnApplicationCommandInteraction: func(e *events.ApplicationCommandInteractionCreate) {
			if cmd, ok := registry[e.Data.CommandName()]; ok {
				cmd.Handle(e, sharedCtx)
			}
		},
	}
}
