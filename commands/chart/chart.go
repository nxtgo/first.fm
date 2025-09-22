package chart

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"net/http"

	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/discord"
	"github.com/nxtgo/arikawa/v3/utils/sendpart"
	"go.fm/bild/font"
	"go.fm/bild/imgio"
	"go.fm/bild/transform"
	"go.fm/commands"
	lastfm "go.fm/last.fm"
	"go.fm/pkg/reply"
)

var (
	maxGridSize   = 10
	minGridSize   = 3
	defaultPeriod = "overall"
)

type Entry struct {
	Image  image.Image
	Name   string
	Artist string
}

type deezerSearchResponse struct {
	Data []struct {
		Picture string `json:"picture"`
	} `json:"data"`
}

var data = api.CreateCommandData{
	Name:        "chart",
	Description: "Your top artists/tracks/albums but with images",
	Options: discord.CommandOptions{
		&discord.StringOption{
			OptionName:  "type",
			Description: "artist, track or album",
			Choices: []discord.StringChoice{
				{Name: "artist", Value: "artist"},
				{Name: "track", Value: "track"},
				{Name: "album", Value: "album"},
			},
			Required: true,
		},
		discord.NewIntegerOption("grid-size", fmt.Sprintf("grid size (NxN) (min: %d, max: %d, default: min)", minGridSize, maxGridSize), false),
		&discord.StringOption{
			OptionName:  "period",
			Description: fmt.Sprintf("overall, 7day, 1month, 3month, 6month or 12month (default: %s)", defaultPeriod),
			Choices: []discord.StringChoice{
				{Name: "overall", Value: "overall"},
				{Name: "7day", Value: "7day"},
				{Name: "1month", Value: "1month"},
				{Name: "3month", Value: "3month"},
				{Name: "6month", Value: "6month"},
				{Name: "12month", Value: "12month"},
			},
			Required: false,
		},
		discord.NewStringOption("user", "user to fetch chart for", false),
	},
}

var options struct {
	User     *string `discord:"user"`
	Type     string  `discord:"type"`
	GridSize *int    `discord:"grid-size"`
	Period   *string `discord:"period"`
}

func handler(c *commands.CommandContext) error {
	return c.Reply.AutoDefer(func(edit *reply.EditBuilder) error {
		if err := c.Data.Options.Unmarshal(&options); err != nil {
			return err
		}

		gridSize := minGridSize
		if options.GridSize != nil {
			gridSize = *options.GridSize
		}

		period := defaultPeriod
		if options.Period != nil {
			period = *options.Period
		}

		username, err := c.GetUserOrFallback()
		if err != nil {
			return err
		}

		brokenImage, _ := imgio.Open("assets/img/broken.png")
		brokenImage = transform.Resize(brokenImage, 300, 300, transform.Gaussian)

		fetchImage := func(url string) image.Image {
			resp, err := http.Get(url)
			if err != nil {
				return brokenImage
			}
			defer resp.Body.Close()

			img, _, err := image.Decode(resp.Body)
			if err == nil {
				return img
			}

			gifImg, err := gif.Decode(resp.Body)
			if err != nil {
				return brokenImage
			}
			return gifImg
		}

		var entries []Entry

		switch options.Type {
		case "artist":
			topArtists, err := c.Last.User.GetTopArtists(lastfm.P{"user": username, "limit": gridSize * gridSize, "period": period})
			if err != nil {
				return err
			}

			for _, a := range topArtists.Artists {
				imgURL, err := a.GetDeezerImage()
				if err != nil || imgURL == "" {
					entries = append(entries, Entry{Image: brokenImage, Name: a.Name})
					continue
				}
				img := fetchImage(imgURL)
				img = transform.Resize(img, 300, 300, transform.Gaussian)
				entries = append(entries, Entry{Image: img, Name: a.Name})
			}

		case "track":
			topTracks, err := c.Last.User.GetTopTracks(lastfm.P{"user": username, "limit": gridSize * gridSize, "period": period})
			if err != nil {
				return err
			}
			for _, t := range topTracks.Tracks {
				img := brokenImage
				if len(t.Images) > 0 {
					img = fetchImage(t.Images[len(t.Images)-1].URL)
				}
				entries = append(entries, Entry{Image: img, Name: t.Name, Artist: t.Artist.Name})
			}

		case "album":
			topAlbums, err := c.Last.User.GetTopAlbums(lastfm.P{"user": username, "limit": gridSize * gridSize, "period": period})
			if err != nil {
				return err
			}
			for _, a := range topAlbums.Albums {
				img := brokenImage
				if len(a.Images) > 0 {
					img = fetchImage(a.Images[len(a.Images)-1].URL)
				}
				entries = append(entries, Entry{Image: img, Name: a.Name, Artist: a.Artist.Name})
			}
		}

		if len(entries) == 0 {
			return errors.New("no entries found")
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
			return err
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
				font.DrawText(canvas, x+15, y+labelFace.Metrics().Ascent.Ceil()+subFace.Metrics().Ascent.Ceil()+25,
					entry.Artist, color.RGBA{170, 170, 170, 255}, subFace)
			}
		}

		result, err := imgio.Encode(canvas, imgio.PNGEncoder())
		if err != nil {
			return err
		}

		_, err = edit.File(sendpart.File{Name: "chart.png", Reader: bytes.NewReader(result)}).Send()
		return err
	})
}

func init() {
	commands.Register(data, handler)
}
