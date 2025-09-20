package commands

import (
	"context"
	"os"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/state"
	"go.fm/db"
	lastfm "go.fm/last.fm"
	"go.fm/pkg/reply"
	"go.fm/zlog"
)

var allCommands = []api.CreateCommandData{}
var registry = map[string]CommandHandler{}

func Register(meta api.CreateCommandData, handler CommandHandler) {
	zlog.Log.Debugf("registered command %s", meta.Name)

	allCommands = append(allCommands, meta)
	registry[meta.Name] = handler
}

func RegisterCommands(r *cmdroute.Router, st *state.State, q *db.Queries) {
	lastFMApiKey := os.Getenv("LASTFM_API_KEY")
	if lastFMApiKey == "" {
		zlog.Log.Fatal("missing LASTFM_API_KEY env")
	}

	for name, handler := range registry {
		h := handler
		r.AddFunc(name, func(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
			commandContext := &CommandContext{
				Context: ctx,
				Data:    data,
				State:   st,
				Reply:   reply.New(st, data.Event),
				Query:   q,
				Last:    lastfm.NewServicesWithAPIKey(lastFMApiKey, lastfm.WithTimeout(time.Second*15)),
			}

			defer func() {
				if rec := recover(); rec != nil {
					zlog.Log.Errorf("panic occurred in command handler: %v", rec)
					commandContext.Reply.QuickEmbed(reply.ErrorEmbed("something happened :("))
				}
			}()

			err := h(commandContext)
			if err != nil {
				commandContext.Reply.QuickEmbed(reply.ErrorEmbed(err.Error()))
			}

			return nil
		})
	}
}

func Sync(st *state.State) error {
	defer zlog.Log.Debug("synced commands")
	return cmdroute.OverwriteCommands(st, allCommands)
}
