////////////////////////
///   LAST.FM V2     ///
/// DO NOT TOUCH!!!  ///
///////////////////////

// issuer: @elisiei -- please do not touch this file
// heavy wip, concept concerns.

package lfm

import (
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

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
	cache  *cache.Cache

	User   *userApi
	Album  *albumApi
	Artist *artistApi
	Track  *trackApi
}

func New(key string, c *cache.Cache) *LastFMApi {
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
	api.Album = &albumApi{api: api}
	api.Artist = &artistApi{api: api}
	api.Track = &trackApi{api: api}

	return api
}

func (c *LastFMApi) baseRequest(method string, params P) *httpx.Request {
	req := c.client.Get("").
		Query("api_key", c.apiKey).
		Query("method", method)

	for k, v := range params {
		req.Query(k, fmt.Sprintf("%v", v))
	}

	return req
}

func decodeResponse(body []byte, result any) (err error) {
	var base Envelope
	err = xml.Unmarshal(body, &base)
	if err != nil {
		return
	}
	if base.Status == "failed" {
		var errorDetail ApiError
		err = xml.Unmarshal(base.Inner, &errorDetail)
		if err != nil {
			return
		}
		err = errors.New(errorDetail.Message)
		return
	} else if result == nil {
		return
	}
	err = xml.Unmarshal(base.Inner, result)
	return
}

func generateCacheKey(prefix string, args P) string {
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := []string{prefix}
	for _, k := range keys {
		if v, ok := args[k].(string); ok {
			parts = append(parts, fmt.Sprintf("%s:%v", k, v))
		}
	}

	keyString := strings.Join(parts, "|")

	if len(keyString) > 100 {
		hash := sha256.Sum256([]byte(keyString))
		return fmt.Sprintf("%s|%x", prefix, hash[:8])
	}

	return keyString
}
