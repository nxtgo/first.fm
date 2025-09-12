package constants

var (
	ErrorAcknowledgeCommand = "failed to acknowledge command"
	ErrorNotRegistered      = "you didn't set your last.fm username, use `/set-user`"
	ErrorUserNotFound       = "couldn't find that user in last.fm"
	ErrorAlreadyLinked      = "that last.fm username is already linked to another Discord user"
	ErrorUsernameAlreadySet = "your username is already set to **%s**"
	ErrorSetUsername        = "failed to set your last.fm username"
	ErrorUnexpected         = "an unexpected error happened, try again later"
	ErrorGetUser            = "could not get your last.fm username, use `/set-user`"
	ErrorFetchCurrentTrack  = "could not fetch your current track/artist/album"
	ErrorNotPlaying         = "this user is not currently playing any track"
	ErrorNoListeners        = "no one has listened to this yet"
	ErrorNoTracks           = "no tracks found, oh no"
)
