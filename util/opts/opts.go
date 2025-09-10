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

func GetUser(e *events.ApplicationCommandInteractionCreate, q *db.Queries) (string, bool, error) {
	user, defined := e.SlashCommandInteractionData().OptString("user")
	if !defined {
		userId := e.Member().User.ID.String()
		username, err := q.GetUser(context.Background(), userId)

		return username, true, err
	}

	return user, false, nil
}
