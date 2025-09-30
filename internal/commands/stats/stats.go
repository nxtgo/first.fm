package stats

import (
	"fmt"
	"runtime"
	"time"

	"first.fm/internal/bot"
	"github.com/disgoorg/disgo/discord"
)

var startTime = time.Now()

func init() {
	bot.Register(data, handle)
}

var data = discord.SlashCommandCreate{
	Name:        "stats",
	Description: "display first.fm stats",
	IntegrationTypes: []discord.ApplicationIntegrationType{
		discord.ApplicationIntegrationTypeGuildInstall,
		discord.ApplicationIntegrationTypeUserInstall,
	},
}

func handle(ctx *bot.CommandContext) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	statsText := fmt.Sprintf(
		"uptime: %s\n"+
			"goroutines: %d\n"+
			"os threads: %d\n"+
			"memory allocated: %s\n"+
			"total allocated: %s\n"+
			"system memory: %s\n"+
			"gc runs: %d\n"+
			"last gc pause: %.2fms\n"+
			"go version: %s\n",
		formatUptime(time.Since(startTime)),
		runtime.NumGoroutine(),
		runtime.NumCPU(),
		formatBytes(m.Alloc),
		formatBytes(m.TotalAlloc),
		formatBytes(m.Sys),
		m.NumGC,
		float64(m.PauseNs[(m.NumGC+255)%256])/1e6,
		runtime.Version(),
	)

	component := discord.NewContainer(
		discord.NewTextDisplay(statsText),
	).WithAccentColor(0x00ADD8)

	return ctx.CreateMessage(
		discord.NewMessageCreateBuilder().
			SetIsComponentsV2(true).
			SetComponents(component).
			Build(),
	)
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
