package font

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Font holds a loaded TTF font and can create faces of different sizes.
type Font struct {
	ttf *opentype.Font
}

// LoadFont loads a TTF font from a file path.
func LoadFont(path string) *Font {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read font file: %v", err)
	}
	ttf, err := opentype.Parse(data)
	if err != nil {
		log.Fatalf("failed to parse font: %v", err)
	}
	return &Font{ttf: ttf}
}

// Face returns a font.Face of the specified size (in points) and DPI.
func (f *Font) Face(size float64, dpi float64) font.Face {
	face, err := opentype.NewFace(f.ttf, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatalf("failed to create font face: %v", err)
	}
	return face
}

// DrawText draws text onto an image at a given position with color and font.Face.
func DrawText(canvas draw.Image, x, y int, text string, col color.Color, face font.Face) {
	d := &font.Drawer{
		Dst:  canvas,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)
}
