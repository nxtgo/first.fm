package lastfm

import (
	"os"
	"time"

	httpx "github.com/nxtgo/httpx/client"
)

type Client struct {
	client *httpx.Client
	apiKey string
}

func New() *Client {
	client := httpx.New().
		BaseURL("https://ws.audioscrobbler.com/2.0/").
		Timeout(time.Second * 10)

	return &Client{
		client: client,
		apiKey: os.Getenv("LASTFM_API_KEY"),
	}
}

func (c *Client) req(method string, params map[string]string, target any) error {
	request := c.client.Get("")
	request.
		Query("api_key", c.apiKey).
		Query("format", "json").
		Query("method", method)

	for k, v := range params {
		request = request.Query(k, v)
	}

	return request.JSON(&target)
}
