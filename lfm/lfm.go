package lfm

import (
	"context"
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

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

var defaultRateLimiter = time.Tick(100 * time.Millisecond)

func New(key string, c *cache.Cache) *LastFMApi {
	params := lastFMParams{
		apikey:    key,
		useragent: "go.fm/0.0.1 (discord bot; https://github.com/nxtgo/go.fm; contact: yehorovye@disroot.org)",
	}

	api := &LastFMApi{
		params: &params,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 50,
			},
		},
		apiKey: key,
		cache:  c,
	}

	api.User = &userApi{api: api}
	api.Album = &albumApi{api: api}
	api.Artist = &artistApi{api: api}
	api.Track = &trackApi{api: api}

	return api
}

func (c *LastFMApi) baseRequest(method string, params P) (*http.Response, error) {
	<-defaultRateLimiter

	values := url.Values{}
	values.Set("api_key", c.apiKey)
	values.Set("method", method)
	for k, v := range params {
		values.Set(k, fmt.Sprintf("%v", v))
	}

	u := lastFMBaseURL + "?" + values.Encode()

	req, err := http.NewRequestWithContext(context.Background(), "GET", u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.params.useragent)
	req.Header.Set("Accept", "application/xml")

	return c.client.Do(req)
}

func (c *LastFMApi) doAndDecode(method string, params P, result any) error {
	resp, err := c.baseRequest(method, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return decodeResponse(resp.Body, result)
}

func decodeResponse(r io.Reader, result any) (err error) {
	var base Envelope
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if err = xml.Unmarshal(body, &base); err != nil {
		return err
	}

	if base.Status == "failed" {
		var errorDetail ApiError
		if err = xml.Unmarshal(base.Inner, &errorDetail); err != nil {
			return err
		}
		return errors.New(errorDetail.Message)
	}

	if result != nil {
		return xml.Unmarshal(base.Inner, result)
	}

	return nil
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
