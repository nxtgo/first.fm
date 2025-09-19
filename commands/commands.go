package commands

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/state"

	"go.fm/zlog"
)

type CommandContext struct {
	Ctx  context.Context
	Data cmdroute.CommandData
	St   *state.State
}

type CommandHandler func(c *CommandContext) *api.InteractionResponseData

var allCommands = []api.CreateCommandData{}
var registry = map[string]CommandHandler{}

func Register(meta api.CreateCommandData, handler CommandHandler) {
	zlog.Log.Debugf("registered command %s", meta.Name)

	allCommands = append(allCommands, meta)
	registry[meta.Name] = handler
}

func RegisterCommands(r *cmdroute.Router, st *state.State) {
	for name, handler := range registry {
		h := handler
		r.AddFunc(name, func(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
			return h(&CommandContext{
				Ctx:  ctx,
				Data: data,
				St:   st,
			})
		})
	}
}

func Sync(st *state.State) error {
	zlog.Log.Debug("synced commands")

	return cmdroute.OverwriteCommands(st, allCommands)
}
