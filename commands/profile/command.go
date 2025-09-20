package profile

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"go.fm/commands"
	lastfm "go.fm/last.fm"
	"go.fm/pkg/reply"
)

var data = api.CreateCommandData{
	Name:        "profile",
	Description: "display your last.fm profile or another user's",
	Options: discord.CommandOptions{
		discord.NewStringOption("user", "user to display profile from", false),
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

		user, err := c.Last.User.GetInfo(lastfm.P{"user": username})
		if err != nil {
			return err
		}

		_, err = edit.Content(user.Name).Send()
		return err
	})
}

func init() {
	commands.Register(data, handler)
}
