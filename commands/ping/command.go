package ping

import (
	"go.fm/commands"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

var data = api.CreateCommandData{
	Name:        "ping",
	Description: "display bot's latency",
}

func handler(c *commands.CommandContext) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("pong."),
	}
}

func init() {
	commands.Register(data, handler)
}
