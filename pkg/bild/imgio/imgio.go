/*Package imgio provides basic image file input/output.*/
package imgio

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

// Encoder encodes the provided image and writes it
type Encoder func(io.Writer, image.Image) error

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
