package whoknows

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"go.fm/lfm"
	"go.fm/pkg/bild/colors"
	"go.fm/pkg/constants/emojis"
	"go.fm/pkg/constants/errs"
	"go.fm/pkg/ctx"
	"go.fm/pkg/discord/reply"
)

var (
	maxLimit     = 100
	minLimit     = 3
	defaultLimit = 10
	maxWorkers   = 20
	timeout      = 30 * time.Second
)

type Command struct{}

type Options struct {
	Type     string
	Name     string
	Limit    int
	IsGlobal bool
}

type Query struct {
	Url        string
	Type       string
	Name       string
	ArtistName string
	Thumbnail  string
	BetterName string
}

type Result struct {
	UserID    string
	Username  string
	PlayCount int
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

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) {
	r := reply.New(e)
	if err := r.Defer(); err != nil {
		reply.Error(e, errs.ErrCommandDeferFailed)
		return
	}

	options := parseOptions(e)

	query, err := buildQuery(options, e, ctx)
	if err != nil {
		reply.Error(e, err)
		return
	}

	users, err := getUsers(options.IsGlobal, e, ctx)
	if err != nil {
		reply.Error(e, errs.ErrUnexpected)
		return
	}

	results := fetchPlayCounts(query, users, ctx)
	if len(results) == 0 {
		reply.Error(e, errs.ErrNoListeners)
		return
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].PlayCount > results[j].PlayCount
	})

	sendResponse(e, r, query, results, options)
}

func parseOptions(e *events.ApplicationCommandInteractionCreate) Options {
	data := e.SlashCommandInteractionData()

	limit := defaultLimit
	if l, ok := data.OptInt("limit"); ok {
		limit = l
	}

	name, _ := data.OptString("name")
	isGlobal, _ := data.OptBool("global")

	return Options{
		Type:     data.String("type"),
		Name:     name,
		Limit:    limit,
		IsGlobal: isGlobal,
	}
}

func buildQuery(options Options, e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) (*Query, error) {
	query := &Query{Type: options.Type}

	if options.Name != "" {
		query.Name = options.Name
	} else {
		if err := setQueryFromCurrentTrack(query, e, ctx); err != nil {
			return nil, err
		}
	}

	enrichQuery(query, ctx)
	return query, nil
}

func setQueryFromCurrentTrack(query *Query, e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) error {
	currentUser, err := ctx.Database.GetUser(ctx.Context, e.Member().User.ID.String())
	if err != nil {
		return errs.ErrUserNotFound
	}

	tracks, err := ctx.LastFM.User.GetRecentTracks(lfm.P{"user": currentUser, "limit": 1})
	if err != nil || len(tracks.Tracks) == 0 || tracks.Tracks[0].NowPlaying != "true" {
		return errs.ErrCurrentTrackFetch
	}

	track := tracks.Tracks[0]
	query.ArtistName = track.Artist.Name

	sanitize := func(s string) string {
		return strings.ReplaceAll(s, " ", "+")
	}

	switch query.Type {
	case "artist":
		query.Name = track.Artist.Name
		query.Url = fmt.Sprintf("https://www.last.fm/music/%s", sanitize(track.Artist.Name))
	case "track":
		query.Name = track.Name
		query.Url = track.Url
	case "album":
		query.Name = track.Album.Name
		query.Url = fmt.Sprintf("https://www.last.fm/music/%s/%s",
			sanitize(track.Artist.Name), sanitize(track.Album.Name))
	}

	return nil
}

func enrichQuery(query *Query, ctx ctx.CommandContext) {
	query.BetterName = query.Name
	query.Thumbnail = "https://lastfm.freetls.fastly.net/i/u/avatar170s/818148bf682d429dc215c1705eb27b98.png"

	switch query.Type {
	case "artist":
		if artist, err := ctx.LastFM.Artist.GetInfo(lfm.P{"artist": query.Name}); err == nil {
			if len(artist.Images) > 0 {
				artistImage, err := getArtistImage(ctx, artist.Name)
				if err == nil {
					query.Thumbnail = artistImage
				}
			}
			if artist.Name != "" {
				query.BetterName = artist.Name
			}
		}

	case "track":
		params := lfm.P{"track": query.Name}
		if query.ArtistName != "" {
			params["artist"] = query.ArtistName
		}
		if track, err := ctx.LastFM.Track.GetInfo(params); err == nil {
			if len(track.Album.Images) > 0 {
				query.Thumbnail = track.Album.Images[len(track.Album.Images)-1].Url
			}
			if track.Name != "" {
				query.BetterName = track.Name
			}
		}

	case "album":
		params := lfm.P{"album": query.Name}
		if query.ArtistName != "" {
			params["artist"] = query.ArtistName
		}
		if album, err := ctx.LastFM.Album.GetInfo(params); err == nil {
			if len(album.Images) > 0 {
				query.Thumbnail = album.Images[len(album.Images)-1].Url
			}
			if album.Name != "" {
				query.BetterName = album.Name
			}
		}
	}
}

func getUsers(isGlobal bool, e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) (map[snowflake.ID]string, error) {
	if isGlobal {
		return getAllUsers(ctx)
	}
	return ctx.LastFM.User.GetUsersByGuild(ctx.Context, e, ctx.Database)
}

func getAllUsers(ctx ctx.CommandContext) (map[snowflake.ID]string, error) {
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

	ctx.Cache.Members.Set(snowflake.ID(0), result, 0)
	return result, nil
}

func fetchPlayCounts(query *Query, users map[snowflake.ID]string, ctx ctx.CommandContext) []Result {
	if len(users) == 0 {
		return nil
	}

	workerCount := min(len(users), maxWorkers)
	sem := make(chan struct{}, workerCount)

	var (
		results []Result
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	ctx_timeout, cancel := context.WithTimeout(ctx.Context, timeout)
	defer cancel()

loop:
	for userID, username := range users {
		select {
		case <-ctx_timeout.Done():
			break loop
		default:
		}

		wg.Add(1)
		go func(userID snowflake.ID, username string) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx_timeout.Done():
				return
			}

			count := getUserPlayCount(query, username, ctx)
			if count > 0 {
				mu.Lock()
				results = append(results, Result{
					UserID:    userID.String(),
					Username:  username,
					PlayCount: count,
				})
				mu.Unlock()
			}
		}(userID, username)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx_timeout.Done():
	}

	return results
}

func getUserPlayCount(query *Query, username string, ctx ctx.CommandContext) int {
	params := lfm.P{
		"user": username,
		"name": query.Name,
		"type": query.Type,
	}

	if query.ArtistName != "" {
		params["artist"] = query.ArtistName
	}

	count, err := ctx.LastFM.User.GetPlays(params)
	if err != nil {
		return 0
	}
	return count
}

func sendResponse(e *events.ApplicationCommandInteractionCreate, r *reply.ResponseBuilder, query *Query, results []Result, options Options) {
	scope := "in this server"
	if options.IsGlobal {
		scope = "globally"
	} else if guild, ok := e.Guild(); ok {
		scope = fmt.Sprintf("in %s", guild.Name)
	}

	title := fmt.Sprintf("# %s\n-# Who knows %s %s %s?", query.BetterName, query.Type, query.BetterName, scope)
	list := buildResultsList(results, options.Limit)

	color := 0x00ADD8
	if dominantColor, err := colors.Dominant(query.Thumbnail); err == nil {
		color = dominantColor
	}

	component := discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplay(title),
			discord.NewTextDisplay(list),
		).WithAccessory(
			discord.NewThumbnail(query.Thumbnail),
		),
		discord.NewSmallSeparator(),
		discord.NewActionRow(
			discord.NewButton(
				discord.ButtonStyleLink,
				"Last.fm",
				"",
				url.PathEscape(query.Url),
				snowflake.ID(0),
			).WithEmoji(
				discord.NewCustomComponentEmoji(snowflake.MustParse("1418268922448187492")),
			),
		),
	).WithAccentColor(color)

	r.Flags(discord.MessageFlagIsComponentsV2).Component(component).Edit()
}

func buildResultsList(results []Result, limit int) string {
	if len(results) == 0 {
		return "no listeners found."
	}

	displayCount := min(len(results), limit)
	list := ""

	for i := range displayCount {
		r := results[i]
		count := i + 1

		var prefix string = fmt.Sprintf("%d.", count)
		switch count {
		case 1:
			prefix = emojis.EmojiRankOne
		case 2:
			prefix = emojis.EmojiRankTwo
		case 3:
			prefix = emojis.EmojiRankThree
		}

		list += fmt.Sprintf(
			"%s [%s](<https://www.last.fm/user/%s>) (*<@%s>*) — **%d** plays\n",
			prefix, r.Username, r.Username, r.UserID, r.PlayCount,
		)
	}

	if len(results) == 1 {
		list += "*this is pretty empty...*"
	}

	if len(results) > limit {
		list += fmt.Sprintf("\n*...and %d more listeners*", len(results)-limit)
	}

	return list
}

func getArtistImage(ctx ctx.CommandContext, name string) (string, error) {
	name = url.QueryEscape(name)

	if v, ok := ctx.Cache.Cover.Get(name); ok {
		return v, nil
	}
	endpoint := "https://api.deezer.com/search/artist?q=" + name

	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			PictureXL string `json:"picture_xl"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("no artist found")
	}

	image := result.Data[0].PictureXL

	ctx.Cache.Cover.Set(name, image, 0)

	return image, nil
}
