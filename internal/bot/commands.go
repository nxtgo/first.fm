package bot

import (
	"context"
	"errors"
	"strings"
	"time"

	"first.fm/internal/lastfm"
	"first.fm/internal/logger"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	disgohandler "github.com/disgoorg/disgo/handler"
)

type CommandContext struct {
	*disgohandler.CommandEvent
	*Bot
}

type CommandHandler func(*CommandContext) error

var (
	allCommands []discord.ApplicationCommandCreate
	registry    = map[string]CommandHandler{}
)

func Register(meta discord.ApplicationCommandCreate, handler CommandHandler) {
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
			logger.Errorw("command failed", logger.F{"name": data.CommandName(), "err": err.Error()})
			_ = event.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("error: " + err.Error()).
				SetEphemeral(true).
				Build())
		}

		logger.Infow("executed command", logger.F{"name": data.CommandName(), "time": time.Since(start)})
	}
}

func (ctx *CommandContext) GetLastFMUser(name string) (*lastfm.UserInfo, error) {
	if name == "" {
		name = "user"
	}

	if rawUser, defined := ctx.SlashCommandInteractionData().OptString(name); defined {
		user, err := ctx.LastFM.User.Info(rawUser)
		return user, err
	}

	return nil, errors.New("automatic user detection is being worked on")
}

func normalizeUserInput(input string) string {
	if strings.HasPrefix(input, "<@") && strings.HasSuffix(input, ">") {
		trimmed := strings.TrimSuffix(strings.TrimPrefix(input, "<@"), ">")
		return strings.TrimPrefix(trimmed, "!")
	}
	return input
}
