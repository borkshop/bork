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
	"sort"

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
	d.Foreground.Pix[i] = f.R
	d.Foreground.Pix[i+1] = f.G
	d.Foreground.Pix[i+2] = f.B
	d.Foreground.Pix[i+3] = f.A
	d.Background.Pix[i] = b.R
	d.Background.Pix[i+1] = b.G
	d.Background.Pix[i+2] = b.B
	d.Background.Pix[i+3] = b.A
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
	t = d.Text.At(x, y)
	f = d.Foreground.At(x, y)
	b = d.Background.At(x, y)
	return t, f, b
}

// RGBAAt is a faster version of At.
func (d *Display) RGBAAt(x, y int) (t string, f, b color.RGBA) {
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
	vp := over.Rect
	pt := vp.Min
	i := over.Text.StringsOffset(pt.X, pt.Y)
	buf, cur = cur.Go(buf, pt)
	for {
		ot, of, ob := over.rgbaati(i)
		if len(ot) == 0 {
			ot = " "
		}
		buf, cur = renderColor(buf, cur, of, ob)
		buf, cur = cur.WriteGlyph(buf, ot)
		i++
		if i >= len(over.Text.Strings) {
			break
		}
		pt.X++
		if pt.X >= vp.Max.X {
			pt.X = vp.Min.X
			pt.Y++
			if pt.Y >= vp.Max.Y {
				break
			}
			buf, cur = cur.linedown(buf, 1)
			if vp.Min.X > 0 {
				buf, cur = cur.right(buf, vp.Min.X)
			}
		}
	}
	buf, cur = cur.Reset(buf)
	return buf, cur
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
	j := under.Text.StringsOffset(pt.X, pt.Y)
	buf, cur = cur.Go(buf, pt)
	for i < len(over.Text.Strings) {
		ot, of, ob := over.rgbaati(i)
		ut, uf, ub := under.rgbaati(j)
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
			if dx := pt.X - cur.Position.X; dx > 0 {
				buf, cur = cur.right(buf, dx)
			}
			buf, cur = renderColor(buf, cur, of, ob)
			buf, cur = cur.WriteGlyph(buf, ot)
			under.setrgbai(j, ot, of, ob)
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

// Renderer supports differential display rendering by tracking invalidated
// cells, rather than requiring a classic front/back buffer pair.
type Renderer struct {
	*Display
	Model
	inval [][2]int
	q     int
}

// NewRenderer creates a new differential renderer around the given display
// buffer.
func NewRenderer(m Model, d *Display) *Renderer {
	return &Renderer{
		Display: d,
		Model:   m,
		inval:   make([][2]int, 0, d.Rect.Dx()*d.Rect.Dy()/2),
	}
}

// Diff scans any overlapping region with the given front buffer, updating
// cells in the renderer's backing buffer. Useful to support classical code
// that wants to own and generate a front buffer.
//
// Returns the number of updated cells.
func (r *Renderer) Diff(over *Display) (n int) {
	vp := over.Rect.Intersect(r.Rect)
	pt := vp.Min
	i := over.Text.StringsOffset(pt.X, pt.Y)
	j := r.Text.StringsOffset(pt.X, pt.Y)
	for i < len(over.Text.Strings) {
		if ot, of, ob := over.rgbaati(i); r.setrgbai(j, ot, of, ob) {
			n++
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
	return n
}

// Render all invalidated cells into the given buffer, wrt the given cursor
// state. Returns a extended buffer and updated cursor state (noop if no
// invalidated cells).
func (r *Renderer) Render(buf []byte, cur Cursor) ([]byte, Cursor) {
	if len(r.inval) == 0 {
		return buf, cur
	}
	maxX := r.Rect.Dx()
	stride := r.Text.Stride
	for i := range r.inval {
		j, k := r.inval[i][0], r.inval[i][1]
		buf, cur = cur.Go(buf, image.Pt(j%stride, j/stride))
		for {
			t, f, b := r.rgbaati(j)
			buf, cur = r.RenderRGBA(buf, cur, f, b)
			buf, cur = cur.WriteGlyph(buf, t)
			j++
			if j > k {
				break
			}
			if cur.Position.X >= maxX {
				buf, cur = cur.linedown(buf, 1)
				if r.Rect.Min.X > 0 {
					buf, cur = cur.right(buf, r.Rect.Min.X)
				}
			} else if cur.Position.X < 0 {
				buf, cur = cur.Go(buf, image.Pt(j%stride, j/stride))
			}
		}
	}
	r.inval = r.inval[:0]
	return cur.Reset(buf)
}

// Set a cell, marking it as invalid; does NOT check for difference.
func (r *Renderer) Set(x, y int, t string, f, b color.Color) {
	stride := r.Text.Stride
	i := y*stride + x
	r.invalidate(i)
	r.Display.Set(x, y, t, f, b)
}

// SetRGBA values and text into a cell, invalidating ONLY IF changed.
func (r *Renderer) SetRGBA(x, y int, t string, f, b color.RGBA) {
	stride := r.Text.Stride
	i := y*stride + x
	r.setrgbai(i, t, f, b)
}

func (r *Renderer) setrgbai(i int, t string, f, b color.RGBA) bool {
	ut, uf, ub := r.rgbaati(i)
	if len(t) == 0 {
		t = " "
	}
	if len(ut) == 0 {
		ut = " "
	}
	if t != ut || f != uf || b != ub {
		r.invalidate(i)
		r.Display.setrgbai(i, t, f, b)
		return true
	}
	return false
}

func (r *Renderer) invalidate(i int) {
	r.q = i
	j := sort.Search(len(r.inval), r.search)

	// beyond
	if j == len(r.inval) {
		r.inval = append(r.inval, [2]int{i, i})
	}

	// already invalidated
	if i <= r.inval[j][1] {
		return
	}

	// expand current end
	if k := r.inval[j][1] + 1; k == i {
		// coalesce
		if j < len(r.inval)-1 && r.inval[j+1][0]-1 == i {
			r.inval[j][1] = r.inval[j+1][1]
			copy(r.inval[j+1:], r.inval[j+2:])
			r.inval = r.inval[:len(r.inval)-1]
			return
		}

		r.inval[j][1] = i
		return
	}

	// expand current start
	if k := r.inval[j][0] - 1; k == i {
		// coalesce
		if j > 0 && r.inval[j-1][1]+1 == i {
			r.inval[j-1][1] = r.inval[j][1]
			copy(r.inval[j:], r.inval[j+1:])
			r.inval = r.inval[:len(r.inval)-1]
			return
		}

		r.inval[j][0] = i
		return
	}

	// expand prior
	if j > 0 {
		if k := r.inval[j-1][1] + 1; k == i {
			r.inval[j-1][1] = k
			return
		}
	}

	// insert
	copy(r.inval[j+1:len(r.inval)+1], r.inval[j:])
	r.inval[j] = [2]int{i, i}
}

func (r Renderer) search(i int) bool { return r.inval[i][0] >= r.q }
