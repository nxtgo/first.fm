package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go.fm/util/opts"
	"go.fm/util/res"
)

type FmCommand struct{}

func (FmCommand) Data() discord.ApplicationCommandCreate {
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

func (FmCommand) Handle(e *events.ApplicationCommandInteractionCreate, ctx CommandContext) {
	reply := res.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = res.ErrorReply(e, "failed to acknowledge command")
		return
	}

	user, err := opts.GetUser(e, ctx.Database)
	if err != nil {
		_ = res.ErrorReply(e, err.Error())
		return
	}

	data, err := ctx.LastFM.GetRecentTracks(user, 1)
	if err != nil {
		_ = res.ErrorReply(e, "failed to fetch Last.fm data")
		return
	}

	if len(data.RecentTracks.Track) == 0 {
		_ = res.ErrorReply(e, "no tracks found for that user")
		return
	}

	track := data.RecentTracks.Track[0]
	if track.Attr.Nowplaying != "true" {
		_ = res.ErrorReply(e, "the user isn't listening to anything right now")
		return
	}

	// build embed
	embed := res.QuickEmbed(
		track.Name,
		fmt.Sprintf("by **%s**\n-# *at %s*", track.Artist.Text, track.Album.Text),
		0x00ADD8,
	)
	embed.Author = &discord.EmbedAuthor{
		Name: fmt.Sprintf("%s's current track", user),
		URL:  fmt.Sprintf("https://www.last.fm/user/%s", user),
	}
	embed.URL = track.URL
	if len(track.Image) > 0 {
		embed.Thumbnail = &discord.EmbedResource{URL: track.Image[len(track.Image)-1].Text}
	}

	// edit deferred reply with embed (public)
	_ = reply.Embed(embed).Send()
}

func init() {
	Register(FmCommand{})
}
