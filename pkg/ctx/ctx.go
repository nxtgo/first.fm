package ctx

import (
	"context"
	"strings"
	"time"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"go.fm/cache"
	"go.fm/db"
	"go.fm/lfm"
)

type CommandContext struct {
	LastFM   *lfm.LastFMApi
	Database *db.Queries
	Context  context.Context
	Start    time.Time
	Cache    *cache.Cache
}

func (ctx *CommandContext) GetUser(
	e *events.ApplicationCommandInteractionCreate,
) (string, error) {
	if rawUser, defined := e.SlashCommandInteractionData().OptString("user"); defined {
		userID := normalizeUserInput(rawUser)

		if _, err := snowflake.Parse(userID); err == nil {
			return ctx.Database.GetUser(ctx.Context, userID)
		}

		return rawUser, nil
	}

	userID := e.Member().User.ID.String()
	return ctx.Database.GetUser(ctx.Context, userID)
}

func normalizeUserInput(input string) string {
	if strings.HasPrefix(input, "<@") && strings.HasSuffix(input, ">") {
		trimmed := strings.TrimSuffix(strings.TrimPrefix(input, "<@"), ">")
		return strings.TrimPrefix(trimmed, "!")
	}
	return input
}
