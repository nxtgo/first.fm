package cache

import (
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"

	"github.com/nxtgo/gache"
)

type CustomCaches struct {
	guilds  *gache.Cache[snowflake.ID, discord.Guild]
	members *gache.Cache[snowflake.ID, discord.Member]
	ttl     time.Duration
}

func New() *CustomCaches {
	return &CustomCaches{
		guilds:  gache.New[snowflake.ID, discord.Guild](time.Hour * 24),
		members: gache.New[snowflake.ID, discord.Member](time.Hour * 12),
		ttl:     time.Hour * 24 * 365,
	}
}

func (c *CustomCaches) GetGuild(id snowflake.ID) (discord.Guild, bool) {
	g, ok := c.guilds.Get(id)
	if ok {
		c.guilds.Set(id, g, c.ttl)
	}
	return g, ok
}

func (c *CustomCaches) SetGuild(id snowflake.ID, value discord.Guild) {
	c.guilds.Set(id, value, c.ttl)
}

func (c *CustomCaches) GetMember(id snowflake.ID) (discord.Member, bool) {
	m, ok := c.members.Get(id)
	if ok {
		c.members.Set(id, m, c.ttl)
	}
	return m, ok
}

func (c *CustomCaches) SetMember(id snowflake.ID, value discord.Member) {
	c.members.Set(id, value, c.ttl)
}

func (c *CustomCaches) DeleteMember(id snowflake.ID) {
	c.members.Delete(id)
}
