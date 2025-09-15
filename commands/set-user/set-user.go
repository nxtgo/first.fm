package setuser

import (
	"database/sql"
	"errors"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nxtgo/zlog"

	"go.fm/db"
	"go.fm/lfm"
	"go.fm/logger"
	"go.fm/pkg/constants/errs"
	"go.fm/types/cmd"
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
	reply := ctx.Reply(e)
	if err := reply.Defer(); err != nil {
		ctx.Error(e, errs.ErrCommandDeferFailed)
		return
	}

	username := e.SlashCommandInteractionData().String("username")
	discordID := e.User().ID.String()

	_, err := ctx.LastFM.User.GetInfo(lfm.P{"user": username})
	if err != nil {
		ctx.Error(e, errs.ErrUserNotFound)
		return
	}

	existing, err := ctx.Database.GetUserByUsername(ctx.Context, username)
	if err == nil {
		if existing.DiscordID != discordID {
			ctx.Error(e, errs.ErrUsernameAlreadyUsed)
			return
		}
		if existing.LastfmUsername == username {
			ctx.Error(e, errs.ErrUsernameAlreadySet(username))
			return
		}
	}

	if errors.Is(err, sql.ErrNoRows) || existing.DiscordID == discordID {
		if dbErr := ctx.Database.UpsertUser(ctx.Context, db.UpsertUserParams{
			DiscordID:      discordID,
			LastfmUsername: username,
		}); dbErr != nil {
			logger.Log.Errorw("failed to upsert user", zlog.F{"gid": e.GuildID().String(), "uid": discordID}, dbErr)
			ctx.Error(e, errs.ErrSetUsername)
			return
		}

		reply.Content("your last.fm username has been set to **%s**", username).Edit()
	}
}
