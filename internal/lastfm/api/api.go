package api

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"first.fm/internal/lastfm"
	"golang.org/x/time/rate"
)

var (
	BaseEndpoint = "https://ws.audioscrobbler.com"
	Version      = "2.0"
	Endpoint     = BaseEndpoint + "/" + Version + "/"
)

const (
	DefaultUserAgent      = "first.fm/0.0.1 (discord bot; https://github.com/nxtgo/first.fm; contact: yehorovye@disroot.org)"
	DefaultRetries   uint = 5
	DefaultTimeout        = 30
)

type RequestLevel int

const (
	RequestLevelNone RequestLevel = iota
	RequestLevelAPIKey
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type API struct {
	APIKey      string
	UserAgent   string
	Retries     uint
	Client      HTTPClient
	rateLimiter *rate.Limiter
}

func New(apiKey string) *API {
	return NewWithTimeout(apiKey, DefaultTimeout)
}

func NewWithTimeout(apiKey string, timeout int) *API {
	t := time.Duration(timeout) * time.Second
	return &API{
		APIKey:      apiKey,
		UserAgent:   DefaultUserAgent,
		Retries:     DefaultRetries,
		Client:      &http.Client{Timeout: t},
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 5),
	}
}

func (a *API) SetUserAgent(userAgent string) { a.UserAgent = userAgent }
func (a *API) SetRetries(retries uint)       { a.Retries = retries }

func (a API) CheckCredentials(level RequestLevel) error {
	if level == RequestLevelAPIKey && a.APIKey == "" {
		return NewLastFMError(ErrAPIKeyMissing, APIKeyMissingMessage)
	}
	if a.Client == nil {
		return errors.New("client uninitalized")
	}
	return nil
}

func (a API) Get(dest any, method APIMethod, params any) error {
	return a.Request(dest, http.MethodGet, method, params)
}

func (a API) Post(dest any, method APIMethod, params any) error {
	return a.Request(dest, http.MethodPost, method, params)
}

func (a API) Request(dest any, httpMethod string, method APIMethod, params any) error {
	if err := a.CheckCredentials(RequestLevelAPIKey); err != nil {
		return err
	}

	p, err := lastfm.EncodeToValues(params)
	if err != nil {
		return err
	}

	p.Set("api_key", a.APIKey)
	p.Set("method", string(method))

	switch httpMethod {
	case http.MethodGet:
		return a.GetURL(dest, BuildAPIURL(p))
	case http.MethodPost:
		return a.PostBody(dest, Endpoint, p.Encode())
	default:
		return errors.New("unsupported http method")
	}
}

func (a API) GetURL(dest any, url string) error {
	return a.tryRequest(dest, http.MethodGet, url, "")
}

func (a API) PostBody(dest any, url, body string) error {
	return a.tryRequest(dest, http.MethodPost, url, body)
}

func (a API) tryRequest(dest any, method, url, body string) error {
	if err := a.rateLimiter.Wait(context.Background()); err != nil {
		return err
	}

	var (
		res   *http.Response
		lfm   LFMWrapper
		lferr *LastFMError
		err   error
	)

	for i := uint(0); i <= a.Retries; i++ {
		var req *http.Request
		switch method {
		case http.MethodGet:
			req, err = a.createGetRequest(url)
		case http.MethodPost:
			req, err = a.createPostRequest(url, body)
		default:
			req, err = a.createRequest(method, url, body)
		}
		if err != nil {
			return err
		}

		res, err = a.Client.Do(req)
		if err != nil {
			return err
		}

		err = xml.NewDecoder(res.Body).Decode(&lfm)
		res.Body.Close()
		if err == nil {
			lferr, _ = lfm.UnwrapError()
		}

		if res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests {
			continue
		}
		if lferr != nil && lferr.ShouldRetry() {
			continue
		}
		break
	}

	if lferr != nil {
		return lferr.WrapResponse(res)
	}
	if res.StatusCode < http.StatusOK || res.StatusCode > http.StatusIMUsed {
		return NewHTTPError(res)
	}
	if errors.Is(err, io.EOF) {
		return fmt.Errorf("invalid xml response: %w", err)
	}
	if err != nil {
		return err
	}

	if dest == nil {
		return nil
	}
	if err = lfm.UnmarshalInnerXML(dest); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

func (a API) createGetRequest(url string) (*http.Request, error) {
	return a.createRequest(http.MethodGet, url, "")
}

func (a API) createPostRequest(url, body string) (*http.Request, error) {
	req, err := a.createRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func (a API) createRequest(method, url, body string) (*http.Request, error) {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", a.UserAgent)
	req.Header.Set("Accept", "application/xml")
	return req, nil
}

func BuildAPIURL(params url.Values) string {
	return Endpoint + "?" + params.Encode()
}
