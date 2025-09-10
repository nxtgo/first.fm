package setuser

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nxtgo/zlog"

	"go.fm/constants"
	"go.fm/db"
	"go.fm/logger"
	"go.fm/types/cmd"
	"go.fm/util/res"
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
		_ = res.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	username := e.SlashCommandInteractionData().String("username")
	discordID := e.User().ID.String()

	_, err := ctx.LastFM.GetUserInfo(username)
	if err != nil {
		_ = res.Error(e, constants.ErrorUserNotFound)
		return
	}

	existing, err := ctx.Database.GetUserByUsername(ctx.Context, username)
	if err == nil {
		if existing.DiscordID != discordID {
			_ = res.Error(e, constants.ErrorAlreadyLinked)
			return
		}
		if existing.LastfmUsername == username {
			_ = res.Error(e, fmt.Sprintf(constants.ErrorUsernameAlreadySet, username))
			return
		}
	}

	if errors.Is(err, sql.ErrNoRows) || existing.DiscordID == discordID {
		if dbErr := ctx.Database.UpsertUser(ctx.Context, db.UpsertUserParams{
			DiscordID:      discordID,
			LastfmUsername: username,
		}); dbErr != nil {
			logger.Log.Errorw("failed to upsert user", zlog.F{"gid": e.GuildID().String(), "uid": discordID}, dbErr)
			_ = res.Error(e, constants.ErrorSetUsername)
			return
		}

		reply.Content("your last.fm username has been set to **%s**", username).Edit()
	}

	if err != nil {
		logger.Log.Errorf("unexpected DB error in /set-user: %v", err)
		res.Error(e, constants.ErrorUnexpected)
	}
}
