package lfm

import (
	"time"

	"go.fm/lfm/types"
)

type albumApi struct {
	api *LastFMApi
}

func (a *albumApi) GetInfo(args P) (*types.AlbumGetInfo, error) {
	key := generateCacheKey("album", args)

	if cached, ok := a.api.cache.Album.Get(key); ok {
		return &cached, nil
	}

	var result types.AlbumGetInfo
	if err := a.api.doAndDecode("album.getinfo", args, &result); err != nil {
		return nil, err
	}

	ttl := a.getAdaptiveTTL(args)
	a.api.cache.Album.Set(key, result, ttl)

	return &result, nil
}

func (a *albumApi) getAdaptiveTTL(args P) time.Duration {
	if _, hasUser := args["username"]; hasUser {
		return 6 * time.Hour
	}
	return 24 * time.Hour
}
