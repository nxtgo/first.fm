package stats

import (
	"first.fm/internal/bot"
	"github.com/disgoorg/disgo/discord"
)

func init() {
	bot.Register(data, handle)
}

var data = discord.SlashCommandCreate{
	Name:        "stats",
	Description: "display first.fm stats",
	IntegrationTypes: []discord.ApplicationIntegrationType{
		discord.ApplicationIntegrationTypeGuildInstall,
		discord.ApplicationIntegrationTypeUserInstall,
	},
}

func handle(ctx *bot.CommandContext) error {
	return ctx.Event.CreateMessage(discord.NewMessageCreateBuilder().SetContent("stats").Build())
}
