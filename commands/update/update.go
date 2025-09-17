package update

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"go.fm/lfm"
	"go.fm/lfm/types"
	"go.fm/pkg/constants/emojis"
	"go.fm/pkg/constants/errs"
	"go.fm/pkg/constants/opts"
	"go.fm/pkg/ctx"
	"go.fm/pkg/discord/reply"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "update",
		Description: "refresh cache data",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
		Options: []discord.ApplicationCommandOption{
			opts.UserOption,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) {
	r := reply.New(e)
	if err := r.Defer(); err != nil {
		reply.Error(e, errs.ErrCommandDeferFailed)
		return
	}

	username, err := ctx.GetUser(e)
	if err != nil {
		reply.Error(e, errs.ErrUserNotFound)
		return
	}

	clearUserCaches(username, ctx)
	clearMemberCache(e, ctx)

	userInfo, err := fetchUserInfo(username, ctx)
	if err != nil {
		reply.Error(e, errs.ErrUserNotFound)
		return
	}

	updateCaches(e, username, userInfo, ctx)

	go ctx.LastFM.User.PrefetchUserData(username)

	r.Content("updated your data with fresh information %s", emojis.EmojiStar).Edit()
}

func clearUserCaches(username string, ctx ctx.CommandContext) {
	if ctx.Cache == nil {
		return
	}

	caches := []interface {
		Delete(string)
	}{
		ctx.Cache.User,
		ctx.Cache.TopArtists,
		ctx.Cache.TopAlbums,
		ctx.Cache.TopTracks,
	}

	for _, cache := range caches {
		if cache != nil {
			cache.Delete(username)
		}
	}
}

func clearMemberCache(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) {
	guildID := e.GuildID()
	if guildID == nil || ctx.Cache.Members == nil {
		return
	}

	members, exists := ctx.Cache.Members.Get(*guildID)
	if !exists || members == nil {
		return
	}

	delete(members, e.User().ID)
	ctx.Cache.Members.Set(*guildID, members, 0)
}

func fetchUserInfo(username string, ctx ctx.CommandContext) (*types.UserGetInfo, error) {
	return ctx.LastFM.User.GetInfo(lfm.P{"user": username})
}

func updateCaches(e *events.ApplicationCommandInteractionCreate, username string, userInfo *types.UserGetInfo, ctx ctx.CommandContext) {
	if ctx.Cache.User != nil {
		ctx.Cache.User.Set(username, *userInfo, 0)
	}

	updateGuildMemberCache(e, username, ctx)
}

func updateGuildMemberCache(e *events.ApplicationCommandInteractionCreate, username string, ctx ctx.CommandContext) {
	guildID := e.GuildID()
	if guildID == nil || ctx.Cache.Members == nil {
		return
	}

	members, _ := ctx.Cache.Members.Get(*guildID)
	if members == nil {
		members = make(map[snowflake.ID]string)
	}

	members[e.User().ID] = username
	ctx.Cache.Members.Set(*guildID, members, 0)
}
