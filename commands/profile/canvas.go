package profile

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/nxtgo/arikawa/v3/utils/sendpart"
	"go.fm/bild/blur"
	"go.fm/bild/font"
	"go.fm/bild/imgio"
	"go.fm/bild/mask"
	"go.fm/bild/transform"
	lastfm "go.fm/last.fm"
	"go.fm/pkg/reply"
)

var (
	profileWidth  = 740
	profileHeight = 260

	avatarX      = 23
	avatarY      = 45
	avatarWidth  = 120
	avatarHeight = 120

	containerX = 23
	containerY = 187
	containerW = 120
	containerH = 21
)

func renderCanvas(edit *reply.EditBuilder, user *lastfm.User) error {
	inter := font.LoadFont("assets/font/Inter_24pt-Bold.ttf")
	face := inter.Face(14, 72)

	canvas := image.NewRGBA(image.Rect(0, 0, profileWidth, profileHeight))
	layout, err := imgio.Open("assets/img/profile_layout.png")
	if err != nil {
		return err
	}
	avatar, err := imgio.FromUrl(user.GetLargestImage().URL)
	if err != nil {
		return err
	}

	// background
	background := transform.Resize(avatar, profileWidth, profileHeight, transform.NearestNeighbor)
	background = blur.Gaussian(background, 30)
	backgroundMask := mask.Rounded(profileWidth, profileHeight, 30)

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
	draw.DrawMask(canvas, image.Rect(0, 0, profileWidth, profileHeight), background, image.Point{0, 0}, backgroundMask, image.Point{0, 0}, draw.Over)
	draw.Draw(canvas, image.Rect(0, 0, profileWidth, profileHeight), layout, image.Point{0, 0}, draw.Over)
	draw.DrawMask(canvas, avatarRect, avatar, image.Point{0, 0}, avatarMask, image.Point{0, 0}, draw.Over)

	// text
	scrobbles := fmt.Sprintf("%d", user.GetPlayCount())
	faceAscent := face.Metrics().Ascent.Ceil()
	faceDescent := face.Metrics().Descent.Ceil()

	textWidth := font.Measure(face, scrobbles)
	textHeight := faceAscent + faceDescent

	textX := containerX + (containerW-textWidth)/2
	textY := containerY + (containerH-textHeight)/2 + faceAscent

	// draw the text onto the canvas
	font.DrawText(canvas, textX, textY, scrobbles, color.White, face)

	result, err := imgio.Encode(canvas, imgio.PNGEncoder())
	if err != nil {
		return err
	}

	_, err = edit.File(sendpart.File{Name: "profile.png", Reader: bytes.NewReader(result)}).Send()
	return nil
}
