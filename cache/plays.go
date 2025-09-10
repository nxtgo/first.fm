package cache

import (
	"time"
)

func (c *LastFMCache) GetPlays(key string) (int, bool) {
	return c.plays.Get(key)
}
func (c *LastFMCache) SetPlays(key string, val int, ttl time.Duration) {
	c.plays.Set(key, val, ttl)
}
