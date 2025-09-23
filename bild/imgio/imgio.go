/*Package imgio provides basic image file input/output.*/
package imgio

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
)

// Encoder encodes the provided image and writes it
type Encoder func(io.Writer, image.Image) error

// Open loads and decodes an image from a file and returns it.
//
// Usage example:
//
//	// Decodes an image from a file with the given filename
//	// returns an error if something went wrong
//	img, err := Open("exampleName")
func Open(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// Fetch retrieves the raw image bytes from the given URL.
//
// Usage example:
//
//	data, err := Fetch("https://example.com/image.png")
//	if err != nil {
//	    // handle error
//	}
func Fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// FromUrl retrieves an image from the given URL and decodes it.
func FromUrl(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

// DecodeImage loads and decodes an image from a byte slice and returns it.
//
// Usage example:
//
//	img, err := Decode(data)
//	if err != nil {
//	    // handle error
//	}
func Decode(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// JPEGEncoder returns an encoder to JPEG given the argument 'quality'
func JPEGEncoder(quality int) Encoder {
	return func(w io.Writer, img image.Image) error {
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	}
}

// PNGEncoder returns an encoder to PNG
func PNGEncoder() Encoder {
	return func(w io.Writer, img image.Image) error {
		return png.Encode(w, img)
	}
}

// Encode encodes an image into a byte slice using the provided encoder.
//
// Usage example:
//
//	data, err := Encode(img, imgio.JPEGEncoder(90))
//	if err != nil {
//	    // handle error
//	}
func Encode(img image.Image, encoder Encoder) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := encoder(buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
