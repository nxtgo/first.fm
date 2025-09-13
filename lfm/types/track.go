package types

// track.getInfo
type TrackGetInfo struct {
	Name       string `xml:"name"`
	Mbid       string `xml:"mbid"`
	Url        string `xml:"url"`
	Duration   int    `xml:"duration"`
	Streamable struct {
		Fulltrack string `xml:"fulltrack,attr"`
		Value     string `xml:",chardata"`
	} `xml:"streamable"`
	Listeners     int `xml:"listeners"`
	PlayCount     int `xml:"playcount"`
	UserPlayCount int `xml:"userplaycount"`
	UserLoved     int `xml:"userloved"`
	Artist        struct {
		Name string `xml:"name"`
		Mbid string `xml:"mbid"`
		Url  string `xml:"url"`
	} `xml:"artist"`
	Album struct {
		Artist string `xml:"artist"`
		Title  string `xml:"title"`
		URL    string `xml:"url"`
		Images []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"album"`
	TopTags struct {
		Tags []struct {
			Name string `xml:"name"`
			Url  string `xml:"url"`
		} `xml:"tag"`
	} `xml:"toptags"`
}
