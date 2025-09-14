package lfm

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"go.fm/db"
	"go.fm/lfm/types"
)

type userApi struct {
	api *LastFMApi
}

func (u *userApi) GetPlays(args P) (int, error) {
	username := args["user"].(string)
	queryType := args["type"].(string)
	queryName := args["name"].(string)

	cacheKey := fmt.Sprintf("%s:%s", username, queryName)
	if cached, ok := u.api.cache.Plays.Get(cacheKey); ok {
		return cached, nil
	}

	var playCount int
	var err error

	switch queryType {
	case "artist":
		var artist *types.ArtistGetInfo
		if username != "" {
			artist, err = u.api.Artist.GetInfo(P{"artist": queryName, "username": username})
		} else {
			artist, err = u.api.Artist.GetInfo(P{"artist": queryName})
		}
		if err != nil {
			return 0, err
		}
		if username != "" {
			playCount = artist.Stats.UserPlayCount
		} else {
			fmt.Sscanf(fmt.Sprint(artist.Stats.PlayCount), "%d", &playCount)
		}

	case "album":
		var album *types.AlbumGetInfo
		if username != "" {
			album, err = u.api.Album.GetInfo(P{"artist": args["artist"], "album": queryName, "username": username})
		} else {
			album, err = u.api.Album.GetInfo(P{"artist": args["artist"], "album": queryName})
		}
		if err != nil {
			return 0, err
		}
		if username != "" {
			playCount = album.UserPlayCount
		} else {
			fmt.Sscanf(fmt.Sprint(album.PlayCount), "%d", &playCount)
		}

	case "track":
		var track *types.TrackGetInfo
		if username != "" {
			track, err = u.api.Track.GetInfo(P{"artist": args["artist"], "track": queryName, "username": username})
		} else {
			track, err = u.api.Track.GetInfo(P{"artist": args["artist"], "track": queryName})
		}
		if err != nil {
			return 0, err
		}
		if username != "" {
			playCount = track.UserPlayCount
		} else {
			fmt.Sscanf(fmt.Sprint(track.PlayCount), "%d", &playCount)
		}

	default:
		return 0, fmt.Errorf("unknown query type: %s", queryType)
	}

	u.api.cache.Plays.Set(cacheKey, playCount, 0)
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

	u.api.cache.Members.Set(guildID, users, 0)

	return users, nil
}

// user.getInfo
func (u *userApi) GetInfo(args P) (*types.UserGetInfo, error) {
	username := args["user"].(string)
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

// user.getRecentTracks
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

// user.getTopAlbums
func (u *userApi) GetTopAlbums(args P) (*types.UserGetTopAlbums, error) {
	username := args["user"].(string)

	if albums, ok := u.api.cache.TopAlbums.Get(username); ok {
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

	u.api.cache.TopAlbums.Set(username, result, 0)
	return &result, nil
}

// user.getTopArtists
func (u *userApi) GetTopArtists(args P) (*types.UserGetTopArtists, error) {
	username := args["user"].(string)

	if artists, ok := u.api.cache.TopArtists.Get(username); ok {
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

	u.api.cache.TopArtists.Set(username, result, 0)
	return &result, nil
}

// user.getTopTracks
func (u *userApi) GetTopTracks(args P) (*types.UserGetTopTracks, error) {
	username := args["user"].(string)

	if tracks, ok := u.api.cache.TopTracks.Get(username); ok {
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

	u.api.cache.TopTracks.Set(username, result, 0)

	return &result, nil
}
