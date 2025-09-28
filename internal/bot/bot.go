package bot

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"first.fm/internal/lastfm/api"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
)

type Bot struct {
	Client *bot.Client
	LastFM *api.Client
}

func New(token, key string) (*Bot, error) {
	client, err := disgo.New(
		token,
		bot.WithGatewayConfigOpts(
			gateway.WithCompress(true),
			gateway.WithAutoReconnect(true),
			gateway.WithIntents(
				gateway.IntentGuildMembers,
				gateway.IntentGuilds,
			),
		),
		bot.WithEventListenerFunc(onReady),
		bot.WithEventListenerFunc(Dispatcher()),
	)
	if err != nil {
		return nil, err
	}

	lastfmClient := api.NewClient(key)
	return &Bot{
		Client: client,
		LastFM: lastfmClient,
	}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	if err := b.Client.OpenGateway(ctx); err != nil {
		return err
	}
	defer b.Client.Close(ctx)

	if _, err := b.Client.Rest.SetGuildCommands(b.Client.ApplicationID, snowflake.GetEnv("GUILD_ID"), Commands()); err != nil {
		return err
	}
	slog.Info("registered discord commands")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-stop:
	}

	return nil
}
