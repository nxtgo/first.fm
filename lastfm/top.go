package lastfm

import (
	"fmt"

	"go.fm/types/lastfm"
)

func (c *Client) GetTopArtists(user string, limit int) (*lastfm.TopArtistsResponse, error) {
	cacheKey := fmt.Sprintf("%s:%d", user, limit)
	if cached, ok := c.cache.GetTopArtists(cacheKey); ok {
		return cached, nil
	}

	params := map[string]string{
		"user":  user,
		"limit": fmt.Sprint(limit),
	}

	var resp lastfm.TopArtistsResponse
	if err := c.req("user.getTopArtists", params, &resp); err != nil {
		return nil, err
	}

	c.cache.SetTopArtists(cacheKey, &resp, 0)
	return &resp, nil
}

func (c *Client) GetTopAlbums(user string, limit int) (*lastfm.TopAlbumsResponse, error) {
	cacheKey := fmt.Sprintf("%s:%d", user, limit)
	if cached, ok := c.cache.GetTopAlbums(cacheKey); ok {
		return cached, nil
	}

	params := map[string]string{
		"user":  user,
		"limit": fmt.Sprint(limit),
	}

	var resp lastfm.TopAlbumsResponse
	if err := c.req("user.getTopAlbums", params, &resp); err != nil {
		return nil, err
	}

	c.cache.SetTopAlbums(cacheKey, &resp, 0)
	return &resp, nil
}

func (c *Client) GetTopTracks(user string, limit int) (*lastfm.TopTracksResponse, error) {
	cacheKey := fmt.Sprintf("%s:%d", user, limit)
	if cached, ok := c.cache.GetTopTracks(cacheKey); ok {
		return cached, nil
	}

	params := map[string]string{
		"user":  user,
		"limit": fmt.Sprint(limit),
	}

	var resp lastfm.TopTracksResponse
	if err := c.req("user.getTopTracks", params, &resp); err != nil {
		return nil, err
	}

	c.cache.SetTopTracks(cacheKey, &resp, 0)
	return &resp, nil
}
