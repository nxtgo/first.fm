package cmd

import (
	"context"

	"go.fm/db"
	"go.fm/lastfm"
)

type CommandContext struct {
	LastFM       *lastfm.Client
	Database     *db.Queries
	QueryContext context.Context
	Context      context.Context
}
