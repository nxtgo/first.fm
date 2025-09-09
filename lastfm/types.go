package lastfm

type WhoKnowsResult struct {
	UserID    string
	Username  string
	PlayCount int
}

type UserInfoResponse struct {
	User struct {
		Name        string `json:"name"`
		Age         string `json:"age"`
		Subscriber  string `json:"subscriber"`
		Realname    string `json:"realname"`
		Bootstrap   string `json:"bootstrap"`
		Playcount   string `json:"playcount"`
		ArtistCount string `json:"artist_count"`
		Playlists   string `json:"playlists"`
		TrackCount  string `json:"track_count"`
		AlbumCount  string `json:"album_count"`
		Image       []struct {
			Size string `json:"size"`
			Text string `json:"#text"`
		} `json:"image"`
		Registered struct {
			Unixtime string `json:"unixtime"`
			Text     int    `json:"#text"`
		} `json:"registered"`
		Country string `json:"country"`
		Gender  string `json:"gender"`
		URL     string `json:"url"`
		Type    string `json:"type"`
	} `json:"user"`
}

type TopTracksResponse struct {
	TopTracks struct {
		Track []struct {
			Streamable struct {
				Fulltrack string `json:"fulltrack"`
				Text      string `json:"#text"`
			} `json:"streamable"`
			Mbid  string `json:"mbid"`
			Name  string `json:"name"`
			Image []struct {
				Size string `json:"size"`
				Text string `json:"#text"`
			} `json:"image"`
			Artist struct {
				URL  string `json:"url"`
				Name string `json:"name"`
				Mbid string `json:"mbid"`
			} `json:"artist"`
			URL      string `json:"url"`
			Duration string `json:"duration"`
			Attr     struct {
				Rank string `json:"rank"`
			} `json:"@attr"`
			Playcount string `json:"playcount"`
		} `json:"track"`
		Attr struct {
			User       string `json:"user"`
			TotalPages string `json:"totalPages"`
			Page       string `json:"page"`
			PerPage    string `json:"perPage"`
			Total      string `json:"total"`
		} `json:"@attr"`
	} `json:"toptracks"`
}

type TopArtistsResponse struct {
	TopArtists struct {
		Artist []struct {
			Streamable string `json:"streamable"`
			Image      []struct {
				Size string `json:"size"`
				Text string `json:"#text"`
			} `json:"image"`
			Mbid      string `json:"mbid"`
			URL       string `json:"url"`
			Playcount string `json:"playcount"`
			Attr      struct {
				Rank string `json:"rank"`
			} `json:"@attr"`
			Name string `json:"name"`
		} `json:"artist"`
		Attr struct {
			User       string `json:"user"`
			TotalPages string `json:"totalPages"`
			Page       string `json:"page"`
			PerPage    string `json:"perPage"`
			Total      string `json:"total"`
		} `json:"@attr"`
	} `json:"topartists"`
}

type TopAlbumsResponse struct {
	TopAlbums struct {
		Album []struct {
			Artist struct {
				URL  string `json:"url"`
				Name string `json:"name"`
				Mbid string `json:"mbid"`
			} `json:"artist"`
			Image []struct {
				Size string `json:"size"`
				Text string `json:"#text"`
			} `json:"image"`
			Mbid      string `json:"mbid"`
			URL       string `json:"url"`
			Playcount string `json:"playcount"`
			Attr      struct {
				Rank string `json:"rank"`
			} `json:"@attr"`
			Name string `json:"name"`
		} `json:"album"`
		Attr struct {
			User       string `json:"user"`
			TotalPages string `json:"totalPages"`
			Page       string `json:"page"`
			PerPage    string `json:"perPage"`
			Total      string `json:"total"`
		} `json:"@attr"`
	} `json:"topalbums"`
}

type RecentTracksResponse struct {
	RecentTracks struct {
		Track []struct {
			Artist struct {
				Mbid string `json:"mbid"`
				Text string `json:"#text"`
			} `json:"artist"`
			Streamable string `json:"streamable"`
			Image      []struct {
				Size string `json:"size"`
				Text string `json:"#text"`
			} `json:"image"`
			Mbid  string `json:"mbid"`
			Album struct {
				Mbid string `json:"mbid"`
				Text string `json:"#text"`
			} `json:"album"`
			Name string `json:"name"`
			Attr struct {
				Nowplaying string `json:"nowplaying"`
			} `json:"@attr"`
			URL  string `json:"url"`
			Date struct {
				Uts  string `json:"uts"`
				Text string `json:"#text"`
			} `json:"date"`
		} `json:"track"`
		Attr struct {
			User       string `json:"user"`
			TotalPages string `json:"totalPages"`
			Page       string `json:"page"`
			PerPage    string `json:"perPage"`
			Total      string `json:"total"`
		} `json:"@attr"`
	} `json:"recenttracks"`
}
