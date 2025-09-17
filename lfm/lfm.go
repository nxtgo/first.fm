package lfm

import (
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

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
	client *http.Client
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

	api := &LastFMApi{
		params: &params,
		client: &http.Client{},
		apiKey: params.apikey,
		cache:  c,
	}

	api.User = &userApi{api: api}
	api.Album = &albumApi{api: api}
	api.Artist = &artistApi{api: api}
	api.Track = &trackApi{api: api}

	return api
}

// baseRequest constructs and executes a GET request.
func (c *LastFMApi) baseRequest(method string, params P) (*http.Response, error) {
	// Build query
	values := url.Values{}
	values.Set("api_key", c.apiKey)
	values.Set("method", method)

	for k, v := range params {
		values.Set(k, fmt.Sprintf("%v", v))
	}

	u := lastFMBaseURL + "?" + values.Encode()

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.params.useragent)
	req.Header.Set("Accept", "application/xml")

	return c.client.Do(req)
}

// helper to read and decode
func (c *LastFMApi) doAndDecode(method string, params P, result any) error {
	resp, err := c.baseRequest(method, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return decodeResponse(body, result)
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
		if v, ok := args[k]; ok {
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
