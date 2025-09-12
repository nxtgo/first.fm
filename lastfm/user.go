package lastfm

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"go.fm/db"
	"go.fm/types/lastfm"
)

func (c *Client) GetUsersByGuild(
	ctx context.Context,
	e *events.ApplicationCommandInteractionCreate,
	q *db.Queries,
) (map[snowflake.ID]string, error) {
	guildID := *e.GuildID()

	if cached, ok := c.cache.GetMembers(guildID); ok {
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

	c.cache.SetMembers(guildID, users, 0)

	return users, nil
}

func (c *Client) GetUserInfo(user string) (*lastfm.UserInfoResponse, error) {
	if cached, ok := c.cache.GetUser(user); ok {
		return cached, nil
	}

	params := map[string]string{
		"user": user,
	}
	var resp lastfm.UserInfoResponse
	err := c.req("user.getInfo", params, &resp)
	if err != nil || resp.User.Name == "" {
		return nil, err
	}

	c.cache.SetUser(user, &resp, time.Minute*30)

	return &resp, nil
}

func (c *Client) GetRecentTracks(user string, limit int) (*lastfm.RecentTracksResponse, error) {
	params := map[string]string{
		"user":  user,
		"limit": fmt.Sprint(limit),
	}
	var resp lastfm.RecentTracksResponse
	if err := c.req("user.getRecentTracks", params, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetUserPlays(user, queryType, queryName string, limit int) (int, error) {
	cacheKey := fmt.Sprintf("%s:%s", user, queryName)

	if cached, ok := c.cache.GetPlays(cacheKey); ok {
		return cached, nil
	}

	var playCount int

	switch queryType {
	case "artist":
		resp, err := c.GetTopArtists(user, limit)
		if err != nil {
			return 0, err
		}
		for _, a := range resp.TopArtists.Artist {
			if strings.EqualFold(a.Name, queryName) {
				fmt.Sscanf(a.Playcount, "%d", &playCount)
				break
			}
		}
	case "album":
		resp, err := c.GetTopAlbums(user, limit)
		if err != nil {
			return 0, err
		}
		for _, a := range resp.TopAlbums.Album {
			if strings.EqualFold(a.Name, queryName) {
				fmt.Sscanf(a.Playcount, "%d", &playCount)
				break
			}
		}
	case "track":
		resp, err := c.GetTopTracks(user, limit)
		if err != nil {
			return 0, err
		}
		for _, t := range resp.TopTracks.Track {
			if strings.EqualFold(t.Name, queryName) {
				fmt.Sscanf(t.Playcount, "%d", &playCount)
				break
			}
		}
	default:
		return 0, fmt.Errorf("unknown query type: %s", queryType)
	}

	c.cache.SetPlays(cacheKey, playCount, 0)

	return playCount, nil
}

func (c *Client) WhoKnows(guildID string, queryType, queryName string, users map[string]string) ([]lastfm.WhoKnowsResult, error) {
	results := make([]lastfm.WhoKnowsResult, 0)

	for discordID, lastFMUser := range users {
		plays, err := c.GetUserPlays(lastFMUser, queryType, queryName, 50)
		if err != nil {
			continue
		}
		if plays > 0 {
			results = append(results, lastfm.WhoKnowsResult{
				UserID:    discordID,
				Username:  "",
				PlayCount: plays,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].PlayCount > results[j].PlayCount
	})

	return results, nil
}
