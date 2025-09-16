package fm

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/lfm"
	"go.fm/lfm/types"
	"go.fm/pkg/constants/errs"
	"go.fm/pkg/constants/opts"
	"go.fm/pkg/ctx"
	"go.fm/pkg/discord/reply"
	"go.fm/pkg/image"
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

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) {
	r := reply.New(e)
	if err := r.Defer(); err != nil {
		reply.Error(e, errs.ErrCommandDeferFailed)
		return
	}

	user, err := ctx.GetUser(e)
	if err != nil {
		reply.Error(e, errs.ErrUserNotFound)
		return
	}

	data, err := ctx.LastFM.User.GetRecentTracks(lfm.P{"user": user, "limit": 1})
	if err != nil {
		reply.Error(e, errs.ErrCurrentTrackFetch)
		return
	}

	if len(data.Tracks) == 0 {
		reply.Error(e, errs.ErrNoTracksFound)
		return
	}

	track := data.Tracks[0]
	if track.NowPlaying != "true" {
		reply.Error(e, errs.ErrNotListening)
		return
	}

	thumbnail := track.Images[len(track.Images)-1].Url
	trackData, err := ctx.LastFM.Track.GetInfo(lfm.P{
		"user":   user,
		"track":  track.Name,
		"artist": track.Artist.Name,
	})
	if err != nil || trackData == nil {
		trackData = &types.TrackGetInfo{UserPlayCount: 0}
	}

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

	r.Flags(discord.MessageFlagIsComponentsV2).Component(component).Edit()
}
