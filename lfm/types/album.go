package types

// album.getInfo
type AlbumGetInfo struct {
	Name   string `xml:"name"`
	Artist string `xml:"artist"`
	Mbid   string `xml:"mbid"`
	Url    string `xml:"url"`
	Images []struct {
		Size string `xml:"size,attr"`
		Url  string `xml:",chardata"`
	} `xml:"image"`
	Listeners     int `xml:"listeners"`
	PlayCount     int `xml:"playcount"`
	UserPlayCount int `xml:"userplaycount"`
	Tracks        []struct {
		Rank       int    `xml:"rank,attr"`
		Name       string `xml:"name"`
		Url        string `xml:"url"`
		Duration   string `xml:"duration"`
		Streamable struct {
			Fulltrack string `xml:"fulltrack,attr"`
			Value     string `xml:",chardata"`
		} `xml:"streamable"`
		Artist struct {
			Name string `xml:"name"`
			Mbid string `xml:"mbid"`
			Url  string `xml:"url"`
		} `xml:"artist"`
	} `xml:"tracks>track"`
	Tags []struct {
		Name string `xml:"name"`
		Url  string `xml:"url"`
	} `xml:"tags>tag"`
	Wiki struct {
		Published string `xml:"published"`
		Summary   string `xml:"summary"`
		Content   string `xml:"content"`
	} `xml:"wiki"`
}
