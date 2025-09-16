package lfm

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"go.fm/db"
	"go.fm/lfm/types"
)

type userApi struct {
	api *LastFMApi
}

func (u *userApi) GetInfo(args P) (*types.UserGetInfo, error) {
	username := args["user"].(string)

	// Check if we have cached data
	if user, ok := u.api.cache.User.Get(username); ok {
		return &user, nil
	}

	req := u.api.baseRequest("user.getinfo", args)
	data, err := req.Bytes()
	if err != nil {
		return nil, err
	}

	var result types.UserGetInfo
	if err := decodeResponse(data, &result); err != nil {
		return nil, err
	}

	u.api.cache.User.Set(username, result, 0)

	return &result, nil
}

func (u *userApi) GetRecentTracks(args P) (*types.UserGetRecentTracks, error) {
	req := u.api.baseRequest("user.getrecenttracks", args)
	data, err := req.Bytes()
	if err != nil {
		return nil, err
	}

	var result types.UserGetRecentTracks
	if err := decodeResponse(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (u *userApi) GetTopArtists(args P) (*types.UserGetTopArtists, error) {
	key := generateCacheKey("topartists", args)

	if artists, ok := u.api.cache.TopArtists.Get(key); ok {
		return &artists, nil
	}

	req := u.api.baseRequest("user.gettopartists", args)
	data, err := req.Bytes()
	if err != nil {
		return nil, err
	}

	var result types.UserGetTopArtists
	if err := decodeResponse(data, &result); err != nil {
		return nil, err
	}

	ttl := u.getTopDataTTL(args)
	u.api.cache.TopArtists.Set(key, result, ttl)

	return &result, nil
}

func (u *userApi) GetTopAlbums(args P) (*types.UserGetTopAlbums, error) {
	key := generateCacheKey("topalbums", args)

	if albums, ok := u.api.cache.TopAlbums.Get(key); ok {
		return &albums, nil
	}

	req := u.api.baseRequest("user.gettopalbums", args)
	data, err := req.Bytes()
	if err != nil {
		return nil, err
	}

	var result types.UserGetTopAlbums
	if err := decodeResponse(data, &result); err != nil {
		return nil, err
	}

	ttl := u.getTopDataTTL(args)
	u.api.cache.TopAlbums.Set(key, result, ttl)

	return &result, nil
}

func (u *userApi) GetTopTracks(args P) (*types.UserGetTopTracks, error) {
	key := generateCacheKey("toptracks", args)

	if tracks, ok := u.api.cache.TopTracks.Get(key); ok {
		return &tracks, nil
	}

	req := u.api.baseRequest("user.gettoptracks", args)
	data, err := req.Bytes()
	if err != nil {
		return nil, err
	}

	var result types.UserGetTopTracks
	if err := decodeResponse(data, &result); err != nil {
		return nil, err
	}

	ttl := u.getTopDataTTL(args)
	u.api.cache.TopTracks.Set(key, result, ttl)

	return &result, nil
}

func (u *userApi) GetPlays(args P) (int, error) {
	key := generateCacheKey("plays", args)

	if cached, ok := u.api.cache.Plays.Get(key); ok {
		return cached, nil
	}

	username := args["user"].(string)
	queryType := args["type"].(string)
	queryName := args["name"].(string)

	var playCount int
	var cacheTTL time.Duration

	switch queryType {
	case "artist":
		artist, err := u.api.Artist.GetInfo(P{"artist": queryName, "username": username})
		if err != nil {
			return 0, err
		}
		playCount = artist.Stats.UserPlayCount
		cacheTTL = 10 * time.Minute

	case "album":
		album, err := u.api.Album.GetInfo(P{
			"artist":   args["artist"],
			"album":    queryName,
			"username": username,
		})
		if err != nil {
			return 0, err
		}
		playCount = album.UserPlayCount
		cacheTTL = 15 * time.Minute

	case "track":
		track, err := u.api.Track.GetInfo(P{
			"artist":   args["artist"],
			"track":    queryName,
			"username": username,
		})
		if err != nil {
			return 0, err
		}
		playCount = track.UserPlayCount
		cacheTTL = 5 * time.Minute

	default:
		return 0, fmt.Errorf("unknown query type: %s", queryType)
	}

	u.api.cache.Plays.Set(key, playCount, cacheTTL)
	return playCount, nil
}

func (u *userApi) GetUsersByGuild(
	ctx context.Context,
	e *events.ApplicationCommandInteractionCreate,
	q *db.Queries,
) (map[snowflake.ID]string, error) {
	guildID := *e.GuildID()

	if cached, ok := u.api.cache.Members.Get(guildID); ok {
		return cached, nil
	}

	registered, err := q.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	memberIDs := make(map[snowflake.ID]struct{})
	cached := e.Client().Caches.Members(guildID)
	for m := range cached {
		memberIDs[m.User.ID] = struct{}{}
	}

	users := make(map[snowflake.ID]string)
	for _, u := range registered {
		id := snowflake.MustParse(u.DiscordID)
		if _, ok := memberIDs[id]; ok {
			users[id] = u.LastfmUsername
		}
	}

	u.api.cache.Members.Set(guildID, users, 10*time.Minute)

	return users, nil
}

func (u *userApi) GetInfoWithPrefetch(args P) (*types.UserGetInfo, error) {
	username := args["user"].(string)

	userInfo, err := u.GetInfo(args)
	if err != nil {
		return nil, err
	}

	go u.PrefetchUserData(username)

	return userInfo, nil
}

func (u *userApi) PrefetchUserData(username string) {
	if _, ok := u.api.cache.TopArtists.Get(generateCacheKey("topartists", P{"user": username})); !ok {
		u.GetTopArtists(P{"user": username, "limit": 10})
	}

	if _, ok := u.api.cache.TopAlbums.Get(generateCacheKey("topalbums", P{"user": username})); !ok {
		u.GetTopAlbums(P{"user": username, "limit": 10})
	}

	if _, ok := u.api.cache.TopTracks.Get(generateCacheKey("toptracks", P{"user": username})); !ok {
		u.GetTopTracks(P{"user": username, "limit": 10})
	}
}

func (u *userApi) InvalidateUserCache(username string) {
	periods := []string{"7day", "1month", "3month", "6month", "12month", "overall"}

	for _, period := range periods {
		key := generateCacheKey("topartists", P{"user": username, "period": period})
		u.api.cache.TopArtists.Delete(key)

		key = generateCacheKey("topalbums", P{"user": username, "period": period})
		u.api.cache.TopAlbums.Delete(key)

		key = generateCacheKey("toptracks", P{"user": username, "period": period})
		u.api.cache.TopTracks.Delete(key)
	}

	defaultKeys := []P{
		{"user": username},
		{"user": username, "limit": 10},
		{"user": username, "limit": 50},
	}

	for _, args := range defaultKeys {
		u.api.cache.TopArtists.Delete(generateCacheKey("topartists", args))
		u.api.cache.TopAlbums.Delete(generateCacheKey("topalbums", args))
		u.api.cache.TopTracks.Delete(generateCacheKey("toptracks", args))
	}

	u.api.cache.User.Delete(username)
}

func (u *userApi) getTopDataTTL(args P) time.Duration {
	if period, ok := args["period"].(string); ok {
		switch period {
		case "7day":
			return 30 * time.Minute
		case "1month":
			return 2 * time.Hour
		case "3month", "6month":
			return 6 * time.Hour
		case "12month", "overall":
			return 12 * time.Hour
		}
	}

	return 15 * time.Minute
}
