package register

import (
	"errors"
	"strings"

	"first.fm/internal/bot"
	"first.fm/internal/persistence/sqlc"
	"github.com/disgoorg/disgo/discord"
)

func init() {
	bot.Register(data, handle)
}

var data = discord.SlashCommandCreate{
	Name:        "register",
	Description: "link your last.fm username",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "username",
			Description: "your last.fm username",
			Required:    true,
		},
	},
}

func handle(ctx *bot.CommandContext) error {
	username := ctx.SlashCommandInteractionData().String("username")

	_, err := ctx.LastFM.User.Info(username)
	if err != nil {
		return errors.New("last.fm user not found")
	}

	err = ctx.Queries.UpsertUser(ctx.Ctx, sqlc.UpsertUserParams{
		UserID:         ctx.User().ID,
		LastfmUsername: username,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("another discord user already uses this username")
		}

		return err
	}

	return ctx.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("successfully linked your account to **%s**", username).
		Build())
}
