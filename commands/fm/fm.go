package fm

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	"go.fm/types/cmd"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "fm",
		Description: "get an user's current track",
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

	user, err := ctx.GetUser(e)
	if err != nil {
		_ = ctx.Error(e, err.Error())
		return
	}

	data, err := ctx.LastFM.GetRecentTracks(user, 1)
	if err != nil {
		_ = ctx.Error(e, constants.ErrorFetchCurrentTrack)
		return
	}

	if len(data.RecentTracks.Track) == 0 {
		_ = ctx.Error(e, constants.ErrorNoTracks)
		return
	}

	track := data.RecentTracks.Track[0]
	if track.Attr.Nowplaying != "true" {
		_ = ctx.Error(e, constants.ErrorNotPlaying)
		return
	}

	embed := ctx.QuickEmbed(
		track.Name,
		fmt.Sprintf("by **%s**\n-# *at %s*", track.Artist.Text, track.Album.Text),
	)
	embed.Author = &discord.EmbedAuthor{
		Name: fmt.Sprintf("%s's current track", user),
		URL:  fmt.Sprintf("https://www.last.fm/user/%s", user),
	}
	embed.URL = track.URL
	if len(track.Image) > 0 {
		embed.Thumbnail = &discord.EmbedResource{
			URL: track.Image[len(track.Image)-1].Text,
		}
	}

	_ = reply.Embed(embed).Edit()
}
