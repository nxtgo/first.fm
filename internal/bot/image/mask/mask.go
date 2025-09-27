package mask

import (
	"image"
	"image/color"
	"math"

	"go.fm/internal/bot/image/parallel"
)

func Rounded(width, height, radius int) *image.Alpha {
	mask := image.NewAlpha(image.Rect(0, 0, width, height))

	parallel.Line(height, func(start, end int) {
		for y := start; y < end; y++ {
			for x := range width {
				alpha := 255.0

				// top-left corner
				dx := float64(radius - x)
				dy := float64(radius - y)
				if dx > 0 && dy > 0 {
					dist := math.Hypot(dx, dy)
					if dist > float64(radius) {
						alpha = 0
					} else if dist > float64(radius)-1 {
						alpha = 255 * (float64(radius) - dist)
					}
				}

				// top-right corner
				dx = float64(x - (width - radius - 1))
				dy = float64(radius - y)
				if dx > 0 && dy > 0 {
					dist := math.Hypot(dx, dy)
					if dist > float64(radius) {
						alpha = 0
					} else if dist > float64(radius)-1 {
						alpha = math.Min(alpha, 255*(float64(radius)-dist))
					}
				}

				// bottom-left corner
				dx = float64(radius - x)
				dy = float64(y - (height - radius - 1))
				if dx > 0 && dy > 0 {
					dist := math.Hypot(dx, dy)
					if dist > float64(radius) {
						alpha = 0
					} else if dist > float64(radius)-1 {
						alpha = math.Min(alpha, 255*(float64(radius)-dist))
					}
				}

				// bottom-right corner
				dx = float64(x - (width - radius - 1))
				dy = float64(y - (height - radius - 1))
				if dx > 0 && dy > 0 {
					dist := math.Hypot(dx, dy)
					if dist > float64(radius) {
						alpha = 0
					} else if dist > float64(radius)-1 {
						alpha = math.Min(alpha, 255*(float64(radius)-dist))
					}
				}

				mask.SetAlpha(x, y, color.Alpha{A: uint8(alpha)})
			}
		}
	})

	return mask
}

func GradientHorizontal(width, height int, reverse bool) *image.Alpha {
	mask := image.NewAlpha(image.Rect(0, 0, width, height))

	parallel.Line(height, func(start, end int) {
		for y := start; y < end; y++ {
			for x := range width {
				var alpha uint8
				if reverse {
					alpha = uint8((x * 255) / width)
				} else {
					alpha = uint8(255 - (x * 255 / width))
				}
				mask.SetAlpha(x, y, color.Alpha{A: alpha})
			}
		}
	})

	return mask
}
