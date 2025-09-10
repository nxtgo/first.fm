package profile

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	"go.fm/util"
	"go.fm/util/opts"
	"go.fm/util/res"
	"go.fm/util/shared/cmd"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "profile",
		Description: "display a last.fm user info",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
		Options: []discord.ApplicationCommandOption{
			opts.UserOption,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := res.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = res.ErrorReply(e, constants.ErrorAcknowledgeCommand)
		return
	}

	username, _, err := opts.GetUser(e, ctx.Database)
	if err != nil {
		_ = res.ErrorReply(e, constants.ErrorNotRegistered)
		return
	}

	user, err := ctx.LastFM.GetUserInfo(username)
	if err != nil {
		_ = res.ErrorReply(e, constants.ErrorUserNotFound)
		return
	}

	realName := user.User.Realname
	if realName == "" {
		realName = user.User.Name
	}

	embed := res.QuickEmbed(
		fmt.Sprintf("%s's profile %s", realName, constants.EmojiWondering),
		fmt.Sprintf(
			"-# *@%s*\n\n"+
				"📀 **%s** scrobbles\n"+
				"🎤 **%s** artists\n"+
				"💿 **%s** albums\n"+
				"🎵 **%s** unique tracks",
			user.User.Name,
			user.User.Playcount,
			user.User.ArtistCount,
			user.User.AlbumCount,
			user.User.TrackCount,
		),
		0x00ADD8,
	)

	registeredTime := time.Unix(int64(user.User.Registered.Text), 0)

	embed.URL = user.User.URL
	embed.Thumbnail = &discord.EmbedResource{URL: user.User.Image[len(user.User.Image)-1].Text}
	embed.Footer = &discord.EmbedFooter{
		Text: fmt.Sprintf("account created %s", util.TimeAgo(registeredTime)),
	}
	// embed.Fields = append(embed.Fields, discord.EmbedField{Name: "favourite artist", Value: user.User.Fa})

	_ = res.Reply(e).Embed(embed).Edit()

	// if registered {
	// err = res.Reply(e).Ephemeral().Content("you are registered hehe").FollowUp()
	// return
	// }
}
