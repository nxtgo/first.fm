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

func (s *UserService) GetInfo(params P) (*User, error) {
	if params["user"] == "" {
		return nil, fmt.Errorf("user parameter is required")
	}

	body, err := s.client.makeRequest("user.getinfo", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	var response userInfoResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.User, nil
}

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
