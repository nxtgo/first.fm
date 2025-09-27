package update

import (
	"fmt"
	"slices"
	"strings"

	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/discord"
	"go.fm/internal/bot/commands"
	"go.fm/internal/bot/discord/reply"
	"go.fm/internal/bot/lastfm"
)

var validDataTypes = []string{
	"profile", "getinfo",
	"topalbums",
	"topartists",
	"toptracks",
	"all",
}

var data = api.CreateCommandData{
	Name:        "update",
	Description: "Update your Last.fm data or another user's",
	Options: discord.CommandOptions{
		discord.NewStringOption("data",
			"comma-separated list of data to update",
			true,
		),
		discord.NewStringOption("user", "user to update data for", false),
	},
}

var options struct {
	User *string `discord:"user"`
	Data string  `discord:"data"`
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

		dataTypes := strings.Split(options.Data, ",")
		for i := range dataTypes {
			dataTypes[i] = strings.TrimSpace(strings.ToLower(dataTypes[i]))
		}

		for _, t := range dataTypes {
			if !slices.Contains(validDataTypes, t) {
				return fmt.Errorf("`%s` is not a valid data type. valid options: %s", t, strings.Join(validDataTypes, ", "))
			}
		}

		userParams := lastfm.P{"user": username}
		updated := make([]string, 0, len(dataTypes))

		for _, t := range dataTypes {
			switch t {
			case "profile", "getinfo":
				update(c, "user.getinfo", c.Last.User.GetInfo, userParams)
				updated = append(updated, "profile")

			case "topalbums":
				update(c, "user.gettopalbums", c.Last.User.GetTopAlbums, userParams)
				updated = append(updated, "topalbums")

			case "topartists":
				update(c, "user.gettopartists", c.Last.User.GetTopArtists, userParams)
				updated = append(updated, "topartists")

			case "toptracks":
				update(c, "user.gettoptracks", c.Last.User.GetTopTracks, userParams)
				updated = append(updated, "toptracks")

			case "all":
				update(c, "user.getinfo", c.Last.User.GetInfo, userParams)
				update(c, "user.gettopalbums", c.Last.User.GetTopAlbums, userParams)
				update(c, "user.gettopartists", c.Last.User.GetTopArtists, userParams)
				update(c, "user.gettoptracks", c.Last.User.GetTopTracks, userParams)
				updated = append(updated, "all")
			}
		}

		embed := reply.SuccessEmbed(fmt.Sprintf("updated the following data for `%s`:\n\\- %s",
			username,
			strings.Join(updated, "\n\\- ")))
		_, err = edit.Embed(embed).Send()
		return err
	})
}

func update[T any](c *commands.CommandContext, key string, fn func(lastfm.P) (*T, error), params lastfm.P) {
	c.Cache.User.Delete(lastfm.GenerateCacheKey(key, params))
	go fn(params)
}

func init() {
	commands.Register(data, handler)
}
