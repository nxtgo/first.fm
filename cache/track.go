package cache

import (
	"time"

	"go.fm/types/lastfm"
)

func (c *LastFMCache) GetTrack(key string) (*lastfm.TrackInfoResponse, bool) {
	return c.track.Get(key)
}

func (c *LastFMCache) SetTrack(key string, val *lastfm.TrackInfoResponse, ttl time.Duration) {
	c.track.Set(key, val, ttl)
}
