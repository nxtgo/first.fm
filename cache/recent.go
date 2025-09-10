package cache

import (
	"time"

	"go.fm/types/lastfm"
)

func (c *LastFMCache) GetRecentTracks(key string) (*lastfm.RecentTracksResponse, bool) {
	return c.recentTracks.Get(key)
}

func (c *LastFMCache) SetRecentTracks(key string, val *lastfm.RecentTracksResponse, ttl time.Duration) {
	c.recentTracks.Set(key, val, ttl)
}
