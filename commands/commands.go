package commands

import (
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/commands/botinfo"
	"go.fm/commands/fm"
	"go.fm/commands/profile"
	"go.fm/commands/set-user"
	"go.fm/commands/top"
	"go.fm/commands/update"
	"go.fm/commands/who-knows"
	"go.fm/pkg/ctx"
)

var sharedCtx ctx.CommandContext
var registry = map[string]Command{}

func init() {
	Register(fm.Command{})
	Register(whoknows.Command{})
	Register(setuser.Command{})
	Register(profile.Command{})
	Register(top.Command{})
	Register(update.Command{})

	// non-lastfm commands :prayge:
	Register(botinfo.Command{})
}

type Command interface {
	Data() discord.ApplicationCommandCreate
	Handle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext)
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

func InitDependencies(ctx ctx.CommandContext) {
	ctx.Start = time.Now()
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
