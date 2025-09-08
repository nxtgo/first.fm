package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go.fm/util"
)

type FmCommand struct{}

func (FmCommand) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "fm",
		Description: "get an user's current track",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "user",
				Description: "user to get track from",
				Required:    true,
			},
		},
	}
}

func (FmCommand) Handle(e *events.ApplicationCommandInteractionCreate, ctx CommandContext) {
	user, _ := e.SlashCommandInteractionData().OptString("user")

	data, err := ctx.LastFM.GetRecentTracks(user, 1)
	if err != nil {
		_ = util.Reply(e).Content("failed to fetch last.fm data").Send()
		return
	}

	track := data.Recenttracks.Track[0]
	if track.Attr.Nowplaying != "true" {
		_ = util.Reply(e).Content("the provided user isn't listening to anything right now").Send()
		return
	}

	embed := util.QuickEmbed(
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

	_ = util.Reply(e).Embed(embed).Send()
}

func init() {
	Register(FmCommand{})
}
