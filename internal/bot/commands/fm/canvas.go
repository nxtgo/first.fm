package fm

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"

	"github.com/nxtgo/arikawa/v3/utils/sendpart"
	"go.fm/internal/bot/discord/reply"
	"go.fm/internal/bot/image/blur"
	"go.fm/internal/bot/image/font"
	"go.fm/internal/bot/image/imgio"
	"go.fm/internal/bot/image/mask"
	"go.fm/internal/bot/image/transform"
	"go.fm/internal/bot/lastfm"
)

var (
	fmWidth  = 555
	fmHeight = 147

	titleX  = 20
	titleY  = 26
	artistX = 20
	artistY = 84

	coverY        = 0
	coverWidth    = 313
	coverHeight   = 147
	textSampleW   = 400
	titleSampleH  = 50
	artistSampleH = 30
)

func renderCanvas(edit *reply.EditBuilder, track *lastfm.RecentTrack) error {
	interBold := font.LoadFont("assets/font/Inter_24pt-Bold.ttf")
	titleFace := interBold.Face(48, 72)
	artistFace := interBold.Face(24, 72)

	canvas := image.NewNRGBA(image.Rect(0, 0, fmWidth, fmHeight))

	coverImage, err := imgio.FromUrl(track.GetLargestImage().URL)
	if err != nil {
		return err
	}

	// blur background
	blurredCover := transform.Resize(coverImage, fmWidth, fmHeight, transform.Gaussian)
	blurredCover = blur.Gaussian(blurredCover, 30)
	draw.Draw(canvas, canvas.Bounds(), blurredCover, image.Point{}, draw.Over)

	// mask
	sharpCover := transform.Resize(coverImage, coverWidth, coverHeight, transform.Gaussian)
	gradientMask := mask.GradientHorizontal(coverWidth, coverHeight, true)
	coverEndX := fmWidth - coverWidth
	draw.DrawMask(canvas,
		image.Rect(coverEndX, coverY, coverEndX+coverWidth, coverY+coverHeight),
		sharpCover,
		image.Point{},
		gradientMask,
		image.Point{},
		draw.Over,
	)

	// text
	titleColor := getContrastColor(canvas, titleX, titleY, textSampleW, titleSampleH)
	artistColor := getContrastColor(canvas, artistX, artistY, textSampleW, artistSampleH)

	titleAscent := titleFace.Metrics().Ascent.Ceil()
	font.DrawText(canvas, titleX, titleY+titleAscent, track.Name, titleColor, titleFace)

	artistAscent := artistFace.Metrics().Ascent.Ceil()
	font.DrawText(canvas, artistX, artistY+artistAscent, track.Artist.Name, artistColor, artistFace)

	// rounded corners
	roundedMask := mask.Rounded(fmWidth, fmHeight, 20)
	final := image.NewNRGBA(canvas.Bounds())
	draw.DrawMask(final, final.Bounds(), canvas, image.Point{}, roundedMask, image.Point{}, draw.Over)

	result, err := imgio.Encode(final, imgio.PNGEncoder())
	if err != nil {
		return err
	}
	_, err = edit.File(sendpart.File{Name: "profile.png", Reader: bytes.NewReader(result)}).Send()
	return err
}

func getContrastColor(img image.Image, x, y, w, h int) color.Color {
	var total float64
	count := 0

	bounds := img.Bounds()
	for iy := y; iy < y+h && iy < bounds.Max.Y; iy++ {
		for ix := x; ix < x+w && ix < bounds.Max.X; ix++ {
			r, g, b, _ := img.At(ix, iy).RGBA()
			lum := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			total += lum
			count++
		}
	}

	if count == 0 {
		return color.White
	}

	avgLum := total / float64(count)
	if avgLum > 128 {
		return color.Black
	}
	return color.White
}
