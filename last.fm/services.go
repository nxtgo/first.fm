package lastfm

// Services groups all Last.fm API services
type Services struct {
	User *UserService
}

// NewServices creates a new services group with all available services
func NewServices(client *Client) *Services {
	return &Services{
		User: NewUserService(client),
	}
}

// NewServicesWithAPIKey is a convenience function to create services with just an API key
func NewServicesWithAPIKey(apiKey string, options ...ClientOption) *Services {
	client := NewClient(apiKey, options...)
	return NewServices(client)
}
