package lastfm

import "fmt"

func (c *Client) GetRecentTracks(user string, limit int) (*RecentTracksResponse, error) {
	params := map[string]string{
		"user":  user,
		"limit": fmt.Sprint(limit),
	}
	var resp RecentTracksResponse
	err := c.req("user.getRecentTracks", params, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
