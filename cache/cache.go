package cache

import (
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"

	"github.com/nxtgo/gache"
)

type CustomCaches struct {
	guilds *gache.Cache[snowflake.ID, discord.Guild]
}

func New() *CustomCaches {
	return &CustomCaches{
		guilds: gache.New[snowflake.ID, discord.Guild](time.Duration(time.Hour * 12)),
	}
}

func (c *CustomCaches) Guild(id snowflake.ID) (discord.Guild, bool) {
	g, ok := c.guilds.Get(id)

	return g, ok
}

func (c *CustomCaches) SaveGuild(id snowflake.ID, value discord.Guild) {
	c.guilds.Set(id, value, time.Hour*6)
}
