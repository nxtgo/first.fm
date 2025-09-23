package lastfm

import (
	"encoding/xml"
	"fmt"
)

type UserService struct {
	client *Client
}

func NewUserService(client *Client) *UserService {
	return &UserService{
		client: client,
	}
}

type userInfoResponse struct {
	XMLName xml.Name `xml:"lfm"`
	Status  string   `xml:"status,attr"`
	User    User     `xml:"user"`
}

type recentTracksResponse struct {
	XMLName      xml.Name     `xml:"lfm"`
	Status       string       `xml:"status,attr"`
	RecentTracks RecentTracks `xml:"recenttracks"`
}

type topAlbumsResponse struct {
	XMLName   xml.Name  `xml:"lfm"`
	Status    string    `xml:"status,attr"`
	TopAlbums TopAlbums `xml:"topalbums"`
}

type topArtistsResponse struct {
	XMLName    xml.Name   `xml:"lfm"`
	Status     string     `xml:"status,attr"`
	TopArtists TopArtists `xml:"topartists"`
}

type topTracksResponse struct {
	XMLName   xml.Name  `xml:"lfm"`
	Status    string    `xml:"status,attr"`
	TopTracks TopTracks `xml:"toptracks"`
}

/// user-based methods

func (s *UserService) GetInfo(params P) (*User, error) {
	if params["user"] == "" {
		return nil, fmt.Errorf("user parameter is required")
	}

	key := GenerateCacheKey("user.getinfo", params)

	if user, cached := s.client.Cache.User.Get(key); cached {
		return &user, nil
	}

	body, err := s.client.makeRequest("user.getinfo", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	var response userInfoResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	s.client.Cache.User.Set(key, response.User, 0)

	return &response.User, nil
}

/// user tracks

func (s *UserService) GetRecentTracks(params P) (*RecentTracks, error) {
	if params["user"] == "" {
		return nil, fmt.Errorf("user parameter is required")
	}

	body, err := s.client.makeRequest("user.getrecenttracks", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent tracks: %w", err)
	}

	var response recentTracksResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.RecentTracks, nil
}

/// user tops

func (s *UserService) GetTopAlbums(params P) (*TopAlbums, error) {
	if params["user"] == "" {
		return nil, fmt.Errorf("user parameter is required")
	}

	key := GenerateCacheKey("user.gettopalbums", params)

	if topAlbums, cached := s.client.Cache.UserTopAlbums.Get(key); cached {
		return &topAlbums, nil
	}

	body, err := s.client.makeRequest("user.gettopalbums", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get top albums: %w", err)
	}

	var response topAlbumsResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	s.client.Cache.UserTopAlbums.Set(key, response.TopAlbums, 0)

	return &response.TopAlbums, nil
}

func (s *UserService) GetTopArtists(params P) (*TopArtists, error) {
	if params["user"] == "" {
		return nil, fmt.Errorf("user parameter is required")
	}

	key := GenerateCacheKey("user.gettopartists", params)

	if topArtists, cached := s.client.Cache.UserTopArtists.Get(key); cached {
		return &topArtists, nil
	}

	body, err := s.client.makeRequest("user.gettopartists", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get top artists: %w", err)
	}

	var response topArtistsResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	s.client.Cache.UserTopArtists.Set(key, response.TopArtists, 0)

	return &response.TopArtists, nil
}

func (s *UserService) GetTopTracks(params P) (*TopTracks, error) {
	if params["user"] == "" {
		return nil, fmt.Errorf("user parameter is required")
	}

	key := GenerateCacheKey("user.gettoptracks", params)

	if topTracks, cached := s.client.Cache.UserTopTracks.Get(key); cached {
		return &topTracks, nil
	}

	body, err := s.client.makeRequest("user.gettoptracks", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get top tracks: %w", err)
	}

	var response topTracksResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	s.client.Cache.UserTopTracks.Set(key, response.TopTracks, 0)

	return &response.TopTracks, nil
}
