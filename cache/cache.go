package cache

import (
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/nxtgo/gce"

	"go.fm/lfm/types"
)

type Cache struct {
	User       *gce.Cache[string, types.UserGetInfo]
	Members    *gce.Cache[snowflake.ID, map[snowflake.ID]string]
	Album      *gce.Cache[string, types.AlbumGetInfo]
	Artist     *gce.Cache[string, types.ArtistGetInfo]
	Track      *gce.Cache[string, types.TrackGetInfo]
	TopAlbums  *gce.Cache[string, types.UserGetTopAlbums]
	TopArtists *gce.Cache[string, types.UserGetTopArtists]
	TopTracks  *gce.Cache[string, types.UserGetTopTracks]
	Tracks     *gce.Cache[string, types.UserGetRecentTracks]
	Plays      *gce.Cache[string, int]
}

func NewCache() *Cache {
	return &Cache{
		User: gce.New[string, types.UserGetInfo](
			gce.WithDefaultTTL(time.Minute),
			gce.WithMaxEntries(50_000),
		),
		Members: gce.New[snowflake.ID, map[snowflake.ID]string](
			gce.WithDefaultTTL(time.Minute*10),
			gce.WithMaxEntries(2000),
		),
		Album: gce.New[string, types.AlbumGetInfo](
			gce.WithDefaultTTL(time.Hour*12),
			gce.WithMaxEntries(64_000),
		),
		Artist: gce.New[string, types.ArtistGetInfo](
			gce.WithDefaultTTL(time.Hour*12),
			gce.WithMaxEntries(64_000),
		),
		Track: gce.New[string, types.TrackGetInfo](
			gce.WithDefaultTTL(time.Hour*12),
			gce.WithMaxEntries(64_000),
		),
		TopAlbums: gce.New[string, types.UserGetTopAlbums](
			gce.WithDefaultTTL(time.Minute*15),
			gce.WithMaxEntries(1000),
		),
		TopArtists: gce.New[string, types.UserGetTopArtists](
			gce.WithDefaultTTL(time.Minute*15),
			gce.WithMaxEntries(1000),
		),
		TopTracks: gce.New[string, types.UserGetTopTracks](
			gce.WithDefaultTTL(time.Minute*15),
			gce.WithMaxEntries(1000),
		),
		Tracks: gce.New[string, types.UserGetRecentTracks](
			gce.WithDefaultTTL(time.Minute*15),
			gce.WithMaxEntries(1000),
		),
		Plays: gce.New[string, int](
			gce.WithDefaultTTL(time.Minute*15),
			gce.WithMaxEntries(50_000),
		),
	}
}

func (c *Cache) StatsString() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%-10s %-8s %-8s %-8s %-10s %-6s\n",
		"cache", "hits", "misses", "loads", "evictions", "size")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 53))

	writeStats := func(name string, s gce.Stats) {
		fmt.Fprintf(&sb, "%-10s %-8d %-8d %-8d %-10d %-6d\n",
			name, s.Hits, s.Misses, s.Loads, s.Evictions, s.CurrentSize)
	}

	if c.User != nil {
		writeStats("User", c.User.Stats())
	}
	if c.Members != nil {
		writeStats("Members", c.Members.Stats())
	}
	if c.Album != nil {
		writeStats("Album", c.Album.Stats())
	}
	if c.Artist != nil {
		writeStats("Artist", c.Artist.Stats())
	}
	if c.Track != nil {
		writeStats("Track", c.Track.Stats())
	}
	if c.TopAlbums != nil {
		writeStats("TopAlbums", c.TopAlbums.Stats())
	}
	if c.TopArtists != nil {
		writeStats("TopArtists", c.TopArtists.Stats())
	}
	if c.TopTracks != nil {
		writeStats("TopTracks", c.TopTracks.Stats())
	}
	if c.Tracks != nil {
		writeStats("Tracks", c.Tracks.Stats())
	}
	if c.Plays != nil {
		writeStats("Plays", c.Plays.Stats())
	}

	return sb.String()
}

func (c *Cache) Close() {
	if c.User != nil {
		c.User.Close()
	}
	if c.Members != nil {
		c.Members.Close()
	}
	if c.Album != nil {
		c.Album.Close()
	}
	if c.Artist != nil {
		c.Artist.Close()
	}
	if c.Track != nil {
		c.Track.Close()
	}
	if c.TopAlbums != nil {
		c.TopAlbums.Close()
	}
	if c.TopArtists != nil {
		c.TopArtists.Close()
	}
	if c.TopTracks != nil {
		c.TopTracks.Close()
	}
	if c.Tracks != nil {
		c.Tracks.Close()
	}
	if c.Plays != nil {
		c.Plays.Close()
	}
}
