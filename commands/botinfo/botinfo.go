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
	cacheStats := ctx.Cache.Stats()

	var sb strings.Builder

	fmt.Fprintf(&sb, "registered last.fm users: %d\n", lastFMUsers)
	fmt.Fprintf(&sb, "goroutines: %d\n", runtime.NumGoroutine())
	fmt.Fprintf(&sb, "memory usage:\n")
	fmt.Fprintf(&sb, "  - alloc: %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Fprintf(&sb, "  - total: %.2f MB\n", float64(m.TotalAlloc)/1024/1024)
	fmt.Fprintf(&sb, "  - sys  : %.2f MB\n", float64(m.Sys)/1024/1024)
	fmt.Fprintf(&sb, "uptime: %s\n", time.Since(ctx.Start).Truncate(time.Second))
	fmt.Fprintf(&sb, "git:\n")
	fmt.Fprintf(&sb, "  - branch : %s\n", branch)
	fmt.Fprintf(&sb, "  - commit : %s\n", commit)
	fmt.Fprintf(&sb, "  - message: %s\n", message)

	headers := []string{"cache", "hits", "misses", "loads", "evicts", "size"}
	rows := make([][]string, len(cacheStats))
	for i, cs := range cacheStats {
		rows[i] = []string{
			cs.Name,
			fmt.Sprintf("%d", cs.Stats.Hits),
			fmt.Sprintf("%d", cs.Stats.Misses),
			fmt.Sprintf("%d", cs.Stats.Loads),
			fmt.Sprintf("%d", cs.Stats.Evictions),
			fmt.Sprintf("%d", cs.Stats.CurrentSize),
		}
	}

	cacheTable := markdown.MD(markdown.Table(headers, rows)).CodeBlock("ts")
	botStats := markdown.MD(sb.String()).CodeBlock("yaml")

	r.Content("## bot stats\n%s\n## cache stats\n%s", botStats, cacheTable).Edit()
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
