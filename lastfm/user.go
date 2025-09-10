package lastfm

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"go.fm/db"
)

func GetUsersByGuild(
	ctx context.Context,
	e *events.ApplicationCommandInteractionCreate,
	q *db.Queries,
) (map[string]string, error) {
	registered, err := q.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	memberIDs := make(map[snowflake.ID]struct{})
	e.Client().Caches().MembersForEach(*e.GuildID(), func(m discord.Member) {
		memberIDs[m.User.ID] = struct{}{}
	})

	users := make(map[string]string)
	for _, u := range registered {
		if _, ok := memberIDs[snowflake.MustParse(u.DiscordID)]; ok {
			users[u.DiscordID] = u.LastfmUsername
		}
	}

	return users, nil
}

func (c *Client) GetUserInfo(user string) (*UserInfoResponse, error) {
	params := map[string]string{
		"user": user,
	}
	var resp UserInfoResponse
	err := c.req("user.getInfo", params, &resp)
	if err != nil || resp.User.Name == "" {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetRecentTracks(user string, limit int) (*RecentTracksResponse, error) {
	params := map[string]string{
		"user":  user,
		"limit": fmt.Sprint(limit),
	}
	var resp RecentTracksResponse
	err := c.req("user.getRecentTracks", params, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetUserPlays(user, queryType, queryName string, limit int) (int, error) {
	params := map[string]string{
		"user":  user,
		"limit": fmt.Sprint(limit),
	}

	var playCount int

	switch queryType {
	case "artist":
		var resp TopArtistsResponse
		err := c.req("user.getTopArtists", params, &resp)
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
		var resp TopAlbumsResponse
		err := c.req("user.getTopAlbums", params, &resp)
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
		var resp TopTracksResponse
		err := c.req("user.getTopTracks", params, &resp)
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

	return playCount, nil
}

func (c *Client) WhoKnows(guildID string, queryType, queryName string, users map[string]string) ([]WhoKnowsResult, error) {
	results := make([]WhoKnowsResult, 0)

	for discordID, lastFMUser := range users {
		plays, err := c.GetUserPlays(lastFMUser, queryType, queryName, 50)
		if err != nil {
			continue
		}
		if plays > 0 {
			results = append(results, WhoKnowsResult{
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
