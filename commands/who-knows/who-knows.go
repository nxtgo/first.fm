package whoknows

import (
	"fmt"
	"sort"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	"go.fm/lfm"
	"go.fm/types/cmd"
	"go.fm/utils/image"
)

type Command struct{}

var (
	maxLimit int = 100
	minLimit int = 5
)

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "who-knows",
		Description: "see who in this guild has listened to a track/artist/album the most",
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
				Description: "max entries for the list (max: 100, min: 5)",
				Required:    false,
				MinValue:    &minLimit,
				MaxValue:    &maxLimit,
			},
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := ctx.Reply(e)
	if err := reply.Defer(); err != nil {
		_ = ctx.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	t := e.SlashCommandInteractionData().String("type")
	limit, limitDefined := e.SlashCommandInteractionData().OptInt("limit")
	name, defined := e.SlashCommandInteractionData().OptString("name")
	var artistName string

	if !limitDefined {
		limit = 10
	}

	if !defined {
		currentUser, err := ctx.Database.GetUser(ctx.Context, e.Member().User.ID.String())
		if err != nil {
			_ = ctx.Error(e, constants.ErrorGetUser)
			return
		}

		tracks, err := ctx.LastFM.User.GetRecentTracks(lfm.P{"user": currentUser, "limit": 1})
		if err != nil || len(tracks.Tracks) == 0 || tracks.Tracks[0].NowPlaying != "true" {
			_ = ctx.Error(e, constants.ErrorFetchCurrentTrack)
			return
		}

		current := tracks.Tracks[0]
		switch t {
		case "artist":
			name = current.Artist.Name
		case "track":
			name = current.Name
			artistName = current.Artist.Name
		case "album":
			name = current.Album.Name
			artistName = current.Artist.Name
		}
	}

	users, err := ctx.LastFM.User.GetUsersByGuild(ctx.Context, e, ctx.Database)
	if err != nil {
		_ = ctx.Error(e, constants.ErrorUnexpected)
		return
	}

	type result struct {
		UserID    string
		Username  string
		PlayCount int
	}

	var (
		results []result
		mu      sync.Mutex
		wg      sync.WaitGroup
		sem     = make(chan struct{}, limit)
	)

	for id, username := range users {
		idCopy, usernameCopy := id.String(), username
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			count, err := ctx.LastFM.User.GetPlays(lfm.P{
				"user":   usernameCopy,
				"name":   name,
				"type":   t,
				"artist": artistName,
			})
			if err != nil || count == 0 {
				return
			}

			mu.Lock()
			results = append(results, result{
				UserID:    idCopy,
				Username:  usernameCopy,
				PlayCount: count,
			})
			mu.Unlock()
		})
	}

	wg.Wait()

	if len(results) == 0 {
		ctx.Error(e, constants.ErrorNoListeners)
		return
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].PlayCount > results[j].PlayCount
	})

	betterName := name
	var thumbnail string

	switch t {
	case "artist":
		artist, err := ctx.LastFM.Artist.GetInfo(lfm.P{"artist": name})
		if err == nil {
			if len(artist.Images) > 0 {
				thumbnail = artist.Images[len(artist.Images)-1].Url
			}
			if artist.Name != "" {
				betterName = artist.Name
			}
		}

	case "track":
		params := lfm.P{"track": name}
		if artistName != "" {
			params["artist"] = artistName
		}
		track, err := ctx.LastFM.Track.GetInfo(params)
		if err == nil {
			if len(track.Album.Images) > 0 {
				thumbnail = track.Album.Images[len(track.Album.Images)-1].Url
			}
			if track.Name != "" {
				betterName = track.Name
			}
		}

	case "album":
		params := lfm.P{"album": name}
		if artistName != "" {
			params["artist"] = artistName
		}
		album, err := ctx.LastFM.Album.GetInfo(params)
		if err == nil {
			if len(album.Images) > 0 {
				thumbnail = album.Images[len(album.Images)-1].Url
			}
			if album.Name != "" {
				betterName = album.Name
			}
		}
	}

	if thumbnail == "" {
		thumbnail = "https://lastfm.freetls.fastly.net/i/u/avatar170s/818148bf682d429dc215c1705eb27b98.png"
	}

	list := ""
	for i, r := range results {
		if i >= limit {
			break
		}
		list += fmt.Sprintf(
			"%d. [%s](<https://www.last.fm/user/%s>) (*<@%s>*) — **%d** plays\n",
			i+1, r.Username, r.Username, r.UserID, r.PlayCount,
		)
	}

	color, err := image.DominantColor(thumbnail)
	if err != nil {
		color = 0x00ADD8
	}

	guild, _ := e.Guild()
	component := discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplayf("### Who knows %s %s?\n-# in *%s*", t, betterName, guild.Name),
			discord.NewTextDisplay(list),
		).WithAccessory(discord.NewThumbnail(thumbnail)),
	).WithAccentColor(color)

	reply.Flags(discord.MessageFlagIsComponentsV2).Component(component).Edit()
}
