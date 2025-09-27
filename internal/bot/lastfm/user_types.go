package lastfm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// User represents a Last.fm user
type User struct {
	Name       string    `xml:"name"`
	RealName   string    `xml:"realname"`
	URL        string    `xml:"url"`
	Country    string    `xml:"country"`
	Age        string    `xml:"age"`
	Gender     string    `xml:"gender"`
	Subscriber string    `xml:"subscriber"`
	PlayCount  int64     `xml:"playcount"`
	Registered Timestamp `xml:"registered"`
	Images     []Image   `xml:"image"`
}

// Image represents a Last.fm image
type Image struct {
	Size string `xml:"size,attr"`
	URL  string `xml:",chardata"`
}

// Timestamp represents a Last.fm timestamp
type Timestamp struct {
	UnixTime string `xml:"unixtime,attr"`
	Text     string `xml:",chardata"`
}

// TopAlbums wraps the list of albums and pagination info
type TopAlbums struct {
	User       string     `xml:"user,attr"`
	Page       int        `xml:"page,attr"`
	PerPage    int        `xml:"perPage,attr"`
	TotalPages int        `xml:"totalPages,attr"`
	Total      int        `xml:"total,attr"`
	Albums     []TopAlbum `xml:"album"`
}

// TopAlbum represents a single album
type TopAlbum struct {
	Rank      int            `xml:"rank,attr"`
	Name      string         `xml:"name"`
	Playcount int            `xml:"playcount"`
	MBID      string         `xml:"mbid"`
	URL       string         `xml:"url"`
	Artist    MinifiedArtist `xml:"artist"`
	Images    []Image        `xml:"image"`
}

// TopArtists wraps the list of artists and pagination info
type TopArtists struct {
	User       string      `xml:"user,attr"`
	Page       int         `xml:"page,attr"`
	PerPage    int         `xml:"perPage,attr"`
	TotalPages int         `xml:"totalPages,attr"`
	Total      int         `xml:"total,attr"`
	Artists    []TopArtist `xml:"artist"`
}

// TopArtist represents a single artist
type TopArtist struct {
	Rank       int     `xml:"rank,attr"`
	Name       string  `xml:"name"`
	Playcount  int     `xml:"playcount"`
	MBID       string  `xml:"mbid"`
	URL        string  `xml:"url"`
	Streamable bool    `xml:"streamable"`
	Images     []Image `xml:"image"`
}

// TopTracks wraps the list of tracks and pagination info
type TopTracks struct {
	User       string     `xml:"user,attr"`
	Page       int        `xml:"page,attr"`
	PerPage    int        `xml:"perPage,attr"`
	TotalPages int        `xml:"totalPages,attr"`
	Total      int        `xml:"total,attr"`
	Tracks     []TopTrack `xml:"track"`
}

// TopTrack represents a single track
type TopTrack struct {
	Rank       int            `xml:"rank,attr"`
	Name       string         `xml:"name"`
	Duration   int            `xml:"duration"`
	Playcount  int            `xml:"playcount"`
	MBID       string         `xml:"mbid"`
	URL        string         `xml:"url"`
	Streamable bool           `xml:"streamable"` // Can be "0" or "1" in XML
	Artist     MinifiedArtist `xml:"artist"`
	Images     []Image        `xml:"image"`
}

// RecentTrack represents a Last.fm track
type RecentTrack struct {
	Artist     RecentTrackArtist `xml:"artist"`
	Name       string            `xml:"name"`
	Streamable string            `xml:"streamable"`
	MBID       string            `xml:"mbid"`
	Album      RecentTrackAlbum  `xml:"album"`
	URL        string            `xml:"url"`
	Images     []Image           `xml:"image"`
	Date       *RecentTrackDate  `xml:"date"`
	NowPlaying string            `xml:"nowplaying,attr"`
}

// RecentTracks represents the recent tracks response
type RecentTracks struct {
	User       string        `xml:"user,attr"`
	Page       int           `xml:"page,attr"`
	PerPage    int           `xml:"perPage,attr"`
	TotalPages int           `xml:"totalPages,attr"`
	Total      int           `xml:"total,attr"`
	Tracks     []RecentTrack `xml:"track"`
}

// RecentTrackArtist represents an artist in a track
type RecentTrackArtist struct {
	MBID string `xml:"mbid,attr"`
	Name string `xml:",chardata"`
}

// RecentTrackAlbum represents an album in a track
type RecentTrackAlbum struct {
	MBID string `xml:"mbid,attr"`
	Name string `xml:",chardata"`
}

// RecentTrackDate represents a track's play date
type RecentTrackDate struct {
	UTS  string `xml:"uts,attr"`
	Text string `xml:",chardata"`
}

// MinifiedArtist represents the artist with less data idk
type MinifiedArtist struct {
	Name string `xml:"name"`
	MBID string `xml:"mbid"`
	URL  string `xml:"url"`
}

// IsNowPlaying returns true if the track is currently playing
func (t *RecentTrack) IsNowPlaying() bool {
	return t.NowPlaying == "true"
}

// GetPlayTime returns the track's play time as a Go time.Time
func (t *RecentTrack) GetPlayTime() (time.Time, error) {
	if t.Date == nil || t.Date.UTS == "" {
		return time.Time{}, nil
	}

	unixTime, err := strconv.ParseInt(t.Date.UTS, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(unixTime, 0), nil
}

// GetLargestImage returns the largest available image for the track
func (t *RecentTrack) GetLargestImage() *Image {
	sizes := []string{"mega", "extralarge", "large", "medium", "small"}
	for _, size := range sizes {
		for _, img := range t.Images {
			if img.Size == size && img.URL != "" {
				return &img
			}
		}
	}
	return nil
}

// GetImageBySize returns an image of the specified size for the track
func (t *RecentTrack) GetImageBySize(size string) *Image {
	for _, img := range t.Images {
		if img.Size == size && img.URL != "" {
			return &img
		}
	}
	return nil
}

func (t Timestamp) Time() (time.Time, error) {
	if t.UnixTime == "" {
		return time.Time{}, nil
	}

	unixTime, err := strconv.ParseInt(t.UnixTime, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(unixTime, 0), nil
}

func (u *User) GetImageBySize(size string) *Image {
	for _, img := range u.Images {
		if img.Size == size {
			return &img
		}
	}
	return nil
}

func (u *User) GetLargestImage() Image {
	sizes := []string{"mega", "extralarge", "large", "medium", "small"}
	for _, size := range sizes {
		if img := u.GetImageBySize(size); img != nil && img.URL != "" {
			return *img
		}
	}
	return Image{Size: "none", URL: ""}
}

func (u *User) GetPlayCount() int64 {
	return u.PlayCount
}

type deezerSearchResponse struct {
	Data []deezerArtist `json:"data"`
}

type deezerArtist struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Picture string `json:"picture_big"`
}

// GetDeezerImage fetches the artist image from Deezer matching the exact name
func (a *TopArtist) GetDeezerImage() (string, error) {
	baseURL := "https://api.deezer.com/search/artist"
	query := url.QueryEscape(a.Name)
	resp, err := http.Get(fmt.Sprintf("%s?q=%s", baseURL, query))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result deezerSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	for _, artist := range result.Data {
		if strings.EqualFold(artist.Name, a.Name) {
			return artist.Picture, nil
		}
	}

	return "", fmt.Errorf("artist %q not found on Deezer", a.Name)
}
