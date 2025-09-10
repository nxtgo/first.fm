package lastfm

import (
	"os"
	"time"

	httpx "github.com/nxtgo/httpx/client"
	"go.fm/cache"
)

type Client struct {
	client *httpx.Client
	apiKey string
	cache  *cache.LastFMCache
}

func New(cache *cache.LastFMCache) *Client {
	client := httpx.New().
		BaseURL("https://ws.audioscrobbler.com/2.0/").
		Timeout(time.Second * 10)

	return &Client{
		client: client,
		apiKey: os.Getenv("LASTFM_API_KEY"),
		cache:  cache,
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

	return request.JSON(target)
}

func (c *Client) CacheStats() string {
	return c.cache.Stats()
}
