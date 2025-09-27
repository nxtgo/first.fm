package commands

import (
	"context"
	"os"
	"time"

	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/api/cmdroute"
	"github.com/nxtgo/arikawa/v3/state"
	"go.fm/internal/bot/discord/reply"
	"go.fm/internal/bot/lastfm"
	"go.fm/internal/bot/logging"
	"go.fm/internal/bot/persistence/sqlc"
)

var allCommands = []api.CreateCommandData{}
var registry = map[string]CommandHandler{}

func Register(meta api.CreateCommandData, handler CommandHandler) {
	logging.Debugf("registered command %s", meta.Name)

	allCommands = append(allCommands, meta)
	registry[meta.Name] = handler
}

func RegisterCommands(r *cmdroute.Router, st *state.State, q *sqlc.Queries, c *lastfm.Cache) {
	lastFMApiKey := os.Getenv("LASTFM_API_KEY")
	if lastFMApiKey == "" {
		logging.Fatal("missing LASTFM_API_KEY env")
	}

	for name, handler := range registry {
		h := handler
		r.AddFunc(name, func(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
			commandContext := &CommandContext{
				// command mandatory stuff
				Context: ctx,
				Data:    data,
				State:   st,

				// reply helper
				Reply: reply.New(st, data.Event),

				// database
				Query: q,

				// last.fm stuff
				Last:  lastfm.NewServices(lastFMApiKey, c),
				Cache: c,
			}

			// debugging purposes
			start := time.Now()
			err := h(commandContext)
			logging.Debugw("executed command %s", logging.F{"time": time.Since(start)}, name)

			if err != nil {
				commandContext.Reply.QuickEmbed(reply.ErrorEmbed(err.Error()))
			}

			return nil
		})
	}
}

func Sync(st *state.State) error {
	defer logging.Infow("synced commands", logging.F{"count": len(allCommands)})
	return cmdroute.OverwriteCommands(st, allCommands)
}
