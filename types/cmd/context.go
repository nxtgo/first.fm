package cmd

import (
	"context"
	"time"

	"go.fm/cache"
	"go.fm/db"
	"go.fm/lastfm"
)

type CommandContext struct {
	LastFM   *lastfm.Client
	Database *db.Queries
	Context  context.Context
	Start    time.Time
	Cache    *cache.LastFMCache
}
