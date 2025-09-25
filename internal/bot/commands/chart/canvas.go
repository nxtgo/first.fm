package chart

import (
	"image"
	"image/color"
	"image/draw"

	"go.fm/internal/bot/image/font"
	"go.fm/internal/bot/image/imgio"
)

func renderChart(entries []Entry, grid int) ([]byte, error) {
	interRegular := font.LoadFont("assets/font/Inter_24pt-Regular.ttf")
	interBold := font.LoadFont("assets/font/Inter_24pt-Bold.ttf")

	labelSize, subSize := 20.0, 16.0
	if grid >= 10 {
		labelSize, subSize = 12, 10
	}

	labelFace := interBold.Face(labelSize, 72)
	subFace := interRegular.Face(subSize, 72)
	labelAscent, subAscent := labelFace.Metrics().Ascent.Ceil(), subFace.Metrics().Ascent.Ceil()

	cell := entries[0].Image.Bounds()
	canvas := image.NewRGBA(image.Rect(0, 0, cell.Dx()*grid, cell.Dy()*grid))

	gradient, err := imgio.Open("assets/img/chart_gradient.png")
	if err != nil {
		return nil, err
	}

	for i, entry := range entries {
		x, y := (i%grid)*cell.Dx(), (i/grid)*cell.Dy()
		rect := image.Rect(x, y, x+cell.Dx(), y+cell.Dy())

		draw.Draw(canvas, rect, entry.Image, image.Point{}, draw.Over)
		draw.Draw(canvas, rect, gradient, image.Point{}, draw.Over)

		font.DrawText(canvas, x+8, y+labelAscent+8, entry.Name, color.White, labelFace)
		if entry.Artist != "" {
			font.DrawText(canvas, x+8, y+labelAscent+subAscent+12,
				entry.Artist, color.RGBA{170, 170, 170, 255}, subFace)
		}
	}

	return imgio.Encode(canvas, imgio.PNGEncoder())
}
