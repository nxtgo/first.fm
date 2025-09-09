package opts

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go.fm/db"
)

var UserOption = discord.ApplicationCommandOptionString{
	Name:        "user",
	Description: "user to get data from",
	Required:    false,
}

func GetUser(e *events.ApplicationCommandInteractionCreate, q *db.Queries) (string, error) {
	user, defined := e.SlashCommandInteractionData().OptString("user")
	if !defined {
		username, err := q.GetUserByDiscordID(context.Background(), e.Member().User.ID.String())

		return username.Username, err
	}

	return user, nil
}
