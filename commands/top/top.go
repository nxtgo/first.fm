package top

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/lfm"
	"go.fm/pkg/constants/errs"
	"go.fm/pkg/constants/opts"
	"go.fm/pkg/ctx"
	"go.fm/pkg/discord/reply"
)

type Command struct{}

var (
	maxLimit int = 100
	minLimit int = 5
)

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "top",
		Description: "get an user's top tracks/albums/artists",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
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
			discord.ApplicationCommandOptionInt{
				Name:        "limit",
				Description: "max entries for the list (max: 100, min: 5)",
				Required:    false,
				MinValue:    &minLimit,
				MaxValue:    &maxLimit,
			},
			opts.UserOption,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) {
	r := reply.New(e)
	if err := r.Defer(); err != nil {
		reply.Error(e, errs.ErrCommandDeferFailed)
		return
	}

	user, err := ctx.GetUser(e)
	if err != nil {
		reply.Error(e, errs.ErrUserNotFound)
		return
	}

	topType := e.SlashCommandInteractionData().String("type")
	limit, limitDefined := e.SlashCommandInteractionData().OptInt("limit")
	if !limitDefined {
		limit = 10
	}

	var description string

	switch topType {
	case "artist":
		data, err := ctx.LastFM.User.GetTopArtists(lfm.P{
			"user":  user,
			"limit": limit,
		})
		if err != nil {
			reply.Error(e, err)
			return
		}

		for i, a := range data.Artists {
			if i > limit {
				break
			}
			description += fmt.Sprintf("%d. %s — **%s** plays\n", i+1, a.Name, a.PlayCount)
		}

	case "track":
		data, err := ctx.LastFM.User.GetTopTracks(lfm.P{
			"user":  user,
			"limit": limit,
		})
		if err != nil {
			reply.Error(e, err)
			return
		}

		for i, t := range data.Tracks {
			if i > limit {
				break
			}
			description += fmt.Sprintf("%d. %s — *%s* (**%s** plays)\n", i+1, t.Name, t.Artist.Name, t.PlayCount)
		}

	case "album":
		data, err := ctx.LastFM.User.GetTopAlbums(lfm.P{
			"user":  user,
			"limit": limit,
		})
		if err != nil {
			reply.Error(e, err)
			return
		}

		for i, a := range data.Albums {
			if i > limit {
				break
			}
			description += fmt.Sprintf("%d. %s — *%s* (**%s** plays)\n", i+1, a.Name, a.Artist.Name, a.PlayCount)
		}
	}

	if description == "" {
		description = errs.ErrNoTracksFound.Error()
	}

	component := discord.NewContainer(
		discord.NewTextDisplayf("### %s's top %ss", user, topType),
		discord.NewTextDisplay(description),
		discord.NewSmallSeparator(),
		discord.NewTextDisplay("-# *if results are odd, use `/update`*"),
	)

	r.Flags(discord.MessageFlagIsComponentsV2).Component(component).Edit()
}
