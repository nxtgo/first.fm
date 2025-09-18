package lfm

import (
	"time"

	"go.fm/lfm/types"
)

type trackApi struct {
	api *LastFMApi
}

func (t *trackApi) GetInfo(args P) (*types.TrackGetInfo, error) {
	key := generateCacheKey("track", args)

	if cached, ok := t.api.cache.Track.Get(key); ok {
		return &cached, nil
	}

	var result types.TrackGetInfo
	if err := t.api.doAndDecode("track.getinfo", args, &result); err != nil {
		return nil, err
	}

	ttl := t.getAdaptiveTTL(args)
	t.api.cache.Track.Set(key, result, ttl)

	return &result, nil
}

func (t *trackApi) getAdaptiveTTL(args P) time.Duration {
	if _, hasUser := args["username"]; hasUser {
		return 4 * time.Hour
	}
	return 12 * time.Hour
}
