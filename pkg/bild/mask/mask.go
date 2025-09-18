package mask

import (
	"image"
	"image/color"

	"go.fm/pkg/bild/parallel"
)

func Rounded(width, height, radius int) *image.Alpha {
	mask := image.NewAlpha(image.Rect(0, 0, width, height))

	parallel.Line(height, func(start, end int) {
		for y := start; y < end; y++ {
			for x := range width {
				alpha := 255

				// top-left corner
				dx := float64(radius - x)
				dy := float64(radius - y)
				if dx > 0 && dy > 0 && dx*dx+dy*dy > float64(radius*radius) {
					alpha = 0
				}

				// top-right corner
				dx = float64(x - (width - radius - 1))
				dy = float64(radius - y)
				if dx > 0 && dy > 0 && dx*dx+dy*dy > float64(radius*radius) {
					alpha = 0
				}

				// bottom-left corner
				dx = float64(radius - x)
				dy = float64(y - (height - radius - 1))
				if dx > 0 && dy > 0 && dx*dx+dy*dy > float64(radius*radius) {
					alpha = 0
				}

				// bottom-right corner
				dx = float64(x - (width - radius - 1))
				dy = float64(y - (height - radius - 1))
				if dx > 0 && dy > 0 && dx*dx+dy*dy > float64(radius*radius) {
					alpha = 0
				}

				mask.SetAlpha(x, y, color.Alpha{A: uint8(alpha)})
			}
		}
	})

	return mask
}
