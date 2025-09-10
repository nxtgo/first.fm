package setuser

import (
	"context"
	"database/sql"
	"errors"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/db"
	"go.fm/logger"
	"go.fm/util/res"
	"go.fm/util/shared/cmd"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "set-user",
		Description: "link your last.fm username to your Discord account",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "username",
				Description: "your last.fm username",
				Required:    true,
			},
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := res.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = res.ErrorReply(e, "failed to acknowledge command")
		return
	}

	username := e.SlashCommandInteractionData().String("username")
	discordID := e.User().ID.String()

	_, err := ctx.LastFM.GetUserInfo(username)
	if err != nil {
		_ = res.ErrorReply(e, "that username doesn't exist on last.fm")
		return
	}

	existing, err := ctx.Database.GetUserByUsername(context.Background(), username)
	if err == nil {
		if existing.DiscordID != discordID {
			_ = res.ErrorReply(e, "that last.fm username is already linked by another Discord user")
			return
		}
		if existing.LastfmUsername == username {
			_ = reply.Content("your last.fm username is already set to **%s**", username).Send()
			return
		}
	}

	if errors.Is(err, sql.ErrNoRows) || existing.DiscordID == discordID {
		if dbErr := ctx.Database.UpsertUser(context.Background(), db.UpsertUserParams{
			DiscordID:      discordID,
			LastfmUsername: username,
		}); dbErr != nil {
			logger.Log.Errorf("failed to upsert user %s: %v", discordID, dbErr)
			_ = res.ErrorReply(e, "failed to set your last.fm username")
			return
		}

		_ = reply.Content("your last.fm username has been set to **%s**", username).Send()
		return
	}

	if err != nil {
		logger.Log.Errorf("unexpected DB error in /set-user: %v", err)
		_ = res.ErrorReply(e, "an unexpected error occurred, please try again later")
	}
}
