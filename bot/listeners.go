package bot

import (
	"github.com/disgoorg/disgo/events"
	"go.fm/logger"
)

func (b *Bot) GuildReadyListener(e *events.GuildReady) {
	b.Cache.SetGuild(e.GuildID, e.Guild.Guild)
	logger.Log.Debugf("saved guild %s to cache", e.GuildID)

	for _, member := range e.Guild.Members {
		b.Cache.SetMember(member.User.ID, member)
		logger.Log.Debugf("saved member %s to cache", member.User.ID)
	}
}
