package lastfm

import (
	"strconv"
	"time"
)

// Track represents a Last.fm track
type Track struct {
	Artist     TrackArtist `xml:"artist"`
	Name       string      `xml:"name"`
	Streamable string      `xml:"streamable"`
	MBID       string      `xml:"mbid"`
	Album      TrackAlbum  `xml:"album"`
	URL        string      `xml:"url"`
	Images     []Image     `xml:"image"`
	Date       *TrackDate  `xml:"date"`
	NowPlaying string      `xml:"nowplaying,attr"`
}

// TrackArtist represents an artist in a track
type TrackArtist struct {
	MBID string `xml:"mbid,attr"`
	Name string `xml:",chardata"`
}

// TrackAlbum represents an album in a track
type TrackAlbum struct {
	MBID string `xml:"mbid,attr"`
	Name string `xml:",chardata"`
}

// TrackDate represents a track's play date
type TrackDate struct {
	UTS  string `xml:"uts,attr"`
	Text string `xml:",chardata"`
}

// IsNowPlaying returns true if the track is currently playing
func (t *Track) IsNowPlaying() bool {
	return t.NowPlaying == "true"
}

// GetPlayTime returns the track's play time as a Go time.Time
func (t *Track) GetPlayTime() (time.Time, error) {
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
func (t *Track) GetLargestImage() *Image {
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
func (t *Track) GetImageBySize(size string) *Image {
	for _, img := range t.Images {
		if img.Size == size && img.URL != "" {
			return &img
		}
	}
	return nil
}
