package bot

import (
	"github.com/disgoorg/disgo/bot"
	"go.fm/bot/cache"
	"go.fm/db"
	"go.fm/lastfm"
)

type Bot struct {
	Client   *bot.Client
	Cache    *cache.CustomCaches
	LastFM   *lastfm.Client
	Database *db.Queries
}
