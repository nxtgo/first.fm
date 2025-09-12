////////////////////////
///   LAST.FM V2     ///
/// DO NOT TOUCH!!!  ///
///////////////////////

// issuer: @elisiei -- please do not touch this file
// heavy wip, concept concerns.

package lfm

import (
	"encoding/xml"
	"fmt"

	httpx "github.com/nxtgo/httpx/client"
	"go.fm/cache"
)

const (
	lastFMBaseURL = "https://ws.audioscrobbler.com/2.0/"
)

type P map[string]any

type lastFMParams struct {
	apikey    string
	useragent string
}

type LastFMApi struct {
	params *lastFMParams
	client *httpx.Client
	apiKey string
	cache  *cache.LastFMCache

	User *userApi
}

func New(key string, c *cache.LastFMCache) *LastFMApi {
	params := lastFMParams{
		apikey: key,
		useragent: "go.fm/0.0.1 (discord bot; " +
			"https://github.com/nxtgo/go.fm; " +
			"contact: yehorovye@disroot.org)",
	}

	client := httpx.New().
		BaseURL(lastFMBaseURL).
		Header("User-Agent", params.useragent).
		Header("Accept", "application/xml")

	api := &LastFMApi{
		params: &params,
		client: client,
		apiKey: params.apikey,
		cache:  c,
	}

	api.User = &userApi{api: api}

	return api
}

func (c *LastFMApi) baseRequest(method string, params P) *httpx.Request {
	req := c.client.Get("").
		Query("api_key", c.apiKey).
		Query("format", "xml").
		Query("method", method)

	for k, v := range params {
		if str, ok := v.(string); ok {
			req.Query(k, str)
		}
	}

	return req
}

func decodeResponse(data []byte, out any) error {
	var env Envelope
	if err := xml.Unmarshal(data, &env); err != nil {
		return err
	}

	if env.Status == "failed" && env.Error != nil {
		return fmt.Errorf("lastfm error %d: %s", env.Error.Code, env.Error.Message)
	}

	if err := xml.Unmarshal(data, out); err != nil {
		return err
	}

	return nil
}
