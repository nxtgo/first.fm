package lastfm

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURL = "https://ws.audioscrobbler.com/2.0/"
)

// Params.......
type P map[string]any

// Client represents the Last.fm API client
type Client struct {
	APIKey     string
	HTTPClient *http.Client
	BaseURL    string
	Cache      *Cache
}

// ClientOption represents a configuration option for the client
type ClientOption func(*Client)

// NewClient creates a new Last.fm API client
func NewClient(apiKey string, options ...ClientOption) *Client {
	client := &Client{
		APIKey:  apiKey,
		BaseURL: BaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Cache: nil,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// WithCache sets a custom cache client
func WithCache(cache *Cache) ClientOption {
	return func(c *Client) {
		c.Cache = cache
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithTimeout sets a custom timeout for HTTP requests
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.HTTPClient.Timeout = timeout
	}
}

// WithBaseURL sets a custom base URL (useful for testing)
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// buildURL constructs the API URL with parameters
func (c *Client) buildURL(method string, params P) string {
	u, _ := url.Parse(c.BaseURL)
	q := u.Query()

	q.Set("method", method)
	q.Set("api_key", c.APIKey)

	for key, value := range params {
		if value != "" {
			q.Set(key, fmt.Sprintf("%v", value))
		}
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// makeRequest performs an HTTP GET request to the Last.fm API
func (c *Client) makeRequest(method string, params P) ([]byte, error) {
	url := c.buildURL(method, params)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var errorResp struct {
		XMLName xml.Name `xml:"lfm"`
		Status  string   `xml:"status,attr"`
		Error   struct {
			Code int    `xml:"code,attr"`
			Text string `xml:",chardata"`
		} `xml:"error"`
	}

	if err := xml.Unmarshal(body, &errorResp); err == nil && errorResp.Status == "failed" {
		return nil, fmt.Errorf("last.fm error %d: %s", errorResp.Error.Code, errorResp.Error.Text)
	}

	return body, nil
}
