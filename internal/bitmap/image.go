package bitmap

import (
	"image"
	"image/color"
)

// Reader is a view into a bitmap.
type Reader interface {
	At(x, y int) bool
	Bounds() image.Rectangle
}

// Writer is a writable view into a bitmap.
type Writer interface {
	Set(x, y int, b bool)
	Bounds() image.Rectangle
}

// ReaderWriter is a readable/writable view into a bitmap.
type ReaderWriter interface {
	At(x, y int) bool
	Set(x, y int, b bool)
	Bounds() image.Rectangle
}

// FromImage produces a bitmap view of an image, where bits are on or off
// depending on whether they more closely match the on or off color.
func FromImage(i image.Image, on, off color.Color) Reader {
	return &imageReader{
		Image:   i,
		Palette: color.Palette{on, off},
	}
}

// imageReader is a bitmap Reader that provides a bitmap view into an image.
type imageReader struct {
	Image   image.Image
	Palette color.Palette
}

// At returns whether the bit at a point more closely resembles the "on"
// color in the palette.
func (r *imageReader) At(x, y int) bool {
	c := r.Image.At(x, y)
	return color.Model(r.Palette).Convert(c) == r.Palette[0]
}

// Bounds returns the bounds of the underlying image.
func (r *imageReader) Bounds() image.Rectangle {
	return r.Image.Bounds()
}

// ToImage produces a view of a bitmap with colors corresonding to on and off
// bits.
func ToImage(b *Bitmap, on, off color.Color) image.Image {
	return &toImage{
		Bitmap:  b,
		Palette: color.Palette{on, off},
	}
}

// toImage provides an image interpretation of a bitmap.
type toImage struct {
	Bitmap  *Bitmap
	Palette color.Palette
}

// At returns the color at a point.
func (i *toImage) At(x, y int) color.Color {
	if i.Bitmap.At(x, y) {
		return i.Palette[0]
	}
	return i.Palette[1]
}

// Bounds returns the bounds of the bitmap
func (i *toImage) Bounds() image.Rectangle {
	return i.Bitmap.Rect
}

// Set sets the color at a point.
func (i *toImage) Set(x, y int, c color.Color) {
	i.Bitmap.Set(x, y, color.Model(i.Palette).Convert(c) != i.Palette[0])
}

// ColorModel returns the bitmap's palette.
func (i *toImage) ColorModel() color.Model {
	return i.Palette
}
