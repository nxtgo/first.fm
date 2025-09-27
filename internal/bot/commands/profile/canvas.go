package profile

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/nxtgo/arikawa/v3/utils/sendpart"
	"go.fm/internal/bot/discord/reply"
	"go.fm/internal/bot/image/font"
	"go.fm/internal/bot/image/imgio"
	"go.fm/internal/bot/image/mask"
	"go.fm/internal/bot/image/transform"
	"go.fm/internal/bot/lastfm"
)

var (
	profileWidth  = 740
	profileHeight = 260

	avatarX      = 43
	avatarY      = 21
	avatarWidth  = 130
	avatarHeight = 130

	containerX = 45
	containerY = 169
	containerW = 125
	containerH = 40

	realNameX = 200
	realNameY = 21
)

func renderCanvas(edit *reply.EditBuilder, user *lastfm.User) error {
	interBold := font.LoadFont("assets/font/Inter_24pt-Bold.ttf")
	//interRegular := font.LoadFont("assets/font/Inter_24pt-Regular.ttf")

	scrobblesFace := interBold.Face(20, 72)
	realNameFace := interBold.Face(24, 72)
	// usernameFace := interRegular.Face(15, 72)

	canvas := image.NewRGBA(image.Rect(0, 0, profileWidth, profileHeight))
	layout, err := imgio.Open("assets/img/layouts/profile.png")
	if err != nil {
		return err
	}
	avatar, err := imgio.FromUrl(user.GetLargestImage().URL)
	if err != nil {
		return err
	}

	// avatar
	avatar = transform.Resize(avatar, avatarWidth, avatarHeight, transform.NearestNeighbor)
	avatarMask := mask.Rounded(avatarWidth, avatarHeight, 15)
	avatarRect := image.Rect(
		avatarX,
		avatarY,
		avatarX+avatarWidth,
		avatarY+avatarHeight,
	)

	// draw stuff onto the canvas
	draw.Draw(canvas, image.Rect(0, 0, profileWidth, profileHeight), layout, image.Point{0, 0}, draw.Over)
	draw.DrawMask(canvas, avatarRect, avatar, image.Point{0, 0}, avatarMask, image.Point{0, 0}, draw.Over)

	// scrobbles text
	scrobbles := fmt.Sprintf("%d", user.GetPlayCount())
	faceAscent := scrobblesFace.Metrics().Ascent.Ceil()
	faceDescent := scrobblesFace.Metrics().Descent.Ceil()

	textWidth := font.Measure(scrobblesFace, scrobbles)
	textHeight := faceAscent + faceDescent

	textX := containerX + (containerW-textWidth)/2
	textY := containerY + (containerH-textHeight)/2 + faceAscent

	font.DrawText(canvas, textX, textY, scrobbles, color.White, scrobblesFace)

	// real name text
	realNameAscent := realNameFace.Metrics().Ascent.Ceil()
	realName := user.RealName
	if realName == "" {
		realName = user.Name
	}

	font.DrawText(canvas, realNameX, realNameY+realNameAscent, realName, color.White, realNameFace)

	result, err := imgio.Encode(canvas, imgio.PNGEncoder())
	if err != nil {
		return err
	}

	_, err = edit.File(sendpart.File{Name: "profile.png", Reader: bytes.NewReader(result)}).Send()
	return nil
}
