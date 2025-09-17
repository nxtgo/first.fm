package test

// an *over*engineering masterpiece

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"runtime"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/lfm"
	"go.fm/lfm/types"
	"go.fm/pkg/bild/blur"
	"go.fm/pkg/bild/font"
	"go.fm/pkg/bild/imgio"
	"go.fm/pkg/bild/mask"
	"go.fm/pkg/bild/transform"
	"go.fm/pkg/constants/errs"
	"go.fm/pkg/constants/opts"
	"go.fm/pkg/ctx"
	"go.fm/pkg/discord/reply"
	"go.fm/pkg/strng"
)

var inter *font.Font

func init() {
	inter = font.LoadFont("assets/font/Inter_24pt-Regular.ttf")
}

type Command struct{}

type Fav struct {
	Name      string
	PlayCount string
}

func fetchFav[T any](fetch func() (T, error), extract func(T) Fav) Fav {
	data, err := fetch()
	if err != nil {
		return Fav{"none", "0"}
	}
	return extract(data)
}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "test",
		Description: "this is a test command",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
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
		reply.Error(e, errs.ErrUserNotRegistered)
		return
	}

	user, err := ctx.LastFM.User.GetInfo(lfm.P{"user": username})
	if err != nil {
		reply.Error(e, err)
		return
	}

	favTrack := fetchFav(
		func() (*types.UserGetTopTracks, error) {
			return ctx.LastFM.User.GetTopTracks(lfm.P{"user": username, "limit": 1})
		},
		func(tt *types.UserGetTopTracks) Fav {
			if len(tt.Tracks) == 0 {
				return Fav{"none", "0"}
			}
			t := tt.Tracks[0]
			return Fav{t.Name, t.PlayCount}
		},
	)

	favArtist := fetchFav(
		func() (*types.UserGetTopArtists, error) {
			return ctx.LastFM.User.GetTopArtists(lfm.P{"user": username, "limit": 1})
		},
		func(ta *types.UserGetTopArtists) Fav {
			if len(ta.Artists) == 0 {
				return Fav{"none", "0"}
			}
			a := ta.Artists[0]
			return Fav{a.Name, a.PlayCount}
		},
	)

	favAlbum := fetchFav(
		func() (*types.UserGetTopAlbums, error) {
			return ctx.LastFM.User.GetTopAlbums(lfm.P{"user": username, "limit": 1})
		},
		func(ta *types.UserGetTopAlbums) Fav {
			if len(ta.Albums) == 0 {
				return Fav{"none", "0"}
			}
			a := ta.Albums[0]
			return Fav{a.Name, a.PlayCount}
		},
	)

	runtime.GC()
	var mStart, mEnd runtime.MemStats
	runtime.ReadMemStats(&mStart)

	// - start memory measure -
	gradientData, err := imgio.Open("assets/img/profile_gradient.png")
	if err != nil {
		reply.Error(e, fmt.Errorf("failed to load gradient background: %w", err))
		return
	}

	userAvatar := user.Images[len(user.Images)-1].Url
	data, err := imgio.Fetch(userAvatar)
	if err != nil {
		reply.Error(e, err)
		return
	}

	avatarImage, err := imgio.Decode(data)
	if err != nil {
		reply.Error(e, err)
		return
	}

	canvasWidth, canvasHeight := 500, 600
	avatarSize := 180
	avatarPadding := image.Pt(20, 20)
	radius := 10

	canvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))

	bgImage := transform.Resize(avatarImage, canvasWidth, canvasHeight, transform.Linear)
	bgImage = blur.Gaussian(bgImage, 20)
	draw.Draw(canvas, canvas.Bounds(), bgImage, image.Point{0, 0}, draw.Over)

	gradient := transform.Resize(gradientData, canvasWidth, canvasHeight, transform.Linear)
	draw.Draw(canvas, canvas.Bounds(), gradient, image.Point{0, 0}, draw.Over)

	avatarImage = transform.Resize(avatarImage, avatarSize, avatarSize, transform.Gaussian)
	mask := mask.Rounded(avatarSize, avatarSize, radius)

	draw.DrawMask(
		canvas,
		image.Rect(avatarPadding.X, avatarPadding.Y, avatarPadding.X+avatarSize, avatarPadding.Y+avatarSize),
		avatarImage,
		image.Point{0, 0},
		mask,
		image.Point{0, 0},
		draw.Over,
	)

	// real name (if exists)
	realName := user.RealName
	if realName == "" {
		realName = user.Name
	}

	face32 := inter.Face(32, 72)
	metrics32 := face32.Metrics()
	ascent32 := metrics32.Ascent.Ceil()
	textX := avatarPadding.X + avatarSize + 20
	textY1 := avatarPadding.Y + ascent32
	font.DrawText(canvas, textX, textY1, realName, color.White, face32)

	// @username
	face16 := inter.Face(16, 72)
	textY2 := textY1 + face32.Metrics().Height.Ceil() - 10
	font.DrawText(canvas, textX, textY2, fmt.Sprintf("@%s", user.Name), color.White, face16)

	// below avatar btw
	labelFace := inter.Face(20, 72)
	valueFace := inter.Face(26, 72)
	spacing := 6
	infoStartY := avatarPadding.Y + avatarSize + 35

	// colors
	labelColor := color.RGBA{170, 170, 170, 255}
	valueColor := color.White

	// nums
	artistNum := favArtist.PlayCount
	trackNum := favTrack.PlayCount
	albumNum := favAlbum.PlayCount

	// font width
	wArtist := font.Measure(valueFace, artistNum)
	wTrack := font.Measure(valueFace, trackNum)
	wAlbum := font.Measure(valueFace, albumNum)

	maxNumW := wArtist
	maxNumW = max(wTrack, maxNumW)
	maxNumW = max(wAlbum, maxNumW)

	numGap := 12
	xBase := avatarPadding.X
	y := infoStartY

	// favourite artist
	font.DrawText(canvas, xBase, y, "Favourite artist", labelColor, labelFace)
	valY := y + labelFace.Metrics().Height.Ceil() + spacing
	font.DrawText(canvas, xBase, valY, artistNum, labelColor, valueFace)
	font.DrawText(canvas, xBase+maxNumW+numGap, valY, strng.Truncate(favArtist.Name, 25), valueColor, valueFace)

	// favourite track
	y = valY + valueFace.Metrics().Height.Ceil() + spacing
	font.DrawText(canvas, xBase, y, "Favourite track", labelColor, labelFace)
	valY = y + labelFace.Metrics().Height.Ceil() + spacing
	font.DrawText(canvas, xBase, valY, trackNum, labelColor, valueFace)
	font.DrawText(canvas, xBase+maxNumW+numGap, valY, strng.Truncate(favTrack.Name, 25), valueColor, valueFace)

	// favourite album
	y = valY + valueFace.Metrics().Height.Ceil() + spacing
	font.DrawText(canvas, xBase, y, "Favourite album", labelColor, labelFace)
	valY = y + labelFace.Metrics().Height.Ceil() + spacing
	font.DrawText(canvas, xBase, valY, albumNum, labelColor, valueFace)
	font.DrawText(canvas, xBase+maxNumW+numGap, valY, strng.Truncate(favAlbum.Name, 25), valueColor, valueFace)

	// :D
	result, err := imgio.Encode(canvas, imgio.PNGEncoder())
	if err != nil {
		reply.Error(e, err)
		return
	}
	// - end memory measure -

	runtime.ReadMemStats(&mEnd)

	file := discord.NewFile("test.png", "", bytes.NewReader(result))
	r.File(file).Content("ready. (wip command, testing purposes)\n-# *used %vmb*", bToMb(mEnd.Alloc-mStart.Alloc)).Edit()
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
