package profile

import (
	"first.fm/internal/bot"
	"first.fm/internal/emojis"
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

	component := []discord.LayoutComponent{
		discord.NewContainer(
			discord.NewSection(
				discord.NewTextDisplayf("## [%s](%s)", user.Name, user.URL),
				discord.NewTextDisplayf("Since <t:%d:D> %s", user.RegisteredAt.Time().Unix(), emojis.EmojiCalendar),
				discord.NewTextDisplayf("**%d** total scrobbles %s", user.Playcount, emojis.EmojiPlay),
			).WithAccessory(discord.NewThumbnail(user.Avatar.OriginalURL())),
			discord.NewSmallSeparator(),
			discord.NewTextDisplayf(
				"%s **%d** albums\n%s **%d** artists\n%s **%d** unique tracks",
				emojis.EmojiAlbum,
				user.ArtistCount,
				emojis.EmojiMic2,
				user.AlbumCount,
				emojis.EmojiNote,
				user.TrackCount,
			),
		).WithAccentColor(0x00ADD8),
		discord.NewActionRow(
			discord.NewLinkButton("Last.fm", user.URL).WithEmoji(discord.NewCustomComponentEmoji(emojis.EmojiLastFMRed.Snowflake())),
		),
	}

	_, err = ctx.UpdateInteractionResponse(discord.NewMessageUpdateBuilder().
		SetIsComponentsV2(true).
		SetComponents(component...).
		Build())
	return err
}
