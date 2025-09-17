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
	"go.fm/pkg/ctx"
	"go.fm/pkg/discord/reply"
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

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) {
	r := reply.New(e)
	if err := r.Defer(); err != nil {
		reply.Error(e, errs.ErrCommandDeferFailed)
		return
	}

	username := e.SlashCommandInteractionData().String("username")
	discordID := e.User().ID.String()

	_, err := ctx.LastFM.User.GetInfo(lfm.P{"user": username})
	if err != nil {
		reply.Error(e, errs.ErrUserNotFound)
		return
	}

	existing, err := ctx.Database.GetUserByUsername(ctx.Context, username)
	if err == nil {
		if existing.DiscordID != discordID {
			reply.Error(e, errs.ErrUsernameAlreadyUsed)
			return
		}
		if existing.LastfmUsername == username {
			reply.Error(e, errs.ErrUsernameAlreadySet(username))
			return
		}
	}

	if errors.Is(err, sql.ErrNoRows) || existing.DiscordID == discordID {
		if dbErr := ctx.Database.UpsertUser(ctx.Context, db.UpsertUserParams{
			DiscordID:      discordID,
			LastfmUsername: username,
		}); dbErr != nil {
			logger.Log.Errorw("failed to upsert user", zlog.F{"gid": e.GuildID().String(), "uid": discordID}, dbErr)
			reply.Error(e, errs.ErrSetUsername)
			return
		}

		r.Content("your last.fm username has been set to **%s**", username).Edit()
	}
}
