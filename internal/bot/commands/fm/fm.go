package fm

import (
	"time"

	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/discord"
	"go.fm/internal/bot/commands"
	"go.fm/internal/bot/discord/components"
	"go.fm/internal/bot/discord/reply"
	"go.fm/internal/bot/image/colors"
	"go.fm/internal/bot/lastfm"
)

var data = api.CreateCommandData{
	Name:        "fm",
	Description: "display your current track or another user's",
	Options: discord.CommandOptions{
		discord.NewStringOption("user", "user to display track from", false),
	},
}

var options struct {
	User *string `discord:"user"`
}

func handler(c *commands.CommandContext) error {
	return c.Reply.AutoDefer(func(edit *reply.EditBuilder) error {
		if err := c.Data.Options.Unmarshal(&options); err != nil {
			return err
		}

		username, err := c.GetUserOrFallback()
		if err != nil {
			return err
		}

		res, err := c.Last.User.GetRecentTracks(lastfm.P{"user": username, "limit": 1})
		if err != nil {
			return err
		}

		var text *components.TextDisplay

		lastTrack := res.Tracks[0]
		if lastTrack.NowPlaying == "true" {
			text = components.NewTextDisplayf("-# *Current track for %s*", res.User)
		} else {
			playtime, err := lastTrack.GetPlayTime()
			if err != nil {
				playtime = time.Now()
			}

			text = components.NewTextDisplayf("-# *Last track for %s, scrobbled at %s*", res.User, playtime.Format(time.Kitchen))
		}

		thumbnail := lastTrack.GetLargestImage().URL
		color := 0x703487
		if dominantColor, err := colors.Dominant(thumbnail); err == nil {
			color = dominantColor
		}

		container := components.NewContainer(color,
			components.NewSection(
				components.NewTextDisplayf("# %s", lastTrack.Name),
				components.NewTextDisplayf("**%s** **·** %s", lastTrack.Artist.Name, lastTrack.Album.Name),
				text,
			).WithAccessory(components.NewThumbnail(thumbnail)),
			components.NewActionRow(
				components.NewButton(components.ButtonStyleLink, "Last.fm", nil).WithEmoji("1418269025959546943").WithURL(lastTrack.URL),
			),
		)

		_, err = edit.ComponentsV2(container).Send()
		return err
	})
}

func init() {
	commands.Register(data, handler)
}
