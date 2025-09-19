package botinfo

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/pkg/constants/errs"
	"go.fm/pkg/ctx"
	"go.fm/pkg/discord/markdown"
	"go.fm/pkg/discord/reply"
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

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) {
	r := reply.New(e)

	if err := r.Defer(); err != nil {
		reply.Error(e, errs.ErrCommandDeferFailed)
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	lastFMUsers, err := ctx.Database.GetUserCount(ctx.Context)
	if err != nil {
		lastFMUsers = 0
	}

	branch, commit, message := getGitInfo()
	botStats := [][2]string{
		{"users", fmt.Sprintf("%d", lastFMUsers)},
		{"goroutines", fmt.Sprintf("%d", runtime.NumGoroutine())},
		{"alloc", fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024)},
		{"total", fmt.Sprintf("%.2f MB", float64(m.TotalAlloc)/1024/1024)},
		{"sys", fmt.Sprintf("%.2f MB", float64(m.Sys)/1024/1024)},
		{"uptime", time.Since(ctx.Start).Truncate(time.Second).String()},
		{"branch", branch},
		{"commit", fmt.Sprintf("%s (%s)", commit, message)},
	}

	cacheStats := ctx.Cache.Stats()
	cacheRows := make([][2]string, 0, len(cacheStats))
	for _, cs := range cacheStats {
		cacheRow := fmt.Sprintf(
			"hits=%d misses=%d loads=%d evicts=%d size=%d",
			cs.Stats.Hits,
			cs.Stats.Misses,
			cs.Stats.Loads,
			cs.Stats.Evictions,
			cs.Stats.CurrentSize,
		)
		cacheRows = append(cacheRows, [2]string{cs.Name, cacheRow})
	}

	botTable := markdown.MD(markdown.GenerateTable(botStats)).CodeBlock("ts")
	cacheTable := markdown.MD(markdown.GenerateTable(cacheRows)).CodeBlock("ts")

	r.Content("## bot stats\n%s\n## cache stats\n%s", botTable, cacheTable).Edit()

}

func getGitInfo() (branch, commit, message string) {
	branch = runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	commit = runGitCommand("rev-parse", "--short", "HEAD")
	message = runGitCommand("log", "-1", "--pretty=%B")
	return
}

func runGitCommand(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
