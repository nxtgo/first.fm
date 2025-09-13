package lfm

import (
	"fmt"

	"go.fm/lfm/types"
)

type artistApi struct {
	api *LastFMApi
}

// artist.getInfo
func (a *artistApi) GetInfo(args P) (*types.ArtistGetInfo, error) {
	var key string
	if mbid, ok := args["mbid"].(string); ok && mbid != "" {
		key = "mbid:" + mbid
	} else {
		artist, _ := args["artist"].(string)
		username, _ := args["username"].(string)
		key = fmt.Sprintf("%s|%s", artist, username)
	}

	if cached, ok := a.api.cache.Artist.Get(key); ok {
		return &cached, nil
	}

	req := a.api.baseRequest("artist.getinfo", args)
	data, err := req.Bytes()
	if err != nil {
		return nil, err
	}

	var result types.ArtistGetInfo
	if err := decodeResponse(data, &result); err != nil {
		return nil, err
	}

	a.api.cache.Artist.Set(key, result, 0)

	return &result, nil
}
