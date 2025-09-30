package bot

import (
	"context"
	"strings"
	"sync"
	"time"

	"first.fm/internal/emojis"
	"first.fm/internal/lastfm"
	"first.fm/internal/logger"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	disgohandler "github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

type CommandContext struct {
	*disgohandler.CommandEvent
	*Bot
}

type CommandHandler func(*CommandContext) error

var (
	allCommands []discord.ApplicationCommandCreate
	registry    = map[string]CommandHandler{}
	initOnce    sync.Once
)

func Register(meta discord.ApplicationCommandCreate, handler CommandHandler) {
	initOnce.Do(func() {
		allCommands = []discord.ApplicationCommandCreate{}
		registry = map[string]CommandHandler{}
	})

	logger.Infow("registered command", logger.F{"name": meta.CommandName()})
	allCommands = append(allCommands, meta)
	registry[meta.CommandName()] = handler
}

func Commands() []discord.ApplicationCommandCreate {
	return allCommands
}

func Dispatcher(bot *Bot) func(*events.ApplicationCommandInteractionCreate) {
	return func(event *events.ApplicationCommandInteractionCreate) {
		data := event.SlashCommandInteractionData()
		handler, ok := registry[data.CommandName()]
		if !ok {
			_ = event.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("unknown command").
				SetEphemeral(true).
				Build())
			return
		}

		start := time.Now()
		bgCtx := context.Background()
		ctx := &CommandContext{
			Bot: bot,
			CommandEvent: &disgohandler.CommandEvent{
				ApplicationCommandInteractionCreate: event,
				Ctx:                                 bgCtx,
			},
		}

		if err := handler(ctx); err != nil {
			logger.Warnw("command failed", logger.F{"name": data.CommandName(), "err": err.Error()})
			_ = ctx.CreateMessage(discord.NewMessageCreateBuilder().
				SetContentf("%s %v", emojis.EmojiCross, err).
				SetEphemeral(true).
				Build())
		}

		logger.Infow("executed command", logger.F{"name": data.CommandName(), "time": time.Since(start)})
	}
}

// now ima explain why this fucking function fetches the user everytime
// so first of all, it is cached so dont fucking worry ok.
// also this helps to cache the user for future requests do you get me
// so stfu ik this fucking function fetches the entire user instead of only returning
// a fucking username. ~elisiei
// edit: however, i should do an alternative function to get only the username anyways :kekw:. ~elisiei
func (ctx *CommandContext) GetLastFMUser(optionName string) (*lastfm.UserInfo, error) {
	if optionName == "" {
		optionName = "user"
	}

	rawUser, defined := ctx.SlashCommandInteractionData().OptString(optionName)
	if defined && rawUser != "" {
		if id, err := snowflake.Parse(normalizeUserMention(rawUser)); err == nil {
			if dbUser, err := ctx.Queries.GetUserByID(ctx.Ctx, id); err == nil {
				rawUser = dbUser.LastfmUsername
			}
		}

		return ctx.LastFM.User.Info(rawUser)
	}

	user, err := ctx.Queries.GetUserByID(ctx.Ctx, ctx.User().ID)
	if err != nil {
		return nil, err
	}
	return ctx.LastFM.User.Info(user.LastfmUsername)
}

func normalizeUserMention(input string) string {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "<@") && strings.HasSuffix(input, ">") {
		trimmed := strings.TrimSuffix(strings.TrimPrefix(input, "<@"), ">")
		return strings.TrimPrefix(trimmed, "!")
	}
	return input
}
