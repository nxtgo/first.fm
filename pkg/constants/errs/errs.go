package errs

import (
	"errors"
	"fmt"
)

var (
	// generic
	ErrUserNotFound       = errors.New("i coulnd't find that user")
	ErrUserNotRegistered  = errors.New("you need to set your last.fm username. use `/set-user` to get started")
	ErrCommandDeferFailed = errors.New("failed to acknowledge command")
	ErrCurrentTrackFetch  = errors.New("couldn't fetch user's current track")
	ErrNoTracksFound      = errors.New("no tracks were found")
	ErrUnexpected         = errors.New("an unexpected error occurred, try again or visit the support server")
	ErrNotListening       = errors.New("this user is not listening to anything right now")
	ErrNoListeners        = errors.New("no one has listened to this, *yet*")

	// specific
	ErrUsernameAlreadyUsed = errors.New("this username is already in use by another Discord user")
	ErrSetUsername         = errors.New("i couldn't set your last.fm username")
	ErrUsernameAlreadySet  = func(username string) error {
		return fmt.Errorf("your username is already set to %s", username)
	}
)
