package profile

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/lfm"
	"go.fm/lfm/types"
	"go.fm/pkg/bild/colors"
	"go.fm/pkg/constants/emojis"
	"go.fm/pkg/constants/errs"
	"go.fm/pkg/constants/opts"
	"go.fm/pkg/ctx"
	"go.fm/pkg/discord/reply"
)

type Command struct{}
type Fav struct {
	Name      string
	URL       string
	PlayCount string
}

func fetchFav[T any](fetch func() (T, error), extract func(T) Fav) Fav {
	data, err := fetch()
	if err != nil {
		return Fav{"none", "", "0"}
	}
	return extract(data)
}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "profile",
		Description: "display a last.fm user info",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
		Options: []discord.ApplicationCommandOption{
			opts.UserOption,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx ctx.CommandContext) {
	r := reply.New(e)
	if err := r.Defer(); err != nil {
		reply.Error(e, errs.ErrCommandDeferFailed)
		return
	}

	username, err := ctx.GetUser(e)
	if err != nil {
		reply.Error(e, errs.ErrUserNotFound)
		return
	}

	user, err := ctx.LastFM.User.GetInfoWithPrefetch(lfm.P{"user": username})
	if err != nil {
		reply.Error(e, errs.ErrUserNotFound)
		return
	}

	realName := user.RealName
	if realName == "" {
		realName = user.Name
	}

	favTrack := fetchFav(
		func() (*types.UserGetTopTracks, error) {
			return ctx.LastFM.User.GetTopTracks(lfm.P{"user": username, "limit": 1})
		},
		func(tt *types.UserGetTopTracks) Fav {
			if len(tt.Tracks) == 0 {
				return Fav{"none", "", "0"}
			}
			t := tt.Tracks[0]
			return Fav{t.Name, t.Url, t.PlayCount}
		},
	)

	favArtist := fetchFav(
		func() (*types.UserGetTopArtists, error) {
			return ctx.LastFM.User.GetTopArtists(lfm.P{"user": username, "limit": 1})
		},
		func(ta *types.UserGetTopArtists) Fav {
			if len(ta.Artists) == 0 {
				return Fav{"none", "", "0"}
			}
			a := ta.Artists[0]
			return Fav{a.Name, a.Url, a.PlayCount}
		},
	)

	favAlbum := fetchFav(
		func() (*types.UserGetTopAlbums, error) {
			return ctx.LastFM.User.GetTopAlbums(lfm.P{"user": username, "limit": 1})
		},
		func(ta *types.UserGetTopAlbums) Fav {
			if len(ta.Albums) == 0 {
				return Fav{"none", "", "0"}
			}
			a := ta.Albums[0]
			return Fav{a.Name, a.Url, a.PlayCount}
		},
	)

	avatar := ""
	if len(user.Images) > 0 {
		avatar = user.Images[len(user.Images)-1].Url
	}
	if avatar == "" {
		avatar = "https://lastfm.freetls.fastly.net/i/u/avatar170s/818148bf682d429dc215c1705eb27b98.png"
	}

	color := 0x00ADD8
	if dominantColor, err := colors.Dominant(avatar); err == nil {
		color = dominantColor
	}

	component := discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplayf("## [%s](%s)", realName, user.Url),
			discord.NewTextDisplayf("-# *__@%s__*\nSince <t:%s:D> %s", user.Name, user.Registered.Unixtime, emojis.EmojiCalendar),
			discord.NewTextDisplayf("**%s** total scrobbles %s", user.PlayCount, emojis.EmojiPlay),
		).WithAccessory(discord.NewThumbnail(avatar)),
		discord.NewSmallSeparator(),
		discord.NewTextDisplay(
			fmt.Sprintf("-# *Favorite album* %s\n[**%s**](%s) — %s plays\n", emojis.EmojiAlbum, favAlbum.Name, favAlbum.URL, favAlbum.PlayCount)+
				fmt.Sprintf("-# *Favorite artist* %s\n[**%s**](%s) — %s plays\n", emojis.EmojiMic2, favArtist.Name, favArtist.URL, favArtist.PlayCount)+
				fmt.Sprintf("-# *Favorite track* %s\n[**%s**](%s) — %s plays\n", emojis.EmojiNote, favTrack.Name, favTrack.URL, favTrack.PlayCount),
		),
		discord.NewSmallSeparator(),
		discord.NewTextDisplayf(
			"%s **%s** albums\n%s **%s** artists\n%s **%s** unique tracks",
			emojis.EmojiAlbum,
			user.ArtistCount,
			emojis.EmojiMic2,
			user.AlbumCount,
			emojis.EmojiNote,
			user.TrackCount,
		),
	).WithAccentColor(color)

	r.Flags(discord.MessageFlagIsComponentsV2).Component(component).Edit()
}
