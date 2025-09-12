package cache

import (
	"time"

	"go.fm/types/lastfm"
)

func (c *LastFMCache) GetTopArtists(key string) (*lastfm.TopArtistsResponse, bool) {
	return c.topArtists.Get(key)
}

func (c *LastFMCache) SetTopArtists(key string, val *lastfm.TopArtistsResponse, ttl time.Duration) {
	c.topArtists.Set(key, val, ttl)
}

func (c *LastFMCache) GetTopAlbums(key string) (*lastfm.TopAlbumsResponse, bool) {
	return c.topAlbums.Get(key)
}

func (c *LastFMCache) SetTopAlbums(key string, val *lastfm.TopAlbumsResponse, ttl time.Duration) {
	c.topAlbums.Set(key, val, ttl)
}

func (c *LastFMCache) GetTopTracks(key string) (*lastfm.TopTracksResponse, bool) {
	return c.topTracks.Get(key)
}

func (c *LastFMCache) SetTopTracks(key string, val *lastfm.TopTracksResponse, ttl time.Duration) {
	c.topTracks.Set(key, val, ttl)
}
