package commands

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nxtgo/zlog"
	"go.fm/cache"
	"go.fm/lastfm"
	"go.fm/logger"
	"go.fm/util"
)

type CommandContext struct {
	LastFM *lastfm.Client
	Cache  *cache.CustomCaches
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
				guild, err := sharedCtx.Cache.Guild(*e.GuildID())
				if !err {
					_ = util.Reply(e).Content("you can't use me outside guilds :(").Send()
					return
				}

				logger.Log.Debugw(
					"ran command %s in guild %s",
					zlog.F{
						"gid": guild.ID.String(),
						"uid": e.Member().User.ID.String(),
					},
					e.Data.CommandName(),
					guild.Name,
				)

				cmd.Handle(e, sharedCtx)
			}
		},
	}
}
