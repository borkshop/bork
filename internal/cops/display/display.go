// Package display models, composes, and renders virtual terminal displays
// using ANSI escape sequences.
//
// Models displays as three layers: a text layer and foreground and background
// color layers as images in any logical color space.
//
// Also included are colors, palettes, and rendering models for terminal
// displays supporting 0, 3, 4, 8, and 24 bit color.
//
// Finally a cursor that tracks the known state of the terminal cursor is
// included; it is useful for appending ANSI escape sequences that
// incrementally modify the terminal cursor state .
package display

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/borkshop/bork/internal/cops/textile"
)

// New returns a new display with the given bounding rectangle, which need not
// rest at the origin.
func New(r image.Rectangle) *Display {
	return &Display{
		Background: image.NewRGBA(r),
		Foreground: image.NewRGBA(r),
		Text:       textile.New(r),
		Rect:       r,
	}
}

// New2 returns a pair of displays with the same rectangle, suitable for
// creating front and back buffers.
//
//     bounds := term.Bounds()
//     front, back := display.New2(bounds)
func New2(r image.Rectangle) (*Display, *Display) {
	return New(r), New(r)
}

// Display models a terminal display's state as three images.
type Display struct {
	Background *image.RGBA
	Foreground *image.RGBA
	Text       *textile.Textile
	Rect       image.Rectangle
	// TODO underline and intensity
}

// SubDisplay returns a mutable sub-region within the display, sharing the same
// memory.
func (d *Display) SubDisplay(r image.Rectangle) *Display {
	r = r.Intersect(d.Rect)
	return &Display{
		Background: d.Background.SubImage(r).(*image.RGBA),
		Foreground: d.Foreground.SubImage(r).(*image.RGBA),
		Text:       d.Text.SubText(r),
		Rect:       r,
	}
}

// Fill overwrites every cell with the given text and foreground and background
// colors.
func (d *Display) Fill(r image.Rectangle, t string, f, b color.Color) {
	r = r.Intersect(d.Rect)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			d.Set(x, y, t, f, b)
		}
	}
}

// Clear fills the display with transparent cells.
func (d *Display) Clear(r image.Rectangle) {
	d.Fill(r, "", color.Transparent, color.Transparent)
}

// Set overwrites the text and foreground and background colors of the cell at
// the given position.
func (d *Display) Set(x, y int, t string, f, b color.Color) {
	d.Text.Set(x, y, t)
	d.Foreground.Set(x, y, f)
	d.Background.Set(x, y, b)
}

// SetRGBA is a faster Set.
func (d *Display) SetRGBA(x, y int, t string, f, b color.RGBA) {
	if i := d.Text.StringsOffset(x, y); i >= 0 && i < len(d.Text.Strings) {
		d.setrgbai(i, t, f, b)
	}
}

func (d *Display) setrgbai(i int, t string, f, b color.RGBA) {
	d.Text.Strings[i] = t
	j := i * 4
	d.Foreground.Pix[j] = f.R
	d.Background.Pix[j] = b.R
	j++
	d.Foreground.Pix[j] = f.G
	d.Background.Pix[j] = b.G
	j++
	d.Foreground.Pix[j] = f.B
	d.Background.Pix[j] = b.B
	j++
	d.Foreground.Pix[j] = f.A
	d.Background.Pix[j] = b.A
}

// Draw composes one display over another. The bounds dictate the region of the
// destination. The offset dictates the position within the source. Draw will:
//
// Overwrite the text layer for all non-empty text cells inside the rectangle.
// Fill the text with space " " to overdraw all cells.
//
// Draw the foreground of the source over the foreground of the destination
// image. Typically, the foreground is transparent for all cells empty of
// text. Otherwise, this operation can have interesting results.
//
// Draw the background of the source over the *foreground* of the destination
// image. This allows for translucent background colors on the source image
// partially obscuring the text of the destination image.
//
// Draw the background of the source over the background of the destination
// image.
func Draw(dst *Display, r image.Rectangle, src *Display, sp image.Point, op draw.Op) {
	dst.Draw(r, src, sp, op)
}

// Draw composes one display over another. The bounds dictate the region of the
// destination. The offset dictates the position within the source. Draw will:
//
// Overwrite the text layer for all non-empty text cells inside the rectangle.
// Fill the text with space " " to overdraw all cells.
//
// Draw the foreground of the source over the foreground of the destination
// image. Typically, the foreground is transparent for all cells empty of
// text. Otherwise, this operation can have interesting results.
//
// Draw the background of the source over the *foreground* of the destination
// image. This allows for translucent background colors on the source image
// partially obscuring the text of the destination image.
//
// Draw the background of the source over the background of the destination
// image.
func (d *Display) Draw(r image.Rectangle, src *Display, sp image.Point, op draw.Op) {
	clip(d.Bounds(), &r, src.Bounds(), &sp, nil, nil)
	if r.Empty() {
		return
	}
	draw.Draw(d.Background, r, src.Background, sp, op)
	draw.Draw(d.Foreground, r, src.Background, sp, op)
	draw.Draw(d.Foreground, r, src.Foreground, sp, op)
	textile.Draw(d.Text, r, src.Text, sp)
}

// At returns the text and foreground and background colors at the given
// coordinates.
func (d *Display) At(x, y int) (t string, f, b color.Color) {
	if d == nil {
		return "", Colors[7], color.Transparent
	}
	t = d.Text.At(x, y)
	f = d.Foreground.At(x, y)
	b = d.Background.At(x, y)
	return t, f, b
}

// RGBAAt is a faster version of At.
func (d *Display) RGBAAt(x, y int) (t string, f, b color.RGBA) {
	if d == nil {
		return "", Colors[7], color.RGBA{}
	}
	if i := d.Text.StringsOffset(x, y); i >= 0 && i < len(d.Text.Strings) {
		return d.rgbaati(i)
	}
	return t, f, b
}

func (d *Display) rgbaati(i int) (t string, f, b color.RGBA) {
	t = d.Text.Strings[i]
	i *= 4
	f.R = d.Foreground.Pix[i]
	f.G = d.Foreground.Pix[i+1]
	f.B = d.Foreground.Pix[i+2]
	f.A = d.Foreground.Pix[i+3]
	b.R = d.Background.Pix[i]
	b.G = d.Background.Pix[i+1]
	b.B = d.Background.Pix[i+2]
	b.A = d.Background.Pix[i+3]
	return t, f, b
}

// Bounds returns the bounding rectangle of the display.
func (d *Display) Bounds() image.Rectangle {
	return d.Rect
}

// Render appends ANSI escape sequences to a byte slice to overwrite an entire
// terminal window, using the best matching colors in the terminal color model.
func Render(buf []byte, cur Cursor, over *Display, renderColor ColorModel) ([]byte, Cursor) {
	return RenderOver(buf, cur, over, nil, renderColor)
}

// RenderOver appends ANSI escape sequences to a byte slice to update a
// terminal display to look like the front model, skipping cells that are the
// same in the back model, using escape sequences and the nearest matching
// colors in the given color model.
func RenderOver(buf []byte, cur Cursor, over, under *Display, renderColor ColorModel) ([]byte, Cursor) {
	vp := over.Rect
	if under != nil {
		vp = over.Rect.Intersect(under.Rect)
	}
	pt := vp.Min
	i := over.Text.StringsOffset(pt.X, pt.Y)
	j := 0
	if under != nil {
		j = under.Text.StringsOffset(pt.X, pt.Y)
	}
	buf, cur = cur.Go(buf, pt)
	for i < len(over.Text.Strings) {
		var ut string
		var uf, ub color.RGBA
		ot, of, ob := over.rgbaati(i)
		if under != nil {
			ut, uf, ub = under.rgbaati(j)
		}
		if len(ot) == 0 {
			ot = " "
		}
		if len(ut) == 0 {
			ut = " "
		}
		if ot != ut || of != uf || ob != ub {
			if dy := pt.Y - cur.Position.Y; dy > 0 {
				buf, cur = cur.linedown(buf, dy)
			}
			if cur.Position.X < 0 {
				buf = append(buf, "\r"...)
				cur.Position.X = 0
				buf, cur = cur.right(buf, pt.X)
			} else if dx := pt.X - cur.Position.X; dx > 0 {
				buf, cur = cur.right(buf, dx)
			}
			buf, cur = renderColor(buf, cur, of, ob)
			buf, cur = cur.WriteGlyph(buf, ot)
			if under != nil {
				under.setrgbai(j, ot, of, ob)
			}
		}
		pt.X++
		if pt.X >= vp.Max.X {
			pt.X = vp.Min.X
			pt.Y++
		}
		if pt.Y >= vp.Max.Y {
			break
		}
		i++
		j++
	}
	buf, cur = cur.Reset(buf)
	return buf, cur
}
