package chart

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/lfm"
	"go.fm/pkg/bild/font"
	"go.fm/pkg/bild/imgio"
	"go.fm/pkg/bild/transform"
	"go.fm/pkg/constants/errs"
	"go.fm/pkg/constants/opts"
	"go.fm/pkg/ctx"
	"go.fm/pkg/discord/reply"
)

type Command struct{}

type Entry struct {
	Image  image.Image
	Name   string
	Artist string
}

var (
	maxGridSize   = 10
	minGridSize   = 3
	defaultPeriod = "overall"
)

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "chart",
		Description: "your top artists/tracks/albums but with images",
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
				Name:        "grid-size",
				Description: fmt.Sprintf("grid size (NxN) (min: %d, max: %d, default: min)", minGridSize, maxGridSize),
				Required:    false,
			},
			discord.ApplicationCommandOptionString{
				Name: "period",
				Description: fmt.Sprintf(
					"overall, 7day, 1month, 3month, 6month or 12month (default: %s)",
					defaultPeriod,
				),
				Choices: []discord.ApplicationCommandOptionChoiceString{
					{Name: "overall", Value: "overall"},
					{Name: "7day", Value: "7day"},
					{Name: "1month", Value: "1month"},
					{Name: "3month", Value: "3month"},
					{Name: "6month", Value: "6month"},
					{Name: "12month", Value: "12month"},
				},
				Required: false,
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

	t := e.SlashCommandInteractionData().String("type")
	var entries []Entry

	gridSize, gridSizeDefined := e.SlashCommandInteractionData().OptInt("grid-size")
	if !gridSizeDefined {
		gridSize = minGridSize
	}

	period, periodDefined := e.SlashCommandInteractionData().OptString("period")
	if !periodDefined {
		period = defaultPeriod
	}

	switch t {
	case "artist":
		reply.Error(e, errors.New("artist images are currently unsupported"))
		return
	case "track":
		// todo: workaround for track images. i hate last.fm. ~elisiei
		topTracks, err := ctx.LastFM.User.GetTopTracks(lfm.P{"user": user, "limit": gridSize * gridSize, "period": period})
		if err != nil {
			reply.Error(e, err)
			return
		}
		for _, track := range topTracks.Tracks {
			if len(track.Images) == 0 {
				continue
			}
			imgURL := track.Images[len(track.Images)-1].Url
			imgBytes, err := imgio.Fetch(imgURL)
			if err != nil {
				continue
			}
			img, err := imgio.Decode(imgBytes)
			if err != nil {
				continue
			}
			entries = append(entries, Entry{Image: img, Name: track.Name, Artist: track.Artist.Name})
		}
	case "album":
		topAlbums, err := ctx.LastFM.User.GetTopAlbums(lfm.P{"user": user, "limit": gridSize * gridSize, "period": period})
		if err != nil {
			reply.Error(e, err)
			return
		}
		for _, album := range topAlbums.Albums {
			if len(album.Images) == 0 {
				continue
			}
			imgURL := album.Images[len(album.Images)-1].Url
			imgBytes, err := imgio.Fetch(imgURL)
			if err != nil {
				brokenImage, _ := imgio.Open("assets/img/broken.png")
				resized := transform.Resize(brokenImage, 300, 300, transform.Gaussian)
				entries = append(entries, Entry{Image: resized, Name: album.Name, Artist: album.Artist.Name})
				continue
			}
			img, err := imgio.Decode(imgBytes)
			if err != nil {
				continue
			}
			entries = append(entries, Entry{Image: img, Name: album.Name, Artist: album.Artist.Name})
		}
	}

	if len(entries) == 0 {
		reply.Error(e, errs.ErrNoTracksFound)
		return
	}

	inter := font.LoadFont("assets/font/Inter_24pt-Regular.ttf")
	labelFace := inter.Face(20, 72)
	subFace := inter.Face(16, 72)

	firstBounds := entries[0].Image.Bounds()
	cellWidth := firstBounds.Dx()
	cellHeight := firstBounds.Dy()
	canvasWidth := cellWidth * gridSize
	canvasHeight := cellHeight * gridSize
	canvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))

	chartGradient, err := imgio.Open("assets/img/chart_gradient.png")
	if err != nil {
		reply.Error(e, errors.New("failed to load chart gradient"))
		return
	}

	for i, entry := range entries {
		row := i / gridSize
		col := i % gridSize
		x := col * cellWidth
		y := row * cellHeight
		rect := image.Rect(x, y, x+cellWidth, y+cellHeight)

		draw.Draw(canvas, rect, entry.Image, image.Point{}, draw.Over)
		draw.Draw(canvas, rect, chartGradient, image.Point{}, draw.Over)
		font.DrawText(canvas, x+15, y+labelFace.Metrics().Ascent.Ceil()+15, entry.Name, color.White, labelFace)

		if entry.Artist != "" {
			font.DrawText(
				canvas,
				x+15,
				y+labelFace.Metrics().Ascent.Ceil()+subFace.Metrics().Ascent.Ceil()+25,
				entry.Artist,
				color.RGBA{170, 170, 170, 255},
				subFace,
			)
		}
	}

	result, err := imgio.Encode(canvas, imgio.PNGEncoder())
	if err != nil {
		reply.Error(e, err)
		return
	}

	// see you twin <3
	r.File(discord.NewFile("chart.png", "", bytes.NewReader(result))).Edit()
}
