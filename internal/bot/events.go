package bot

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
)

func onReady(event *events.Ready) {
	slog.Info("started client")
	event.Client().SetPresence(context.Background(), gateway.WithCustomActivity("gwa gwa"))
}
