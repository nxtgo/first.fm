package api

type Client struct {
	*API
	Album  *Album
	Artist *Artist
	Chart  *Chart
	Track  *Track
	User   *User
}

func NewClient(apiKey string) *Client {
	return newClient(New(apiKey))
}

func newClient(a *API) *Client {
	return &Client{
		API:    a,
		Album:  NewAlbum(a),
		Artist: NewArtist(a),
		Chart:  NewChart(a),
		Track:  NewTrack(a),
		User:   NewUser(a),
	}
}
