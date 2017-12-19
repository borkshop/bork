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
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"unicode/utf8"

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
	if f != nil {
		d.Foreground.Set(x, y, f)
	}
	if b != nil {
		d.Background.Set(x, y, b)
	}
}

// SetRGBA is a faster Set.
func (d *Display) SetRGBA(x, y int, t string, f, b color.RGBA) {
	if i := d.Text.StringsOffset(x, y); i >= 0 && i < len(d.Text.Strings) {
		d.setrgbai(i, t, f, b)
	}
}

// MergeRGBA sets the given string and colors if they are
// non-empty and not transparent respectively.
func (d *Display) MergeRGBA(x, y int, t string, f, b color.RGBA) {
	if i := d.Text.StringsOffset(x, y); i >= 0 && i < len(d.Text.Strings) {
		if t != "" {
			d.Text.Strings[i] = t
		}
		if f.A > 0 {
			// TODO blend < 0xff
			d.Foreground.Pix[i] = f.R
			d.Foreground.Pix[i+1] = f.G
			d.Foreground.Pix[i+2] = f.B
			d.Foreground.Pix[i+3] = f.A
		}
		if b.A > 0 {
			// TODO blend < 0xff N.B also over Foreground
			d.Background.Pix[i] = b.R
			d.Background.Pix[i+1] = b.G
			d.Background.Pix[i+2] = b.B
			d.Background.Pix[i+3] = b.A
		}
	}
}

// TODO func (d *Display) Merge(x, y, t, f, b)

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
	clip(dst.Bounds(), &r, src.Bounds(), &sp, nil, nil)
	if r.Empty() {
		return
	}
	draw.Draw(dst.Background, r, src.Background, sp, op)
	draw.Draw(dst.Foreground, r, src.Background, sp, op)
	draw.Draw(dst.Foreground, r, src.Foreground, sp, op)
	textile.Draw(dst.Text, r, src.Text, sp)
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

func unpackColor(c color.Color) (rgba color.RGBA, ok bool) {
	if c != nil {
		r, g, b, a := c.RGBA()
		rgba.R = uint8(r >> 8)
		rgba.G = uint8(g >> 8)
		rgba.B = uint8(b >> 8)
		rgba.A = uint8(a >> 8)
		ok = true
	}
	return rgba, ok
}

// WriteString writes a string into the display at the given position,
// returning how many cells were affected.
//
// NOTE does not support multi-rune glyphs
func (d *Display) WriteString(x, y int, f, b color.Color, mess string, args ...interface{}) int {
	fRGBA, haveF := unpackColor(f)
	bRGBA, haveB := unpackColor(b)
	if len(args) > 0 {
		mess = fmt.Sprintf(mess, args...)
	}
	i := d.Text.StringsOffset(x, y)
	j := i
	for dx := d.Rect.Dx(); len(mess) > 0 && x < dx; x, j = x+1, j+1 {
		_, n := utf8.DecodeRuneInString(mess)
		d.Text.Strings[j] = mess[:n]
		mess = mess[n:]
		k := j * 4
		if haveF {
			d.Foreground.Pix[k] = fRGBA.R
			d.Foreground.Pix[k+1] = fRGBA.G
			d.Foreground.Pix[k+2] = fRGBA.B
			d.Foreground.Pix[k+3] = fRGBA.A
		}
		if haveB {
			d.Background.Pix[k] = bRGBA.R
			d.Background.Pix[k+1] = bRGBA.G
			d.Background.Pix[k+2] = bRGBA.B
			d.Background.Pix[k+3] = bRGBA.A
		}
	}
	return j - i
}

// WriteStringRTL is like WriteString except it goes Right-To-Left (in both the
// string and the diplay).
//
// NOTE does not support multi-rune glyphs
func (d *Display) WriteStringRTL(x, y int, f, b color.Color, mess string, args ...interface{}) int {
	fRGBA, haveF := unpackColor(f)
	bRGBA, haveB := unpackColor(b)
	if len(args) > 0 {
		mess = fmt.Sprintf(mess, args...)
	}
	if x > d.Rect.Max.X {
		x = d.Rect.Max.X - 1
	}
	i := d.Text.StringsOffset(x, y)
	j := i
	for ; len(mess) > 0 && x >= 0; x, j = x-1, j-1 {
		_, n := utf8.DecodeLastRuneInString(mess)
		m := len(mess) - n
		d.Text.Strings[j] = mess[m:]
		mess = mess[:m]
		k := j * 4
		if haveF {
			d.Foreground.Pix[k] = fRGBA.R
			d.Foreground.Pix[k+1] = fRGBA.G
			d.Foreground.Pix[k+2] = fRGBA.B
			d.Foreground.Pix[k+3] = fRGBA.A
		}
		if haveB {
			d.Background.Pix[k] = bRGBA.R
			d.Background.Pix[k+1] = bRGBA.G
			d.Background.Pix[k+2] = bRGBA.B
			d.Background.Pix[k+3] = bRGBA.A
		}
	}
	return i - j
}
