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

// RecentTracks represents the recent tracks response
type RecentTracks struct {
	User       string  `xml:"user,attr"`
	Page       int     `xml:"page,attr"`
	PerPage    int     `xml:"perPage,attr"`
	TotalPages int     `xml:"totalPages,attr"`
	Total      int     `xml:"total,attr"`
	Tracks     []Track `xml:"track"`
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
