package profile

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	"go.fm/types/cmd"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "profile",
		Description: "display a last.fm user info",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
		Options: []discord.ApplicationCommandOption{
			cmd.UserOption,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := ctx.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = ctx.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	username, err := ctx.GetUser(e)
	if err != nil {
		_ = ctx.Error(e, constants.ErrorNotRegistered)
		return
	}

	user, err := ctx.LastFM.GetUserInfo(username)
	if err != nil {
		_ = ctx.Error(e, constants.ErrorUserNotFound)
		return
	}

	realName := user.User.Realname
	if realName == "" {
		realName = user.User.Name
	}

	favTrack := "none"
	favTrackURL := ""
	topTracks, err := ctx.LastFM.GetTopTracks(username, 1)
	if err == nil && len(topTracks.TopTracks.Track) > 0 {
		favTrack = topTracks.TopTracks.Track[0].Name
		favTrackURL = topTracks.TopTracks.Track[0].URL
	}

	favArtist := "none"
	favArtistURL := ""
	topArtists, err := ctx.LastFM.GetTopArtists(username, 1)
	if err == nil && len(topArtists.TopArtists.Artist) > 0 {
		favArtist = topArtists.TopArtists.Artist[0].Name
		favArtistURL = topArtists.TopArtists.Artist[0].URL
	}

	favAlbum := "none"
	favAlbumURL := ""
	topAlbums, err := ctx.LastFM.GetTopAlbums(username, 1)
	if err == nil && len(topAlbums.TopAlbums.Album) > 0 {
		favAlbum = topAlbums.TopAlbums.Album[0].Name
		favAlbumURL = topAlbums.TopAlbums.Album[0].URL
	}

	avatar := user.User.Image[len(user.User.Image)-1].Text
	if avatar == "" {
		avatar = "https://lastfm.freetls.fastly.net/i/u/avatar170s/818148bf682d429dc215c1705eb27b98.png"
	}
	if dot := strings.LastIndex(avatar, "."); dot != -1 {
		avatar = avatar[:dot] + ".gif"
	}

	component := discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplayf("## [%s](%s)", realName, user.User.URL),
			discord.NewTextDisplayf("-# *__@%s__*\nsince <t:%s:D>", user.User.Name, user.User.Registered.Unixtime),
			discord.NewTextDisplayf("**%s** total scrobbles", user.User.Playcount),
		).WithAccessory(discord.NewThumbnail(avatar)),
		discord.NewSmallSeparator(),
		discord.NewTextDisplay(
			fmt.Sprintf("-# *Favorite album* \\💿\n[**%s**](%s)\n", favAlbum, favAlbumURL)+
				fmt.Sprintf("-# *Favorite artist* \\🎤\n[**%s**](%s)\n", favArtist, favArtistURL)+
				fmt.Sprintf("-# *Favorite track* \\🎵\n[**%s**](%s)\n", favTrack, favTrackURL),
		),
		discord.NewSmallSeparator(),
		discord.NewTextDisplayf(
			"\\🎤 **%s** artists\n\\💿 **%s** albums\n\\🎵 **%s** unique tracks",
			user.User.ArtistCount,
			user.User.AlbumCount,
			user.User.TrackCount,
		),
	).WithAccentColor(0x00ADD8)

	reply.Flags(discord.MessageFlagIsComponentsV2).Component(component).Edit()
}
