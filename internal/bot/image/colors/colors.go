package colors

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"sync/atomic"
	"time"

	"go.fm/internal/bot/image/parallel"
)

func rgbToHsl(r, g, b float64) (h, s, l float64) {
	r /= 255
	g /= 255
	b /= 255

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2

	if max == min {
		h, s = 0, 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h /= 6
	}
	return
}

func hslToRgb(h, s, l float64) (r, g, b int) {
	var rF, gF, bF float64

	if s == 0 {
		rF, gF, bF = l, l, l
	} else {
		var hue2rgb = func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2 {
				return q
			}
			if t < 2.0/3 {
				return p + (q-p)*(2.0/3-t)*6
			}
			return p
		}

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		rF = hue2rgb(p, q, h+1.0/3)
		gF = hue2rgb(p, q, h)
		bF = hue2rgb(p, q, h-1.0/3)
	}

	return int(rF * 255), int(gF * 255), int(bF * 255)
}

func Dominant(url string) (int, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	limitedReader := &io.LimitedReader{R: resp.Body, N: 10 << 20}

	img, _, err := image.Decode(limitedReader)
	if err != nil {
		return 0, err
	}

	bounds := img.Bounds()
	if bounds.Dx() > 4000 || bounds.Dy() > 4000 {
		return 0x00ADD8, nil
	}
	height := bounds.Dy()

	var rTotal, gTotal, bTotal, count uint64

	parallel.Line(height, func(start, end int) {
		var rLocal, gLocal, bLocal, cLocal uint64
		for y := start; y < end; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				rLocal += uint64(r >> 8)
				gLocal += uint64(g >> 8)
				bLocal += uint64(b >> 8)
				cLocal++
			}
		}
		atomic.AddUint64(&rTotal, rLocal)
		atomic.AddUint64(&gTotal, gLocal)
		atomic.AddUint64(&bTotal, bLocal)
		atomic.AddUint64(&count, cLocal)
	})

	if count == 0 {
		return 0, fmt.Errorf("image has no pixels")
	}

	rAvg := float64(rTotal / count)
	gAvg := float64(gTotal / count)
	bAvg := float64(bTotal / count)

	h, s, l := rgbToHsl(rAvg, gAvg, bAvg)
	s = math.Min(1.0, s*1.5)

	rBoost, gBoost, bBoost := hslToRgb(h, s, l)

	colorInt := (rBoost << 16) | (gBoost << 8) | bBoost
	return colorInt, nil
}
