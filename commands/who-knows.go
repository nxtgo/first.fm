package commands

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go.fm/bot/cache"
	"go.fm/util/res"
)

type WhoKnowsCommand struct{}

func (WhoKnowsCommand) Data() discord.ApplicationCommandCreate {
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

func (WhoKnowsCommand) Handle(e *events.ApplicationCommandInteractionCreate, ctx CommandContext) {
	reply := res.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = res.ErrorReply(e, "failed to acknowledge command")
		return
	}

	tType := e.SlashCommandInteractionData().String("type")
	name, defined := e.SlashCommandInteractionData().OptString("name")
	if !defined {
		currentUser, err := ctx.Database.GetUserByDiscordID(
			context.Background(),
			e.Member().User.ID.String(),
		)
		if err != nil {
			_ = res.ErrorReply(e, "could not get your last.fm username, use `/set-user`")
			return
		}

		tracks, err := ctx.LastFM.GetRecentTracks(currentUser.Username, 1)
		current := tracks.RecentTracks.Track[0]
		if err != nil || current.Attr.Nowplaying == "false" {
			_ = res.ErrorReply(e, "could not fetch your current track/artist/album")
			return
		}

		if current.Attr.Nowplaying != "true" {
			_ = res.ErrorReply(e, "you're not currently playing any track")
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

	users, err := ctx.Database.GetUsersByGuild(context.Background(), e.GuildID().String())
	if err != nil {
		_ = res.ErrorReply(e, err.Error())
		return
	}

	type result struct {
		UserID    string
		Username  string
		PlayCount int
	}

	results := make([]result, 0)

	for _, user := range users {
		count, err := whoKnowCache.GetOrFetch(user.Username, func(username string) (int, error) {
			return ctx.LastFM.GetUserPlays(username, tType, name, 1000)
		})
		if err != nil {
			continue
		}
		if count == 0 {
			continue
		}

		results = append(results, result{
			UserID:    user.DiscordID,
			Username:  user.Username,
			PlayCount: count,
		})
	}
	if len(results) == 0 {
		_ = res.ErrorReply(e, "no one has listened to this yet")
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

var whoKnowCache *cache.FuncCache[string, int]

func init() {
	whoKnowCache = cache.NewFuncCache[string, int](time.Second * 60)
	Register(WhoKnowsCommand{})
}
