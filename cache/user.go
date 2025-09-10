package cache

import (
	"time"

	"go.fm/types/lastfm"
)

func (c *LastFMCache) GetUser(key string) (*lastfm.UserInfoResponse, bool) {
	return c.user.Get(key)
}

func (c *LastFMCache) SetUser(key string, val *lastfm.UserInfoResponse, ttl time.Duration) {
	c.user.Set(key, val, ttl)
}
