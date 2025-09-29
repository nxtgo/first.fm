package api

import (
	"time"

	"first.fm/internal/cache"
	"first.fm/internal/lastfm"
)

type recentTracksExtendedParams struct {
	lastfm.RecentTracksParams
	Extended bool `url:"extended,int,omitempty"`
}

type User struct {
	api       *API
	InfoCache *cache.Cache[string, *lastfm.UserInfo]
}

// NewUser creates and returns a new User API route.
func NewUser(api *API) *User {
	return &User{
		api:       api,
		InfoCache: cache.New[string, *lastfm.UserInfo](time.Minute*10, 1000),
	}
}

// Friends returns the friends of a user.
func (u *User) Friends(params lastfm.FriendsParams) (*lastfm.Friends, error) {
	var res lastfm.Friends
	return &res, u.api.Get(&res, UserGetFriendsMethod, params)
}

// Info returns the information of a user with caching.
func (u *User) Info(user string) (*lastfm.UserInfo, error) {
	if cached, ok := u.InfoCache.Get(user); ok {
		return cached, nil
	}

	var res lastfm.UserInfo
	p := lastfm.UserInfoParams{User: user}
	err := u.api.Get(&res, UserGetInfoMethod, p)
	if err != nil {
		return nil, err
	}

	u.InfoCache.Set(user, &res)
	return &res, nil
}

// LovedTracks returns the loved tracks of a user.
func (u *User) LovedTracks(params lastfm.LovedTracksParams) (*lastfm.LovedTracks, error) {
	var res lastfm.LovedTracks
	return &res, u.api.Get(&res, UserGetLovedTracksMethod, params)
}

// RecentTrack returns the most recent track of a user. This is a convenience
// method that calls RecentTracks with limit=1.
func (u *User) RecentTrack(user string) (*lastfm.RecentTrack, error) {
	var res lastfm.RecentTrack
	p := lastfm.RecentTracksParams{User: user, Limit: 1}
	return &res, u.api.Get(&res, UserGetRecentTracksMethod, p)
}

// RecentTracks returns the recent tracks of a user.
func (u *User) RecentTracks(params lastfm.RecentTracksParams) (*lastfm.RecentTracks, error) {
	var res lastfm.RecentTracks
	return &res, u.api.Get(&res, UserGetRecentTracksMethod, params)
}

// RecentTrackExtended returns the most recent track of a user with extended
// information. This is a convenience method that calls RecentTracksExtended
// with limit=1.
func (u *User) RecentTrackExtended(user string) (*lastfm.RecentTrackExtended, error) {
	var res lastfm.RecentTrackExtended
	p := lastfm.RecentTracksParams{User: user, Limit: 1}
	exp := recentTracksExtendedParams{RecentTracksParams: p, Extended: true}
	return &res, u.api.Get(&res, UserGetRecentTracksMethod, exp)
}

// RecentTracksExtended returns the recent tracks of a user with extended
// information.
func (u *User) RecentTracksExtended(
	params lastfm.RecentTracksParams) (*lastfm.RecentTracksExtended, error) {

	var res lastfm.RecentTracksExtended
	exp := recentTracksExtendedParams{RecentTracksParams: params, Extended: true}
	return &res, u.api.Get(&res, UserGetRecentTracksMethod, exp)
}

// TopAlbums returns the top albums of a user.
func (u *User) TopAlbums(params lastfm.UserTopAlbumsParams) (*lastfm.UserTopAlbums, error) {
	var res lastfm.UserTopAlbums
	return &res, u.api.Get(&res, UserGetTopAlbumsMethod, params)
}

// TopArtists returns the top artists of a user.
func (u *User) TopArtists(params lastfm.UserTopArtistsParams) (*lastfm.UserTopArtists, error) {
	var res lastfm.UserTopArtists
	return &res, u.api.Get(&res, UserGetTopArtistsMethod, params)
}

// TopTags returns the top tags of a user.
func (u *User) TopTags(params lastfm.UserTopTagsParams) (*lastfm.UserTopTags, error) {
	var res lastfm.UserTopTags
	return &res, u.api.Get(&res, UserGetTopTagsMethod, params)
}

// TopTracks returns the top tracks of a user.
func (u *User) TopTracks(params lastfm.UserTopTracksParams) (*lastfm.UserTopTracks, error) {
	var res lastfm.UserTopTracks
	return &res, u.api.Get(&res, UserGetTopTracksMethod, params)
}

// WeeklyAlbumChart returns the weekly album chart of a user.
func (u *User) WeeklyAlbumChart(
	params lastfm.WeeklyAlbumChartParams) (*lastfm.WeeklyAlbumChart, error) {

	var res lastfm.WeeklyAlbumChart
	return &res, u.api.Get(&res, UserGetWeeklyAlbumChartMethod, params)
}

// WeeklyArtistChart returns the weekly artist chart of a user.
func (u *User) WeeklyArtistChart(
	params lastfm.WeeklyArtistChartParams) (*lastfm.WeeklyArtistChart, error) {

	var res lastfm.WeeklyArtistChart
	return &res, u.api.Get(&res, UserGetWeeklyArtistChartMethod, params)
}

// WeeklyChartList returns the weekly chart list of a user.
func (u *User) WeeklyChartList(user string) (*lastfm.WeeklyChartList, error) {
	var res lastfm.WeeklyChartList
	p := lastfm.WeeklyChartListParams{User: user}
	return &res, u.api.Get(&res, UserGetWeeklyChartListMethod, p)
}

// WeeklyTrackChart returns the weekly track chart of a user.
func (u *User) WeeklyTrackChart(
	params lastfm.WeeklyTrackChartParams) (*lastfm.WeeklyTrackChart, error) {

	var res lastfm.WeeklyTrackChart
	return &res, u.api.Get(&res, UserGetWeeklyTrackChartMethod, params)
}
