package shapes

import (
	"image"
	"image/color"
	"image/draw"

	"go.fm/internal/bot/image/mask"
)

// DrawCircle draws a filled circle at (cx, cy) with radius r
func DrawCircle(img *image.RGBA, cx, cy, r int, col color.Color) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			if (x-cx)*(x-cx)+(y-cy)*(y-cy) <= r*r {
				img.Set(x, y, col)
			}
		}
	}
}

// DrawRectangle draws a filled rectangle
func DrawRectangle(img *image.RGBA, x, y, w, h int, col color.Color) {
	draw.Draw(img, image.Rect(x, y, x+w, y+h), &image.Uniform{col}, image.Point{}, draw.Src)
}

// DrawRoundedRectangle draws a rectangle with rounded corners
func DrawRoundedRectangle(img *image.RGBA, x, y, w, h, radius int, col color.Color) {
	maskImg := mask.Rounded(w, h, radius)
	draw.DrawMask(img, image.Rect(x, y, x+w, y+h), &image.Uniform{col}, image.Point{}, maskImg, image.Point{}, draw.Over)
}
