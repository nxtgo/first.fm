package commands

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/bot/cache"
	"go.fm/db"
	"go.fm/lastfm"
	"go.fm/logger"
	"go.fm/util/res"
)

type SetUserCommand struct{}

func (SetUserCommand) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "set-user",
		Description: "link your Last.fm username to this Discord server",
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
	guildID := e.GuildID().String()
	discordID := e.Member().User.ID.String()

	exists, err := userInfoCache.GetOrFetch(
		username,
		func(username string) (*lastfm.UserInfoResponse, error) {
			return ctx.LastFM.GetUserInfo(username)
		})
	if exists.User.Name == "" || err != nil {
		_ = res.ErrorReply(e, "that username doesn't exist in last.fm")
		return
	}

	existingByName, err := ctx.Database.GetDiscordByUsernameAndGuild(context.Background(),
		db.GetDiscordByUsernameAndGuildParams{
			Username: username,
			GuildID:  guildID,
		})
	if err == nil && existingByName.DiscordID != discordID {
		_ = res.ErrorReply(e, "that last.fm username is already linked by another member in this server")
		return
	}

	existingUser, err := ctx.Database.GetUserByDiscordIDAndGuild(context.Background(),
		db.GetUserByDiscordIDAndGuildParams{
			DiscordID: discordID,
			GuildID:   guildID,
		})

	if err == nil {
		if existingUser.Username == username {
			_ = reply.Content("your last.fm username is already set to **%s**", username).Send()
			return
		}

		if err := ctx.Database.UpdateUsername(context.Background(), db.UpdateUsernameParams{
			DiscordID: discordID,
			GuildID:   guildID,
			Username:  username,
		}); err != nil {
			logger.Log.Errorf("failed to update username for user %s in guild %s: %v", discordID, guildID, err)
			_ = res.ErrorReply(e, "failed to update your last.fm username")
			return
		}

		_ = reply.Content("updated your last.fm username to **%s**", username).Send()
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		if err := ctx.Database.InsertUser(context.Background(), db.InsertUserParams{
			GuildID:   guildID,
			DiscordID: discordID,
			Username:  username,
		}); err != nil {
			logger.Log.Errorf("failed to insert user %s in guild %s: %v", discordID, guildID, err)
			_ = res.ErrorReply(e, "failed to set your last.fm username")
			return
		}

		_ = reply.Content("set your last.fm username to **%s**", username).Send()
		return
	}

	logger.Log.Errorf("unexpected DB error in /set-user: %v", err)
	_ = res.ErrorReply(e, "an unexpected error occurred, please try again later")
}

var userInfoCache *cache.FuncCache[string, *lastfm.UserInfoResponse]

func init() {
	userInfoCache = cache.NewFuncCache[string, *lastfm.UserInfoResponse](time.Hour * 6)
	Register(SetUserCommand{})
}
