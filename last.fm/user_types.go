package lastfm

import (
	"strconv"
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

// RecentTracks represents the recent tracks response
type RecentTracks struct {
	User       string        `xml:"user,attr"`
	Page       int           `xml:"page,attr"`
	PerPage    int           `xml:"perPage,attr"`
	TotalPages int           `xml:"totalPages,attr"`
	Total      int           `xml:"total,attr"`
	Tracks     []RecentTrack `xml:"track"`
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
