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

	if user, ok := u.api.cache.User.Get(username); ok {
		return &user, nil
	}

	var result types.UserGetInfo
	if err := u.api.doAndDecode("user.getinfo", args, &result); err != nil {
		return nil, err
	}

	u.api.cache.User.Set(username, result, 0)
	return &result, nil
}

func (u *userApi) GetRecentTracks(args P) (*types.UserGetRecentTracks, error) {
	var result types.UserGetRecentTracks
	if err := u.api.doAndDecode("user.getrecenttracks", args, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (u *userApi) GetTopArtists(args P) (*types.UserGetTopArtists, error) {
	key := generateCacheKey("topartists", args)

	if artists, ok := u.api.cache.TopArtists.Get(key); ok {
		return &artists, nil
	}

	var result types.UserGetTopArtists
	if err := u.api.doAndDecode("user.gettopartists", args, &result); err != nil {
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

	var result types.UserGetTopAlbums
	if err := u.api.doAndDecode("user.gettopalbums", args, &result); err != nil {
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

	var result types.UserGetTopTracks
	if err := u.api.doAndDecode("user.gettoptracks", args, &result); err != nil {
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

	fetchers := map[string]func() (int, time.Duration, error){
		"artist": func() (int, time.Duration, error) {
			a, err := u.api.Artist.GetInfo(P{"artist": queryName, "username": username})
			if err != nil {
				return 0, 0, err
			}
			return a.Stats.UserPlayCount, 10 * time.Minute, nil
		},
		"album": func() (int, time.Duration, error) {
			a, err := u.api.Album.GetInfo(P{"artist": args["artist"], "album": queryName, "username": username})
			if err != nil {
				return 0, 0, err
			}
			return a.UserPlayCount, 15 * time.Minute, nil
		},
		"track": func() (int, time.Duration, error) {
			t, err := u.api.Track.GetInfo(P{"artist": args["artist"], "track": queryName, "username": username})
			if err != nil {
				return 0, 0, err
			}
			return t.UserPlayCount, 5 * time.Minute, nil
		},
	}

	fetch, ok := fetchers[queryType]
	if !ok {
		return 0, fmt.Errorf("unknown query type: %s", queryType)
	}

	count, ttl, err := fetch()
	if err != nil {
		return 0, err
	}

	u.api.cache.Plays.Set(key, count, ttl)
	return count, nil
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

	u.PrefetchUserData(username)

	return userInfo, nil
}

func (u *userApi) PrefetchUserData(username string) {
	if _, ok := u.api.cache.TopArtists.Get(generateCacheKey("topartists", P{"user": username})); !ok {
		go u.GetTopArtists(P{"user": username, "limit": 10})
	}

	if _, ok := u.api.cache.TopAlbums.Get(generateCacheKey("topalbums", P{"user": username})); !ok {
		go u.GetTopAlbums(P{"user": username, "limit": 10})
	}

	if _, ok := u.api.cache.TopTracks.Get(generateCacheKey("toptracks", P{"user": username})); !ok {
		go u.GetTopTracks(P{"user": username, "limit": 10})
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
