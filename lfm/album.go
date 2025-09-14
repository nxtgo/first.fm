package lfm

import (
	"fmt"

	"go.fm/lfm/types"
)

type albumApi struct {
	api *LastFMApi
}

// album.getInfo
func (a *albumApi) GetInfo(args P) (*types.AlbumGetInfo, error) {
	var key string
	if mbid, ok := args["mbid"].(string); ok && mbid != "" {
		key = "mbid:" + mbid
	} else {
		artist, _ := args["artist"].(string)
		album, _ := args["album"].(string)
		key = fmt.Sprintf("%s|%s", artist, album)
	}

	if cached, ok := a.api.cache.Album.Get(key); ok {
		return &cached, nil
	}

	req := a.api.baseRequest("album.getinfo", args)
	data, err := req.Bytes()
	if err != nil {
		return nil, err
	}

	var result types.AlbumGetInfo
	if err := decodeResponse(data, &result); err != nil {
		return nil, err
	}

	a.api.cache.Album.Set(key, result, 0)

	return &result, nil
}
