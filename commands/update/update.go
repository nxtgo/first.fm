package update

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"go.fm/lfm"
	"go.fm/pkg/constants/errs"
	"go.fm/pkg/constants/opts"
	"go.fm/pkg/ctx"
	"go.fm/pkg/discord/reply"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "update",
		Description: "refresh cache data",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
		Options: []discord.ApplicationCommandOption{
			opts.UserOption,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) {
	r := reply.New(e)
	if err := r.Defer(); err != nil {
		reply.Error(e, errs.ErrCommandDeferFailed)
		return
	}

	username, err := ctx.GetUser(e)
	if err != nil {
		reply.Error(e, errs.ErrUserNotFound)
		return
	}

	if ctx.Cache != nil {
		if ctx.Cache.User != nil {
			ctx.Cache.User.Delete(username)
		}
		if ctx.Cache.TopArtists != nil {
			ctx.Cache.TopArtists.Delete(username)
		}
		if ctx.Cache.TopAlbums != nil {
			ctx.Cache.TopAlbums.Delete(username)
		}
		if ctx.Cache.TopTracks != nil {
			ctx.Cache.TopTracks.Delete(username)
		}
		if e.GuildID() != nil && ctx.Cache.Members != nil {
			members, _ := ctx.Cache.Members.Get(*e.GuildID())
			if members != nil {
				delete(members, e.User().ID)
				ctx.Cache.Members.Set(*e.GuildID(), members, 0)
			}
		}
	}

	userInfo, err := ctx.LastFM.User.GetInfo(lfm.P{
		"user": username,
	})
	if err != nil {
		reply.Error(e, errs.ErrUserNotFound)
		return
	}
	ctx.Cache.User.Set(username, *userInfo, 0)

	if e.GuildID() != nil && ctx.Cache.Members != nil {
		members, _ := ctx.Cache.Members.Get(*e.GuildID())
		if members == nil {
			members = make(map[snowflake.ID]string)
		}
		members[e.User().ID] = username
		ctx.Cache.Members.Set(*e.GuildID(), members, 0)
	}

	go ctx.LastFM.User.PrefetchUserData(username)

	r.Content("updated your data with fresh one :)").Edit()
}
