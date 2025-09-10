package fm

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	"go.fm/types/cmd"
	"go.fm/util/opts"
	"go.fm/util/res"
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
			opts.UserOption,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := res.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = res.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	user, _, err := opts.GetUser(e, ctx.Database)
	if err != nil {
		_ = res.Error(e, err.Error())
		return
	}

	data, err := ctx.LastFM.GetRecentTracks(user, 1)
	if err != nil {
		_ = res.Error(e, constants.ErrorFetchCurrentTrack)
		return
	}

	if len(data.RecentTracks.Track) == 0 {
		_ = res.Error(e, constants.ErrorNoTracks)
		return
	}

	track := data.RecentTracks.Track[0]
	if track.Attr.Nowplaying != "true" {
		_ = res.Error(e, constants.ErrorNotPlaying)
		return
	}

	embed := res.QuickEmbed(
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
