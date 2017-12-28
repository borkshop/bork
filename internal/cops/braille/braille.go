// Package braille composites bitmaps as braille.
package braille

import (
	"image"
	"image/color"

	"github.com/borkshop/bork/internal/bitmap"
	"github.com/borkshop/bork/internal/cops/display"
)

// Margin is the typical margin of skipped bits necessary to make Braille line
// art look straight. Pass as the margin argument to DrawBitmap.
var Margin = image.Point{1, 2}

// BitmapAt returns the braille glyph that coresponds to the 2x4 grid at the
// given point in a bitmap.
func BitmapAt(src *bitmap.Bitmap, sp image.Point) string {
	var r rune
	if src.At(sp.X, sp.Y) {
		r |= 0x1
	}
	if src.At(sp.X, sp.Y+1) {
		r |= 0x2
	}
	if src.At(sp.X, sp.Y+2) {
		r |= 0x4
	}
	if src.At(sp.X, sp.Y+3) {
		r |= 0x40
	}
	if src.At(sp.X+1, sp.Y) {
		r |= 0x8
	}
	if src.At(sp.X+1, sp.Y+1) {
		r |= 0x10
	}
	if src.At(sp.X+1, sp.Y+2) {
		r |= 0x20
	}
	if src.At(sp.X+1, sp.Y+3) {
		r |= 0x80
	}
	if r == 0 {
		return ""
	}
	return string(0x2800 + r)
}

// DrawBitmap draws a braille bitmap onto a display, setting the foreground
// color for any cells with a prsent braille character. Skips over pixels in
// the given margin between cells. Passing braille.Margin drops pixels between
// cells to preserve the appearance of straight lines. Passing image.ZP
// preserves the entire image, but will render discontinuities in the margin
// between cells.
func DrawBitmap(dst *display.Display, r image.Rectangle, src *bitmap.Bitmap, sp image.Point, m image.Point, fg color.Color) {
	r = r.Intersect(dst.Bounds())
	if r.Empty() {
		return
	}

	w, h := r.Dx(), r.Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			pt := image.Pt(x*(2+m.X), y*(4+m.Y)).Add(sp)
			dx := r.Min.X + x
			dy := r.Min.Y + y
			t := BitmapAt(src, pt)
			if t != "" {
				dst.Text.Set(dx, dy, t)
				dst.Foreground.Set(dx, dy, fg)
			}
		}
	}
}

// Bounds takes a rectangle describing cells on a display to the cells of a
// braille bitmap covering the cells of the display.
// Accepts a margin, bits to skip over between cells. Passing braille.Margin
// drops some pixels to achieve the possibility of rendering straighter lines.
// Passing image.ZP covers every bit of the bitmap, though all readable fonts
// will draw a margin between braille characters.
func Bounds(r image.Rectangle, m image.Point) image.Rectangle {
	w, h := r.Dx(), r.Dy()
	return image.Rectangle{
		r.Min,
		r.Min.Add(image.Pt(w*(2+m.X), h*(4+m.Y))).Sub(m),
	}
}
