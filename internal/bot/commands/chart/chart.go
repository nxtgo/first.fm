package chart

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"net/http"
	"sync"
	"time"

	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/discord"
	"github.com/nxtgo/arikawa/v3/utils/sendpart"
	"go.fm/internal/bot/commands"
	"go.fm/internal/bot/discord/reply"
	"go.fm/internal/bot/image/imgio"
	"go.fm/internal/bot/image/transform"
	"go.fm/internal/bot/lastfm"
)

var (
	maxGridSize   = 10
	minGridSize   = 3
	defaultPeriod = "overall"
	maxConcurrent = 8

	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 10 * time.Second,
	}
)

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
		discord.NewIntegerOption("grid-size",
			fmt.Sprintf("grid size (NxN) (min: %d, max: %d, default: min)", minGridSize, maxGridSize), false),
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

		grid := minGridSize
		if options.GridSize != nil {
			grid = *options.GridSize
		}

		period := defaultPeriod
		if options.Period != nil {
			period = *options.Period
		}

		user, err := c.GetUserOrFallback()
		if err != nil {
			return err
		}

		entries, err := fetchChartEntries(c, options.Type, user, grid, period)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			return errors.New("no entries found")
		}

		img, err := renderChart(entries, grid)
		if err != nil {
			return err
		}

		_, err = edit.
			Contentf("%s %s chart for %s", period, options.Type, user).
			File(sendpart.File{Name: "chart.png", Reader: bytes.NewReader(img)}).
			Send()
		return err
	})
}

func fetchChartEntries(c *commands.CommandContext, kind, user string, grid int, period string) ([]Entry, error) {
	limit := grid * grid

	var urls, names, artists []string
	switch kind {
	case "artist":
		res, err := c.Last.User.GetTopArtists(lastfm.P{"user": user, "limit": limit, "period": period})
		if err != nil {
			return nil, err
		}
		urls, names = make([]string, len(res.Artists)), make([]string, len(res.Artists))
		for i, a := range res.Artists {
			if u, _ := a.GetDeezerImage(); u != "" {
				urls[i] = u
			}
			names[i] = a.Name
		}
	case "track":
		res, err := c.Last.User.GetTopTracks(lastfm.P{"user": user, "limit": limit, "period": period})
		if err != nil {
			return nil, err
		}
		urls, names, artists = make([]string, len(res.Tracks)), make([]string, len(res.Tracks)), make([]string, len(res.Tracks))
		for i, t := range res.Tracks {
			if len(t.Images) > 0 {
				urls[i] = t.Images[len(t.Images)-1].URL
			}
			names[i], artists[i] = t.Name, t.Artist.Name
		}
	case "album":
		res, err := c.Last.User.GetTopAlbums(lastfm.P{"user": user, "limit": limit, "period": period})
		if err != nil {
			return nil, err
		}
		urls, names, artists = make([]string, len(res.Albums)), make([]string, len(res.Albums)), make([]string, len(res.Albums))
		for i, a := range res.Albums {
			if len(a.Images) > 0 {
				urls[i] = a.Images[len(a.Images)-1].URL
			}
			names[i], artists[i] = a.Name, a.Artist.Name
		}
	}

	cellSize := 300
	if grid >= 10 {
		cellSize = 100
	}

	broken, _ := imgio.Open("assets/img/broken.png")
	broken = transform.Resize(broken, cellSize, cellSize, transform.NearestNeighbor)

	fetched := fetchEntries(urls)
	entries := make([]Entry, len(fetched))
	for i, e := range fetched {
		if e.Image == nil {
			e.Image = broken
		}
		e.Image = transform.Resize(e.Image, cellSize, cellSize, transform.NearestNeighbor)
		e.Name = names[i]
		if artists != nil {
			e.Artist = artists[i]
		}
		entries[i] = e
	}
	return entries, nil
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

func init() {
	commands.Register(data, handler)
}
