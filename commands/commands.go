package commands

import (
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nxtgo/zlog"

	"go.fm/commands/botinfo"
	"go.fm/commands/fm"
	"go.fm/commands/profile"
	profilev2 "go.fm/commands/profile/v2"
	"go.fm/commands/setuser"
	"go.fm/commands/top"
	"go.fm/commands/update"
	"go.fm/commands/whoknows"

	"go.fm/logger"
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

	// delete me
	Register(profilev2.Command{})
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
		logger.Log.Debugf("added command %s to registry", cmd.Data().CommandName())
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
				go cmd.Handle(e, sharedCtx)

				// logger stuff, may be removed later
				guildName := "user_app"
				guildId := "nil"
				guild, ok := e.Client().Caches.Guild(*e.GuildID())
				if ok {
					guildName = guild.Name
					guildId = guild.ID.String()
				}

				logger.Log.Debugw(
					"executed command %s",
					zlog.F{
						"guild_name":  guildName,
						"guild_id":    guildId,
						"author_name": e.Member().User.Username,
						"author_id":   e.Member().User.ID.String(),
					},
					cmd.Data().CommandName(),
				)
			}
		},
	}
}
