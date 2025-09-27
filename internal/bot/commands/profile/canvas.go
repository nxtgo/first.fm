package profile

import (
	"bytes"
	"image"
	"image/color"

	"github.com/nxtgo/arikawa/v3/utils/sendpart"
	"go.fm/internal/bot/discord/reply"
	"go.fm/internal/bot/image/imgio"
	"go.fm/internal/bot/image/shapes"
	"go.fm/internal/bot/lastfm"
)

var (
	canvasWidth  = 760
	canvasHeight = 260

	bgColor    = color.RGBA{211, 211, 211, 255} // #D3D3D3
	whiteColor = color.RGBA{255, 255, 255, 255} // #FFFFFF
)

func renderCanvas(edit *reply.EditBuilder, user *lastfm.User) error {
	canvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))

	// Background
	shapes.DrawRoundedRectangle(canvas, 0, 0, canvasWidth, canvasHeight, 30, bgColor)

	// Avatar panel
	shapes.DrawRectangle(canvas, 28, 0, 165, 260, whiteColor)
	shapes.DrawRoundedRectangle(canvas, 45, 20, 131, 131, 20, bgColor)

	// Reputation
	shapes.DrawRoundedRectangle(canvas, 45, 168, 131, 41, 11, bgColor)

	// User flag
	shapes.DrawRoundedRectangle(canvas, 208, 38, 50, 32, 11, whiteColor)

	// Server icon
	shapes.DrawCircle(canvas, 234, 126, 17, bgColor)

	// Progress bar
	shapes.DrawRoundedRectangle(canvas, 290, 110, 463, 33, 11, bgColor)

	// Badge panel
	shapes.DrawRoundedRectangle(canvas, 278, 191, 463, 51, 11, bgColor)

	// Encode and send
	result, err := imgio.Encode(canvas, imgio.PNGEncoder())
	if err != nil {
		return err
	}

	_, err = edit.File(sendpart.File{Name: "profile.png", Reader: bytes.NewReader(result)}).Send()
	return err
}
