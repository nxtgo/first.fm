package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type PingCommand struct{}

func (PingCommand) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "ping",
		Description: "get bot's latency",
	}
}

func (PingCommand) Handle(e *events.ApplicationCommandInteractionCreate, _ CommandContext) {
	_ = e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent("pong").
		Build())
}

func init() {
	Register(PingCommand{})
}
