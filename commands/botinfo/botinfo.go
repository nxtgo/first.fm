package botinfo

import (
	"fmt"
	"runtime"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go.fm/util/res"
	"go.fm/util/shared/cmd"
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
	err := res.Reply(e).Defer()
	if err != nil {
		_ = res.ErrorReply(e, "error deferring command")
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	lastFMUsers, err := ctx.Database.GetUserCount(ctx.Context)
	if err != nil {
		lastFMUsers = 0
	}

	stats := fmt.Sprintf(
		"* registered last.fm users: %d\n"+
			"* goroutines: %d\n"+
			"* alloc: %.2f MB\n"+
			"* total alloc: %.2f MB\n"+
			"* sys: %.2f MB\n"+
			"* uptime: %s\n",
		lastFMUsers,
		runtime.NumGoroutine(),
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		time.Since(ctx.Start).Truncate(time.Second),
	)

	_ = res.Reply(e).Content(stats).Edit()
}
