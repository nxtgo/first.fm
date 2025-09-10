package whoknows

import (
	"context"
	"fmt"
	"sort"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	"go.fm/lastfm"
	"go.fm/util/res"
	"go.fm/util/shared/cmd"
)

type Command struct{}

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
					{
						Name:  "artist",
						Value: "artist",
					},
					{
						Name:  "track",
						Value: "track",
					},
					{
						Name:  "album",
						Value: "album",
					},
				},
				Required: true,
			},
			discord.ApplicationCommandOptionString{
				Name:        "name",
				Description: "name of the artist/track/album",
				Required:    false,
			},
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := res.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = res.ErrorReply(e, constants.ErrorAcknowledgeCommand)
		return
	}

	tType := e.SlashCommandInteractionData().String("type")
	name, defined := e.SlashCommandInteractionData().OptString("name")
	if !defined {
		currentUser, err := ctx.Database.GetUser(
			context.Background(),
			e.Member().User.ID.String(),
		)
		if err != nil {
			_ = res.ErrorReply(e, constants.ErrorGetUser)
			return
		}

		tracks, err := ctx.LastFM.GetRecentTracks(currentUser, 1)
		current := tracks.RecentTracks.Track[0]
		if err != nil || current.Attr.Nowplaying == "false" {
			_ = res.ErrorReply(e, constants.ErrorFetchCurrentTrack)
			return
		}

		if current.Attr.Nowplaying != "true" {
			_ = res.ErrorReply(e, constants.ErrorNotPlaying)
			return
		}

		switch tType {
		case "artist":
			name = current.Artist.Text
		case "track":
			name = current.Name
		case "album":
			name = current.Album.Text
		}
	}

	users, err := lastfm.GetUsersByGuild(context.Background(), e, ctx.Database)
	if err != nil {
		_ = res.ErrorReply(e, constants.ErrorUnexpected)
		return
	}

	type result struct {
		UserID    string
		Username  string
		PlayCount int
	}

	results := make([]result, 0)

	for id, username := range users {
		count, err := ctx.LastFM.GetUserPlays(username, tType, name, 1000)
		if err != nil {
			continue
		}
		if count == 0 {
			continue
		}

		results = append(results, result{
			UserID:    id,
			Username:  username,
			PlayCount: count,
		})
	}
	if len(results) == 0 {
		_ = res.ErrorReply(e, constants.ErrorNoListeners)
		return
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].PlayCount > results[j].PlayCount
	})

	msg := fmt.Sprintf("### who knows %s `%s` best:\n", tType, name)

	for i, r := range results {
		if i >= 10 {
			break
		}
		msg += fmt.Sprintf("%d. %s (<@%s>) — %d plays\n", i+1, r.Username, r.UserID, r.PlayCount)
	}

	_ = reply.Content(msg).Send()
}
