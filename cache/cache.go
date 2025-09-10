package cache

import (
	"fmt"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/nxtgo/gce"
	"go.fm/types/lastfm"
)

type LastFMCache struct {
	user         *gce.Cache[string, *lastfm.UserInfoResponse]
	track        *gce.Cache[string, *lastfm.TrackInfoResponse]
	recentTracks *gce.Cache[string, *lastfm.RecentTracksResponse]
	plays        *gce.Cache[string, int]
	guildUsers   *gce.Cache[snowflake.ID, map[snowflake.ID]string]
}

// NewLastFMCache creates caches with tuned defaults for a Discord bot.
func NewLastFMCache() *LastFMCache {
	commonCleanup := gce.WithCleanupInterval(30 * time.Second)
	commonShards := gce.WithShardCount(64)

	return &LastFMCache{
		user: gce.New[string, *lastfm.UserInfoResponse](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(5*time.Minute),
			gce.WithMaxEntries(10_000),
		),
		track: gce.New[string, *lastfm.TrackInfoResponse](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(30*time.Minute),
			gce.WithMaxEntries(50_000),
		),
		recentTracks: gce.New[string, *lastfm.RecentTracksResponse](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(1*time.Minute),
			gce.WithMaxEntries(10_000),
		),
		plays: gce.New[string, int](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(10*time.Minute),
			gce.WithMaxEntries(100_000),
		),
		guildUsers: gce.New[snowflake.ID, map[snowflake.ID]string](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(5*time.Minute),
			gce.WithMaxEntries(100),
		),
	}
}

func (c *LastFMCache) Stats() string {
	stats := "```\n"

	if c.user != nil {
		s := c.user.Stats()
		stats += fmt.Sprintf("user cache       | hits: %6d | misses: %6d | loads: %6d | evictions: %6d | size: %6d\n",
			s.Hits, s.Misses, s.Loads, s.Evictions, s.CurrentSize)
	}

	if c.track != nil {
		s := c.track.Stats()
		stats += fmt.Sprintf("track cache      | hits: %6d | misses: %6d | loads: %6d | evictions: %6d | size: %6d\n",
			s.Hits, s.Misses, s.Loads, s.Evictions, s.CurrentSize)
	}

	if c.recentTracks != nil {
		s := c.recentTracks.Stats()
		stats += fmt.Sprintf("recenttracks     | hits: %6d | misses: %6d | loads: %6d | evictions: %6d | size: %6d\n",
			s.Hits, s.Misses, s.Loads, s.Evictions, s.CurrentSize)
	}

	if c.plays != nil {
		s := c.plays.Stats()
		stats += fmt.Sprintf("plays cache      | hits: %6d | misses: %6d | loads: %6d | evictions: %6d | size: %6d\n",
			s.Hits, s.Misses, s.Loads, s.Evictions, s.CurrentSize)
	}

	if c.guildUsers != nil {
		s := c.guildUsers.Stats()
		stats += fmt.Sprintf("guildusers cache | hits: %6d | misses: %6d | loads: %6d | evictions: %6d | size: %6d\n",
			s.Hits, s.Misses, s.Loads, s.Evictions, s.CurrentSize)
	}

	stats += "```"
	return stats
}

func (c *LastFMCache) Close() {
	if c.user != nil {
		c.user.Close()
	}
	if c.track != nil {
		c.track.Close()
	}
	if c.recentTracks != nil {
		c.recentTracks.Close()
	}
	if c.plays != nil {
		c.plays.Close()
	}
}
