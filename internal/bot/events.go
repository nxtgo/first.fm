package bot

import (
	"context"

	"first.fm/internal/logger"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
)

func onReady(event *events.Ready) {
	logger.Info("started client")
	event.Client().SetPresence(context.Background(), gateway.WithCustomActivity("gwa gwa"))
}
