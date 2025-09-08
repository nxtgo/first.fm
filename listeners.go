package main

import (
	"github.com/disgoorg/disgo/events"
	"go.fm/logger"
)

func GuildReadyListener(e *events.GuildReady) {
	logger.Log.Debugf("saved guild %s to cache", e.GuildID)
	sharedCache.SaveGuild(e.GuildID, e.Guild.Guild)
}
