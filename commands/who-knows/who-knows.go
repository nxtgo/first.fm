// this command represents the future command structure for all
// go.fm commands, heavily wip.

package whoknows

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"go.fm/constants"
	"go.fm/lfm"
	"go.fm/lfm/types"
	"go.fm/types/cmd"
	"go.fm/utils/image"
)

type Command struct{}

var (
	maxLimit int = 100
	minLimit int = 3
)

type PlayResult struct {
	UserID    string
	Username  string
	PlayCount int
}

type QueryInfo struct {
	Type       string
	Name       string
	ArtistName string
	Thumbnail  string
	BetterName string
}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "who-knows",
		Description: "see who has listened to a track/artist/album the most",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
		},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "type",
				Description: "artist, track or album",
				Choices: []discord.ApplicationCommandOptionChoiceString{
					{Name: "artist", Value: "artist"},
					{Name: "track", Value: "track"},
					{Name: "album", Value: "album"},
				},
				Required: true,
			},
			discord.ApplicationCommandOptionString{
				Name:        "name",
				Description: "name of the artist/track/album",
				Required:    false,
			},
			discord.ApplicationCommandOptionInt{
				Name:        "limit",
				Description: "max entries for the list (max: 100, min: 3)",
				Required:    false,
				MinValue:    &minLimit,
				MaxValue:    &maxLimit,
			},
			discord.ApplicationCommandOptionBool{
				Name:        "global",
				Description: "show global stats across all registered users instead of just this guild",
				Required:    false,
			},
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := ctx.Reply(e)
	if err := reply.Defer(); err != nil {
		ctx.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	options := parseCommandOptions(e)

	queryInfo, err := resolveQueryInfo(options, e, ctx)
	if err != nil {
		ctx.Error(e, err.Error())
		return
	}

	users, err := getUserList(options.IsGlobal, e, ctx)
	if err != nil {
		ctx.Error(e, constants.ErrorUnexpected)
		return
	}

	results := fetchPlayCounts(queryInfo, users, options.Limit, ctx)
	if len(results) == 0 {
		ctx.Error(e, constants.ErrorNoListeners)
		return
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].PlayCount > results[j].PlayCount
	})

	sendWhoKnowsResponse(e, reply, queryInfo, results, options)
}

type CommandOptions struct {
	Type     string
	Name     string
	Limit    int
	IsGlobal bool
}

func parseCommandOptions(e *events.ApplicationCommandInteractionCreate) CommandOptions {
	data := e.SlashCommandInteractionData()

	limit := 10
	if l, defined := data.OptInt("limit"); defined {
		limit = l
	}

	isGlobal, _ := data.OptBool("global")
	name, _ := data.OptString("name")

	return CommandOptions{
		Type:     data.String("type"),
		Name:     name,
		Limit:    limit,
		IsGlobal: isGlobal,
	}
}

func resolveQueryInfo(options CommandOptions, e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) (*QueryInfo, error) {
	queryInfo := &QueryInfo{Type: options.Type}

	if options.Name != "" {
		queryInfo.Name = options.Name
	} else {
		current, err := getCurrentTrack(e, ctx)
		if err != nil {
			return nil, err
		}
		currentTrack := current.Tracks[0]

		switch options.Type {
		case "artist":
			queryInfo.Name = currentTrack.Artist.Name
		case "track":
			queryInfo.Name = currentTrack.Name
			queryInfo.ArtistName = currentTrack.Artist.Name
		case "album":
			queryInfo.Name = currentTrack.Album.Name
			queryInfo.ArtistName = currentTrack.Artist.Name
		}
	}

	enrichQueryInfo(queryInfo, ctx)

	return queryInfo, nil
}

func getCurrentTrack(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) (*types.UserGetRecentTracks, error) {
	currentUser, err := ctx.Database.GetUser(ctx.Context, e.Member().User.ID.String())
	if err != nil {
		return nil, fmt.Errorf(constants.ErrorGetUser)
	}

	tracks, err := ctx.LastFM.User.GetRecentTracks(lfm.P{"user": currentUser, "limit": 1})
	if err != nil || len(tracks.Tracks) == 0 || tracks.Tracks[0].NowPlaying != "true" {
		return nil, fmt.Errorf(constants.ErrorFetchCurrentTrack)
	}

	return tracks, nil
}

func enrichQueryInfo(queryInfo *QueryInfo, ctx cmd.CommandContext) {
	queryInfo.BetterName = queryInfo.Name
	queryInfo.Thumbnail = "https://lastfm.freetls.fastly.net/i/u/avatar170s/818148bf682d429dc215c1705eb27b98.png"

	switch queryInfo.Type {
	case "artist":
		if artist, err := ctx.LastFM.Artist.GetInfo(lfm.P{"artist": queryInfo.Name}); err == nil {
			if len(artist.Images) > 0 {
				queryInfo.Thumbnail = artist.Images[len(artist.Images)-1].Url
			}
			if artist.Name != "" {
				queryInfo.BetterName = artist.Name
			}
		}

	case "track":
		params := lfm.P{"track": queryInfo.Name}
		if queryInfo.ArtistName != "" {
			params["artist"] = queryInfo.ArtistName
		}
		if track, err := ctx.LastFM.Track.GetInfo(params); err == nil {
			if len(track.Album.Images) > 0 {
				queryInfo.Thumbnail = track.Album.Images[len(track.Album.Images)-1].Url
			}
			if track.Name != "" {
				queryInfo.BetterName = track.Name
			}
		}

	case "album":
		params := lfm.P{"album": queryInfo.Name}
		if queryInfo.ArtistName != "" {
			params["artist"] = queryInfo.ArtistName
		}
		if album, err := ctx.LastFM.Album.GetInfo(params); err == nil {
			if len(album.Images) > 0 {
				queryInfo.Thumbnail = album.Images[len(album.Images)-1].Url
			}
			if album.Name != "" {
				queryInfo.BetterName = album.Name
			}
		}
	}
}

func getUserList(isGlobal bool, e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) (map[snowflake.ID]string, error) {
	if isGlobal {
		return getAllRegisteredUsers(ctx)
	}
	return ctx.LastFM.User.GetUsersByGuild(ctx.Context, e, ctx.Database)
}

func getAllRegisteredUsers(ctx cmd.CommandContext) (map[snowflake.ID]string, error) {
	if cached, ok := ctx.Cache.Members.Get(snowflake.ID(0)); ok {
		return cached, nil
	}

	users, err := ctx.Database.ListUsers(ctx.Context)
	if err != nil {
		return nil, err
	}

	result := make(map[snowflake.ID]string, len(users))
	for _, user := range users {
		if id, err := snowflake.Parse(user.DiscordID); err == nil {
			result[id] = user.LastfmUsername
		}
	}

	ctx.Cache.Members.Set(snowflake.ID(0), result, 5*time.Minute)

	return result, nil
}

func fetchPlayCounts(queryInfo *QueryInfo, users map[snowflake.ID]string, maxWorkers int, ctx cmd.CommandContext) []PlayResult {
	if len(users) == 0 {
		return nil
	}

	workerLimit := min(maxWorkers, 20)
	sem := make(chan struct{}, workerLimit)

	var (
		results []PlayResult
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	ctx_timeout, cancel := context.WithTimeout(ctx.Context, 30*time.Second)
	defer cancel()

	for userID, username := range users {
		select {
		case <-ctx_timeout.Done():
			goto done
		default:
		}

		wg.Add(1)
		go func(id snowflake.ID, user string) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx_timeout.Done():
				return
			}

			count := fetchUserPlayCount(queryInfo, user, ctx)
			if count > 0 {
				mu.Lock()
				results = append(results, PlayResult{
					UserID:    id.String(),
					Username:  user,
					PlayCount: count,
				})
				mu.Unlock()
			}
		}(userID, username)
	}

done:
	wg.Wait()
	return results
}

func fetchUserPlayCount(queryInfo *QueryInfo, username string, ctx cmd.CommandContext) int {
	params := lfm.P{
		"user": username,
		"name": queryInfo.Name,
		"type": queryInfo.Type,
	}

	if queryInfo.ArtistName != "" {
		params["artist"] = queryInfo.ArtistName
	}

	count, err := ctx.LastFM.User.GetPlays(params)
	if err != nil {
		return 0
	}
	return count
}

func sendWhoKnowsResponse(e *events.ApplicationCommandInteractionCreate, reply *cmd.ResponseBuilder, queryInfo *QueryInfo, results []PlayResult, options CommandOptions) {
	scope := "in this server"
	if options.IsGlobal {
		scope = "globally"
	} else {
		guild, ok := e.Guild()
		if ok {
			scope = fmt.Sprintf("in %s", guild.Name)
		}
	}

	title := fmt.Sprintf("### Who knows %s **%s** %s?", queryInfo.Type, queryInfo.BetterName, scope)

	list := buildResultsList(results, options.Limit)

	color := 0x00ADD8
	if dominantColor, err := image.DominantColor(queryInfo.Thumbnail); err == nil {
		color = dominantColor
	}

	component := discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplay(title),
			discord.NewTextDisplay(list),
		).WithAccessory(discord.NewThumbnail(queryInfo.Thumbnail)),
	).WithAccentColor(color)

	reply.Flags(discord.MessageFlagIsComponentsV2).Component(component).Edit()
}

func buildResultsList(results []PlayResult, limit int) string {
	if len(results) == 0 {
		return "No listeners found."
	}

	list := ""
	displayLimit := min(len(results), limit)

	for i := range displayLimit {
		r := results[i]
		list += fmt.Sprintf(
			"%d. [%s](<https://www.last.fm/user/%s>) (*<@%s>*) — **%d** plays\n",
			i+1, r.Username, r.Username, r.UserID, r.PlayCount,
		)
	}

	if len(results) > limit {
		list += fmt.Sprintf("\n*...and %d more listeners*", len(results)-limit)
	}

	return list
}
