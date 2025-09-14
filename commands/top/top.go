package top

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	"go.fm/lfm"
	"go.fm/types/cmd"
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
			cmd.UserOption,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := ctx.Reply(e)
	if err := reply.Defer(); err != nil {
		ctx.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	user, err := ctx.GetUser(e)
	if err != nil {
		ctx.Error(e, constants.ErrorGetUser)
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
			_ = ctx.Error(e, err.Error())
			return
		}
		for i, a := range data.Artists {
			description += fmt.Sprintf("%d. %s — **%s** plays\n", i+1, a.Name, a.PlayCount)
		}

	case "track":
		data, err := ctx.LastFM.User.GetTopTracks(lfm.P{
			"user":  user,
			"limit": limit,
		})
		if err != nil {
			_ = ctx.Error(e, err.Error())
			return
		}
		for i, t := range data.Tracks {
			description += fmt.Sprintf("%d. %s — *%s* (**%s** plays)\n", i+1, t.Name, t.Artist.Name, t.PlayCount)
		}

	case "album":
		data, err := ctx.LastFM.User.GetTopAlbums(lfm.P{
			"user":  user,
			"limit": limit,
		})
		if err != nil {
			_ = ctx.Error(e, err.Error())
			return
		}
		for i, a := range data.Albums {
			description += fmt.Sprintf("%d. %s — *%s* (**%s** plays)\n", i+1, a.Name, a.Artist.Name, a.PlayCount)
		}
	}

	if description == "" {
		description = constants.ErrorNoTracks
	}

	component := discord.NewContainer(
		discord.NewTextDisplayf("### %s's top %ss", user, topType),
		discord.NewTextDisplay(description),
	)

	reply.Flags(discord.MessageFlagIsComponentsV2).Component(component).Edit()
}
