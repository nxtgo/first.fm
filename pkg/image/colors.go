package image

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"sync/atomic"

	"go.fm/pkg/bild/parallel"
)

func DominantColor(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return 0, err
	}

	bounds := img.Bounds()
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

	rAvg := rTotal / count
	gAvg := gTotal / count
	bAvg := bTotal / count

	colorInt := int(rAvg)<<16 | int(gAvg)<<8 | int(bAvg)
	return colorInt, nil
}
