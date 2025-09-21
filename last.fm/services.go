package lastfm

import "time"

// Services groups all Last.fm API services
type Services struct {
	User *UserService
}

// NewServices creates a new services group with all available services
func NewServices(k string, c *Cache) *Services {
	client := NewClient(k, WithTimeout(time.Second*10), WithCache(c))
	return &Services{
		User: NewUserService(client),
	}
}
