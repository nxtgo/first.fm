package commands

import (
	"context"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/commands/botinfo"
	"go.fm/commands/fm"
	"go.fm/commands/profile"
	"go.fm/commands/set-user"
	"go.fm/commands/who-knows"
	"go.fm/util/shared/cmd"
)

var sharedCtx cmd.CommandContext
var registry = map[string]Command{}

func init() {
	Register(fm.Command{})
	Register(whoknows.Command{})
	Register(setuser.Command{})
	Register(profile.Command{})

	// non-lastfm commands :prayge:
	Register(botinfo.Command{})
}

type Command interface {
	Data() discord.ApplicationCommandCreate
	Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext)
}

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

func InitDependencies(ctx cmd.CommandContext) {
	ctx.Start = time.Now()
	ctx.Context = context.Background()
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
