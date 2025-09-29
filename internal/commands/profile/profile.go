package profile

import (
	"first.fm/internal/bot"
	"github.com/disgoorg/disgo/discord"
)

func init() {
	bot.Register(data, handle)
}

var data = discord.SlashCommandCreate{
	Name:        "profile",
	Description: "display someone's profile",
	IntegrationTypes: []discord.ApplicationIntegrationType{
		discord.ApplicationIntegrationTypeGuildInstall,
		discord.ApplicationIntegrationTypeUserInstall,
	},
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "user",
			Description: "user to get profile from",
			Required:    false,
		},
	},
}

func handle(ctx *bot.CommandContext) error {
	err := ctx.DeferCreateMessage(false)
	if err != nil {
		return err
	}

	user, err := ctx.GetLastFMUser("")
	if err != nil {
		return err
	}

	component := discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplayf("# [%s](%s)", user.Name, user.URL),
		).WithAccessory(discord.NewThumbnail(user.Avatar.OriginalURL())),
	).WithAccentColor(0x00ADD8)

	_, err = ctx.UpdateInteractionResponse(discord.NewMessageUpdateBuilder().
		SetIsComponentsV2(true).
		SetComponents(component).
		Build())
	return err
}
