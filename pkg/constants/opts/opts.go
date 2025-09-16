package opts

import "github.com/disgoorg/disgo/discord"

var (
	UserOption = discord.ApplicationCommandOptionString{
		Name:        "user",
		Description: "user to get data from",
		Required:    false,
	}
)
