package lfm

import (
	"fmt"

	"go.fm/lfm/types"
)

type trackApi struct {
	api *LastFMApi
}

// track.getInfo
func (t *trackApi) GetInfo(args P) (*types.TrackGetInfo, error) {
	var key string
	if mbid, ok := args["mbid"].(string); ok && mbid != "" {
		key = "mbid:" + mbid
	} else {
		artist, _ := args["artist"].(string)
		track, _ := args["track"].(string)
		username, _ := args["username"].(string)
		key = fmt.Sprintf("%s|%s|%s", artist, track, username)
	}

	if cached, ok := t.api.cache.Track.Get(key); ok {
		return &cached, nil
	}

	req := t.api.baseRequest("track.getinfo", args)
	data, err := req.Bytes()
	if err != nil {
		return nil, err
	}

	var result types.TrackGetInfo
	if err := decodeResponse(data, &result); err != nil {
		return nil, err
	}

	t.api.cache.Track.Set(key, result, 0)

	return &result, nil
}
