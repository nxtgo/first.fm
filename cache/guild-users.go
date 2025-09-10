package cache

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
)

func (c *LastFMCache) GetGuildUsers(guildID snowflake.ID) (map[snowflake.ID]string, bool) {
	return c.guildUsers.Get(guildID)
}

func (c *LastFMCache) SetGuildUsers(guildID snowflake.ID, val map[snowflake.ID]string, ttl time.Duration) {
	c.guildUsers.Set(guildID, val, ttl)
}
