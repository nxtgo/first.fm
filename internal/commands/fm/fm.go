package fm

import (
	"errors"
	"time"

	"first.fm/internal/bot"
	"github.com/disgoorg/disgo/discord"
)

func init() {
	bot.Register(data, handle)
}

var data = discord.SlashCommandCreate{
	Name:        "fm",
	Description: "display an user's current track",
	IntegrationTypes: []discord.ApplicationIntegrationType{
		discord.ApplicationIntegrationTypeGuildInstall,
		discord.ApplicationIntegrationTypeUserInstall,
	},
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "user",
			Description: "user to get fm from",
			Required:    false,
		},
	},
}

func handle(ctx *bot.CommandContext) error {
	err := ctx.DeferCreateMessage(false)
	if err != nil {
		return err
	}

	user, err := ctx.GetLastFMUser("")
	if err != nil {
		return err
	}

	recentTrack, err := ctx.LastFM.User.RecentTrack(user.Name)
	if err != nil {
		return errors.New("failed to get recent track")
	}

	var text discord.TextDisplayComponent

	if recentTrack.Track.NowPlaying {
		text = discord.NewTextDisplayf("-# *Current track for **%s***", recentTrack.User)
	} else {
		text = discord.NewTextDisplayf("-# *Last track for **%s**, scrobbled at %s*", recentTrack.User, recentTrack.Track.ScrobbledAt.Format(time.Kitchen))
	}

	component := discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplayf("# %s", recentTrack.Track.Title),
			discord.NewTextDisplayf("**%s** **·** *%s*", recentTrack.Track.Artist.Name, recentTrack.Track.Album.Title),
			text,
		).WithAccessory(discord.NewThumbnail(recentTrack.Track.Image.OriginalURL())),
	)

	_, err = ctx.UpdateInteractionResponse(discord.NewMessageUpdateBuilder().
		SetIsComponentsV2(true).
		SetComponents(component).
		Build())
	return err
}
