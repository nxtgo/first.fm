package update

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"go.fm/constants"
	"go.fm/lfm"
	"go.fm/types/cmd"
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
			cmd.UserOption,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := ctx.Reply(e)
	if err := reply.Defer(); err != nil {
		_ = ctx.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	username, err := ctx.GetUser(e)
	if err != nil {
		_ = reply.Content(constants.ErrorNotRegistered).Edit()
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
		_ = reply.Content(constants.ErrorGetUser).Edit()
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

	if ctx.Cache != nil {
		if ctx.Cache.TopArtists != nil {
			if top, err := ctx.LastFM.User.GetTopArtists(lfm.P{"user": username}); err == nil {
				ctx.Cache.TopArtists.Set(username, *top, 0)
			}
		}
		if ctx.Cache.TopAlbums != nil {
			if top, err := ctx.LastFM.User.GetTopAlbums(lfm.P{"user": username}); err == nil {
				ctx.Cache.TopAlbums.Set(username, *top, 0)
			}
		}
		if ctx.Cache.TopTracks != nil {
			if top, err := ctx.LastFM.User.GetTopTracks(lfm.P{"user": username}); err == nil {
				ctx.Cache.TopTracks.Set(username, *top, 0)
			}
		}
	}

	_ = reply.Content("updated your data with fresh one :)").Edit()
}
