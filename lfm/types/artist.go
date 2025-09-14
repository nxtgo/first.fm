package types

// artist.getInfo
type ArtistGetInfo struct {
	Name   string `xml:"name"`
	MBID   string `xml:"mbid"`
	URL    string `xml:"url"`
	Images []struct {
		Size string `xml:"size,attr"`
		Url  string `xml:",chardata"`
	} `xml:"image"`
	Streamable string `xml:"streamable"`
	OnTour     string `xml:"ontour"`
	Stats      struct {
		Listeners     int `xml:"listeners"`
		PlayCount     int `xml:"playcount"`
		UserPlayCount int `xml:"userplaycount"`
	} `xml:"stats"`
	Similar struct {
		Artists []struct {
			Name   string `xml:"name"`
			Url    string `xml:"url"`
			Images []struct {
				Size string `xml:"size,attr"`
				Url  string `xml:",chardata"`
			} `xml:"image"`
		} `xml:"artist"`
	} `xml:"similar"`
	Tags struct {
		Tags []struct {
			Name string `xml:"name"`
			URL  string `xml:"url"`
		} `xml:"tag"`
	} `xml:"tags"`
	Bio struct {
		Links struct {
			Link struct {
				Rel  string `xml:"rel,attr"`
				Href string `xml:"href,attr"`
			} `xml:"link"`
		} `xml:"links"`
		Published string `xml:"published"`
		Summary   string `xml:"summary"`
		Content   string `xml:"content"`
	} `xml:"bio"`
}
