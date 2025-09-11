package botinfo

import (
	"fmt"
	"runtime"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go.fm/constants"
	"go.fm/types/cmd"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "botinfo",
		Description: "display go.fm's info",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := ctx.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = ctx.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	lastFMUsers, err := ctx.Database.GetUserCount(ctx.Context)
	if err != nil {
		lastFMUsers = 0
	}

	stats := fmt.Sprintf(
		"```\n"+
			"registered last.fm users: %d\n"+
			"goroutines: %d\n"+
			"memory Usage:\n"+
			"  - alloc: %.2f MB\n"+
			"  - total: %.2f MB\n"+
			"  - sys: %.2f MB\n"+
			"uptime: %s\n"+
			"```\n"+
			"**cache stats:**\n%s",
		lastFMUsers,
		runtime.NumGoroutine(),
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		time.Since(ctx.Start).Truncate(time.Second),
		ctx.LastFM.CacheStats(),
	)

	reply.Content(stats).Edit()
}
