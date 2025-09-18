package commands

import (
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/commands/botinfo"
	"go.fm/commands/fm"
	"go.fm/commands/profile"
	profilev2 "go.fm/commands/profile/v2"
	"go.fm/commands/setuser"
	"go.fm/commands/top"
	"go.fm/commands/update"
	"go.fm/commands/whoknows"

	"go.fm/logger"
	"go.fm/pkg/ctx"
)

var registry = map[string]Command{}
var sharedCtx ctx.CommandContext

const (
	DefaultWorkerCount = 50
	DefaultQueueSize   = 1000
)

type CommandJob struct {
	e   *events.ApplicationCommandInteractionCreate
	ctx ctx.CommandContext
}

var jobQueue = make(chan CommandJob, DefaultQueueSize)

func init() {
	Register(fm.Command{})
	Register(whoknows.Command{})
	Register(setuser.Command{})
	Register(profile.Command{})
	Register(top.Command{})
	Register(update.Command{})
	Register(botinfo.Command{})
	Register(profilev2.Command{})
}

type Command interface {
	Data() discord.ApplicationCommandCreate
	Handle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext)
}

func Register(cmd Command) {
	name := cmd.Data().CommandName()
	registry[name] = cmd
	logger.Log.Debugf("registered command %s", name)
}

func All() []discord.ApplicationCommandCreate {
	cmds := make([]discord.ApplicationCommandCreate, 0, len(registry))
	for _, cmd := range registry {
		cmds = append(cmds, cmd.Data())
	}
	return cmds
}

func InitDependencies(ctx ctx.CommandContext) {
	ctx.Start = time.Now()
	sharedCtx = ctx
	StartWorkerPool(DefaultWorkerCount)
}

func StartWorkerPool(numWorkers int) {
	for i := range numWorkers {
		go func(workerID int) {
			for job := range jobQueue {
				safeHandle(job.e, job.ctx)
			}
		}(i)
	}
	logger.Log.Infof("started %d command workers", numWorkers)
}

func safeHandle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("panic in command %s: %v", e.Data.CommandName(), r)
		}
	}()

	cmd, ok := registry[e.Data.CommandName()]
	if !ok {
		logger.Log.Warnf("unknown command: %s", e.Data.CommandName())
		return
	}

	start := time.Now()
	cmd.Handle(e, ctx)
	logger.Log.Debugf("command %s executed in %s", e.Data.CommandName(), time.Since(start))
}

func Handler() bot.EventListener {
	return &events.ListenerAdapter{
		OnApplicationCommandInteraction: func(e *events.ApplicationCommandInteractionCreate) {
			select {
			case jobQueue <- CommandJob{e: e, ctx: sharedCtx}:
			default:
				logger.Log.Warnf("command queue full, dropping command: %s", e.Data.CommandName())
			}
		},
	}
}
