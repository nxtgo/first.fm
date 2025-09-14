package lfm

import (
	"time"

	"go.fm/lfm/types"
)

type artistApi struct {
	api *LastFMApi
}

func (a *artistApi) GetInfo(args P) (*types.ArtistGetInfo, error) {
	key := generateCacheKey("artist", args)

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

	ttl := a.getAdaptiveTTL(args)
	a.api.cache.Artist.Set(key, result, ttl)

	return &result, nil
}

func (ar *artistApi) getAdaptiveTTL(args P) time.Duration {
	if _, hasUser := args["username"]; hasUser {
		return 6 * time.Hour
	}
	return 24 * time.Hour
}
