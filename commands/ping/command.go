package ping

import (
	"fmt"
	"runtime"
	"time"

	"go.fm/commands"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

var startTime = time.Now()

var data = api.CreateCommandData{
	Name:        "stats",
	Description: "display bot's stats",
}

func handler(c *commands.CommandContext) *api.InteractionResponseData {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime).Round(time.Second)

	stats := fmt.Sprintf(
		"uptime: %s\n"+
			"goroutines: %d\n"+
			"memory: %.2f mb (heap: %.2f mb)\n"+
			"gc runs: %d\n"+
			"go version: %s %s/%s",
		uptime,
		runtime.NumGoroutine(),
		float64(m.Alloc)/(1024*1024),
		float64(m.HeapAlloc)/(1024*1024),
		m.NumGC,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)

	return &api.InteractionResponseData{
		Content: option.NewNullableString(stats),
	}
}

func init() {
	commands.Register(data, handler)
}
