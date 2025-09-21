package events

import (
	"github.com/nxtgo/arikawa/v3/gateway"
	"go.fm/zlog"
)

var Events []any
var EventLogger func(name string) *zlog.Logger

func init() {
	Events = append(Events, EventReady)
	EventLogger = func(name string) *zlog.Logger {
		return zlog.WithFields(zlog.F{"event_name": name})
	}
}

func EventReady(c *gateway.ReadyEvent) {
	EventLogger("ready").Infow("client ready", zlog.F{
		"tag":    c.User.Tag(),
		"guilds": len(c.Guilds),
	})
}
