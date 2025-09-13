package image

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
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
	var rTotal, gTotal, bTotal, count uint32

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rTotal += r >> 8
			gTotal += g >> 8
			bTotal += b >> 8
			count++
		}
	}

	if count == 0 {
		return 0, fmt.Errorf("image has no pixels")
	}

	rAvg := rTotal / count
	gAvg := gTotal / count
	bAvg := bTotal / count

	colorInt := int(rAvg)<<16 | int(gAvg)<<8 | int(bAvg)
	return colorInt, nil
}
