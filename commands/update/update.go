package update

import (
	"fmt"
	"strings"

	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/discord"
	"go.fm/commands"
	lastfm "go.fm/last.fm"
	"go.fm/pkg/reply"
)

var data = api.CreateCommandData{
	Name:        "update",
	Description: "update your data or another user's",
	Options: discord.CommandOptions{
		discord.NewStringOption("data", "comma-separated list of data to update", true),
		discord.NewStringOption("user", "user to update data for", false),
	},
}

var options struct {
	User *string `discord:"user"`
	Data string  `discord:"data"`
}

// todo: let the user know which data is available to update
func handler(c *commands.CommandContext) error {
	return c.Reply.AutoDefer(func(edit *reply.EditBuilder) error {
		if err := c.Data.Options.Unmarshal(&options); err != nil {
			return err
		}

		username, err := c.GetUserOrFallback()
		if err != nil {
			return err
		}

		dataTypes := strings.Split(options.Data, ",")
		for i := range dataTypes {
			dataTypes[i] = strings.TrimSpace(strings.ToLower(dataTypes[i]))
		}

		userParams := lastfm.P{"user": username}

		for _, t := range dataTypes {
			switch t {
			case "profile", "getinfo":
				go c.Last.User.GetInfo(userParams)
			case "topalbums":
				go c.Last.User.GetTopAlbums(userParams)
			case "topartists":
				go c.Last.User.GetTopArtists(userParams)
			case "toptracks":
				go c.Last.User.GetTopTracks(userParams)
			default:
				return fmt.Errorf("unknown data type: %s", t)
			}
		}

		_, err = edit.Content(
			fmt.Sprintf(
				"updated the following data for `%s`:\n\\- %s",
				username,
				strings.Join(dataTypes, "\n\\- "),
			),
		).Send()
		return err
	})
}

func init() {
	commands.Register(data, handler)
}
