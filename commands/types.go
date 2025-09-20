package commands

import (
	"context"
	"strings"

	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"go.fm/db"
	lastfm "go.fm/last.fm"
	"go.fm/pkg/reply"
)

type CommandContext struct {
	Context context.Context
	Data    cmdroute.CommandData
	State   *state.State
	Reply   *reply.ResponseManager
	Query   *db.Queries
	Last    *lastfm.Services
}

type CommandHandler func(c *CommandContext) error

func (ctx *CommandContext) GetUserOrFallback() (string, error) {
	optionData := ctx.Data.Options.Find("user")
	option := optionData.String()

	if option == "" {
		user, err := ctx.Query.GetUserByID(ctx.Context, ctx.Data.Event.Member.User.ID.String())
		if err != nil {
			return "", err
		}
		return user.LastfmUsername, nil
	}

	userOption := normalizeUserInput(option)

	if _, err := discord.ParseSnowflake(userOption); err == nil {
		user, err := ctx.Query.GetUserByID(ctx.Context, userOption)
		if err != nil {
			return "", err
		}
		return user.LastfmUsername, nil
	}

	return userOption, nil
}

func normalizeUserInput(input string) string {
	if strings.HasPrefix(input, "<@") && strings.HasSuffix(input, ">") {
		trimmed := strings.TrimSuffix(strings.TrimPrefix(input, "<@"), ">")
		return strings.TrimPrefix(trimmed, "!")
	}
	return input
}
