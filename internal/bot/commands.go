package bot

import (
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type CommandContext struct {
	Context context.Context
	Event   *events.ApplicationCommandInteractionCreate
	Bot     *Bot
}

type CommandHandler func(*CommandContext) error

var (
	allCommands []discord.ApplicationCommandCreate
	registry    = map[string]CommandHandler{}
)

func Register(meta discord.ApplicationCommandCreate, handler CommandHandler) {
	slog.Info("registered command", "name", meta.CommandName())
	allCommands = append(allCommands, meta)
	registry[meta.CommandName()] = handler
}

func Commands() []discord.ApplicationCommandCreate {
	return allCommands
}

func Dispatcher() func(*events.ApplicationCommandInteractionCreate) {
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
		ctx := &CommandContext{
			Context: context.Background(),
			Event:   event,
		}

		if err := handler(ctx); err != nil {
			slog.Error("command failed", "name", data.CommandName(), "err", err)
			_ = event.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("error: " + err.Error()).
				SetEphemeral(true).
				Build())
		}

		slog.Info("executed command", "name", data.CommandName(), "time", time.Since(start))
	}
}
