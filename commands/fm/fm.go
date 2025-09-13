package fm

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	"go.fm/lfm"
	"go.fm/types/cmd"
	"go.fm/utils/image"
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

	data, err := ctx.LastFM.User.GetRecentTracks(lfm.P{"user": user, "limit": 1})
	if err != nil {
		_ = ctx.Error(e, constants.ErrorFetchCurrentTrack)
		return
	}

	if len(data.Tracks) == 0 {
		_ = ctx.Error(e, constants.ErrorNoTracks)
		return
	}

	track := data.Tracks[0]
	if track.NowPlaying != "true" {
		_ = ctx.Error(e, constants.ErrorNotPlaying)
		return
	}

	thumbnail := track.Images[len(track.Images)-1].Url
	trackData, _ := ctx.LastFM.Track.GetInfo(lfm.P{"user": user, "track": track.Name, "artist": track.Artist.Name})
	color, err := image.DominantColor(thumbnail)
	if err != nil {
		color = 0x00ADD8
	}

	component := discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplayf("## [%s](%s)\nby **%s**\n-# *At %s*", track.Name, track.Url, track.Artist.Name, track.Album.Name),
		).WithAccessory(discord.NewThumbnail(thumbnail)),
		discord.NewSmallSeparator(),
		discord.NewTextDisplayf("Scrobbled **%d** times", trackData.UserPlayCount),
	).WithAccentColor(color)

	_ = reply.Flags(discord.MessageFlagIsComponentsV2).Component(component).Edit()
}
