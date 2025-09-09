package commands

import (
	"context"
	"database/sql"
	"errors"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/db"
	"go.fm/logger"
	"go.fm/util/res"
)

type SetUserCommand struct{}

func (SetUserCommand) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "set-user",
		Description: "link your Last.fm username to your Discord account",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "username",
				Description: "your Last.fm username",
				Required:    true,
			},
		},
	}
}

func (SetUserCommand) Handle(e *events.ApplicationCommandInteractionCreate, ctx CommandContext) {
	reply := res.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = res.ErrorReply(e, "failed to acknowledge command")
		return
	}

	username := e.SlashCommandInteractionData().String("username")
	discordID := e.User().ID.String()

	userInfo, err := ctx.LastFM.GetUserInfo(username)
	if err != nil || userInfo.User.Name == "" {
		_ = res.ErrorReply(e, "that username doesn't exist on Last.fm")
		return
	}

	existing, err := ctx.Database.GetUserByUsername(context.Background(), username)
	if err == nil {
		if existing.DiscordID != discordID {
			_ = res.ErrorReply(e, "that Last.fm username is already linked by another Discord user")
			return
		}
		if existing.LastfmUsername == username {
			_ = reply.Content("your Last.fm username is already set to **%s**", username).Send()
			return
		}
	}

	if errors.Is(err, sql.ErrNoRows) || existing.DiscordID == discordID {
		if dbErr := ctx.Database.UpsertUser(context.Background(), db.UpsertUserParams{
			DiscordID:      discordID,
			LastfmUsername: username,
		}); dbErr != nil {
			logger.Log.Errorf("failed to upsert user %s: %v", discordID, dbErr)
			_ = res.ErrorReply(e, "failed to set your Last.fm username")
			return
		}

		_ = reply.Content("your Last.fm username has been set to **%s**", username).Send()
		return
	}

	if err != nil {
		logger.Log.Errorf("unexpected DB error in /set-user: %v", err)
		_ = res.ErrorReply(e, "an unexpected error occurred, please try again later")
	}
}

func init() {
	Register(SetUserCommand{})
}
