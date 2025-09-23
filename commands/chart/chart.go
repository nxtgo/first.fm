package chart

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"net/http"
	"sync"
	"time"

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
	brokenImage   image.Image
	maxConcurrent = 8
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
	Timeout: 10 * time.Second,
}

type Entry struct {
	Image  image.Image
	Name   string
	Artist string
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

		entries := make([]Entry, 0, gridSize*gridSize)

		switch options.Type {
		case "artist":
			topArtists, err := c.Last.User.GetTopArtists(lastfm.P{"user": username, "limit": gridSize * gridSize, "period": period})
			if err != nil {
				return err
			}

			urls := make([]string, len(topArtists.Artists))
			names := make([]string, len(topArtists.Artists))
			for i, a := range topArtists.Artists {
				imgURL, err := a.GetDeezerImage()
				if err != nil || imgURL == "" {
					urls[i] = ""
				} else {
					urls[i] = imgURL
				}
				names[i] = a.Name
			}
			fetched := fetchEntries(urls)
			for i, e := range fetched {
				if e.Image == nil {
					e.Image = brokenImage
				}
				e.Image = transform.Resize(e.Image, 300, 300, transform.NearestNeighbor)
				e.Name = names[i]
				entries = append(entries, e)
			}

		case "track":
			topTracks, err := c.Last.User.GetTopTracks(lastfm.P{"user": username, "limit": gridSize * gridSize, "period": period})
			if err != nil {
				return err
			}

			urls := make([]string, len(topTracks.Tracks))
			names := make([]string, len(topTracks.Tracks))
			artists := make([]string, len(topTracks.Tracks))
			for i, t := range topTracks.Tracks {
				if len(t.Images) > 0 {
					urls[i] = t.Images[len(t.Images)-1].URL
				}
				names[i] = t.Name
				artists[i] = t.Artist.Name
			}
			fetched := fetchEntries(urls)
			for i, e := range fetched {
				if e.Image == nil {
					e.Image = brokenImage
				}
				e.Name = names[i]
				e.Artist = artists[i]
				entries = append(entries, e)
			}

		case "album":
			topAlbums, err := c.Last.User.GetTopAlbums(lastfm.P{"user": username, "limit": gridSize * gridSize, "period": period})
			if err != nil {
				return err
			}

			urls := make([]string, len(topAlbums.Albums))
			names := make([]string, len(topAlbums.Albums))
			artists := make([]string, len(topAlbums.Albums))
			for i, a := range topAlbums.Albums {
				if len(a.Images) > 0 {
					urls[i] = a.Images[len(a.Images)-1].URL
				}
				names[i] = a.Name
				artists[i] = a.Artist.Name
			}
			fetched := fetchEntries(urls)
			for i, e := range fetched {
				if e.Image == nil {
					e.Image = brokenImage
				}
				e.Name = names[i]
				e.Artist = artists[i]
				entries = append(entries, e)
			}
		}

		if len(entries) == 0 {
			return errors.New("no entries found")
		}

		interRegular := font.LoadFont("assets/font/Inter_24pt-Regular.ttf")
		interBold := font.LoadFont("assets/font/Inter_24pt-Bold.ttf")

		labelFace := interBold.Face(20, 72)
		subFace := interRegular.Face(16, 72)

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

		labelAscent := labelFace.Metrics().Ascent.Ceil()
		subAscent := subFace.Metrics().Ascent.Ceil()

		for i, entry := range entries {
			row := i / gridSize
			col := i % gridSize
			x := col * cellWidth
			y := row * cellHeight
			rect := image.Rect(x, y, x+cellWidth, y+cellHeight)

			draw.Draw(canvas, rect, entry.Image, image.Point{}, draw.Over)
			draw.Draw(canvas, rect, chartGradient, image.Point{}, draw.Over)

			font.DrawText(canvas, x+15, y+labelAscent+15, entry.Name, color.White, labelFace)

			if entry.Artist != "" {
				font.DrawText(canvas, x+15, y+labelAscent+subAscent+20,
					entry.Artist, color.RGBA{170, 170, 170, 255}, subFace)
			}
		}

		result, err := imgio.Encode(canvas, imgio.PNGEncoder())
		if err != nil {
			return err
		}

		_, err = edit.Contentf("%s %s chart for %s", period, options.Type, username).File(sendpart.File{Name: "chart.png", Reader: bytes.NewReader(result)}).Send()
		return err
	})
}

func init() {
	// todo: remove this in the future*
	brokenImage, _ = imgio.Open("assets/img/broken.png")
	brokenImage = transform.Resize(brokenImage, 300, 300, transform.NearestNeighbor)
	commands.Register(data, handler)
}

func fetchImage(url string) image.Image {
	if url == "" {
		return nil
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil
	}
	return img
}

func fetchEntries(urls []string) []Entry {
	entries := make([]Entry, len(urls))
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)

	for i, url := range urls {
		i, url := i, url
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			entries[i].Image = fetchImage(url)
		})
	}

	wg.Wait()
	return entries
}
