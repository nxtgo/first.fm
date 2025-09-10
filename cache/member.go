package cache

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
)

func (c *LastFMCache) GetMembers(guildID snowflake.ID) (map[snowflake.ID]string, bool) {
	return c.member.Get(guildID)
}

func (c *LastFMCache) SetMembers(guildID snowflake.ID, val map[snowflake.ID]string, ttl time.Duration) {
	c.member.Set(guildID, val, ttl)
}
