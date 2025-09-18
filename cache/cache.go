package cache

import (
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
	Cover      *gce.Cache[string, string]
}

func NewCache() *Cache {
	return &Cache{
		User: gce.New[string, types.UserGetInfo](
			gce.WithDefaultTTL(time.Hour),
			gce.WithMaxEntries(50_000),
		),
		Members: gce.New[snowflake.ID, map[snowflake.ID]string](
			gce.WithDefaultTTL(time.Hour*6),
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
		Cover: gce.New[string, string](
			gce.WithDefaultTTL(time.Hour*24*365),
			gce.WithMaxEntries(100_000),
		),
	}
}

type CacheStats struct {
	Name  string
	Stats gce.Stats
}

func (c *Cache) Stats() []CacheStats {
	var out []CacheStats

	add := func(name string, s gce.Stats) {
		out = append(out, CacheStats{Name: name, Stats: s})
	}

	if c.User != nil {
		add("User", c.User.Stats())
	}
	if c.Members != nil {
		add("Members", c.Members.Stats())
	}
	if c.Album != nil {
		add("Album", c.Album.Stats())
	}
	if c.Artist != nil {
		add("Artist", c.Artist.Stats())
	}
	if c.Track != nil {
		add("Track", c.Track.Stats())
	}
	if c.TopAlbums != nil {
		add("TopAlbums", c.TopAlbums.Stats())
	}
	if c.TopArtists != nil {
		add("TopArtists", c.TopArtists.Stats())
	}
	if c.TopTracks != nil {
		add("TopTracks", c.TopTracks.Stats())
	}
	if c.Tracks != nil {
		add("Tracks", c.Tracks.Stats())
	}
	if c.Plays != nil {
		add("Plays", c.Plays.Stats())
	}
	if c.Plays != nil {
		add("Cover", c.Cover.Stats())
	}

	return out
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
	if c.Cover != nil {
		c.Cover.Close()
	}
}
