package commands

import (
	"context"
	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/api/cmdroute"
	"github.com/nxtgo/arikawa/v3/state"
	"go.fm/db"
	lastfm "go.fm/last.fm"
	"go.fm/pkg/reply"
	"go.fm/zlog"
	"os"
	"time"
)

var allCommands = []api.CreateCommandData{}
var registry = map[string]CommandHandler{}

func Register(meta api.CreateCommandData, handler CommandHandler) {
	zlog.Log.Debugf("registered command %s", meta.Name)

	allCommands = append(allCommands, meta)
	registry[meta.Name] = handler
}

func RegisterCommands(r *cmdroute.Router, st *state.State, q *db.Queries, c *lastfm.Cache) {
	lastFMApiKey := os.Getenv("LASTFM_API_KEY")
	if lastFMApiKey == "" {
		zlog.Log.Fatal("missing LASTFM_API_KEY env")
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
			zlog.Log.Debugw("executed command %s", zlog.F{"time": time.Since(start)}, name)
			// debugging purposes
			if err != nil {
				zlog.Log.Warn(err.Error())
				commandContext.Reply.QuickEmbed(reply.ErrorEmbed(err.Error()))
			}

			return nil
		})
	}
}

func Sync(st *state.State) error {
	defer zlog.Log.Infow("synced commands", zlog.F{"count": len(allCommands)})
	return cmdroute.OverwriteCommands(st, allCommands)
}
