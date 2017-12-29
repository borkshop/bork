package displaydraw

import (
	"image"
	"image/color"

	"github.com/borkshop/bork/internal/cops/display"
)

// ASCIIBox draws a box onto a display with given foreground and background
// colors using ASCII "|" and "-", with "." and "'' for corners.
func ASCIIBox(d *display.Display, r image.Rectangle, fg, bg color.Color) {
	r.Max = r.Max.Sub(image.Point{1, 1})
	for y := r.Min.Y + 1; y < r.Max.Y; y++ {
		d.Set(r.Min.X, y, "|", fg, bg)
		d.Set(r.Max.X, y, "|", fg, bg)
	}
	for x := r.Min.X + 1; x < r.Max.X; x++ {
		d.Set(x, r.Min.Y, "-", fg, bg)
		d.Set(x, r.Max.Y, "-", fg, bg)
	}
	d.Set(r.Min.X, r.Min.Y, ".", fg, bg)
	d.Set(r.Min.X, r.Max.Y, "'", fg, bg)
	d.Set(r.Max.X, r.Min.Y, ".", fg, bg)
	d.Set(r.Max.X, r.Max.Y, "'", fg, bg)
}

// SpaceBox draws a border on the interior of the given rectangle.
// The border shows a color filled cells, but copying the screen will reveal
// "|" and "-" characters with the same foreground and background.
func SpaceBox(d *display.Display, r image.Rectangle, b image.Point, c color.Color) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := 0; x < b.X; x++ {
			d.Set(r.Min.X+x, y, "|", c, c)
			d.Set(r.Max.X-x-1, y, "|", c, c)
		}
	}
	for x := r.Min.X; x < r.Max.X; x++ {
		for y := 0; y < b.Y; y++ {
			d.Set(x, r.Min.Y+y, "-", c, c)
			d.Set(x, r.Max.Y-y-1, "-", c, c)
		}
	}
}
